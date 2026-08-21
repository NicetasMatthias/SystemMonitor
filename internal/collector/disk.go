package collector

import (
	"context"
	"maps"
	"slices"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/disk"
)

type diskCollector struct {
	stat DiskExport

	maxHistorySize int
	interval       time.Duration
	mx             sync.RWMutex
}

type DiskExport struct {
	MountPoints []MountpointStat
	Devices     map[string]*DeviceStat
}

type MountpointStat struct {
	Mountpoint string
	Device     string
	FS         string
	Capacity   MountpointCapacity
}

type MountpointCapacity struct {
	Valid       bool
	UsedBytes   uint64
	TotalBytes  uint64
	UsedPercent float64
}

type DeviceStat struct {
	History    []DeviceSample
	lastIOData DeviceIOData
}

type DeviceIOData struct {
	Valid      bool
	Timestamp  time.Time
	ReadBytes  uint64
	WriteBytes uint64
	ReadCount  uint64
	WriteCount uint64
}

type DeviceSample struct {
	Valid           bool
	Timestamp       time.Time
	ReadThroughput  float64
	WriteThroughput float64
	ReadIOPS        float64
	WriteIOPS       float64
}

type DiskStatus struct {
	Valid       bool
	Used        float64
	Total       float64
	UsedPercent float64
}

func collectMountpointCapacity(mountpoint string) MountpointCapacity {
	usage, err := disk.Usage(mountpoint)
	if err != nil {
		return MountpointCapacity{
			Valid: false,
		}
	}

	return MountpointCapacity{
		Valid:       true,
		UsedBytes:   usage.Used,
		TotalBytes:  usage.Total,
		UsedPercent: usage.UsedPercent,
	}
}

func collectMountpointStats() []MountpointStat {
	r := []MountpointStat(nil)
	partitions, err := disk.Partitions(false)
	if err != nil {
		//=== TODO: log
		return r
	}

	for _, partition := range partitions {

		r = append(r, MountpointStat{
			Mountpoint: partition.Mountpoint,
			Device:     partition.Device,
			FS:         partition.Fstype,
			Capacity:   collectMountpointCapacity(partition.Mountpoint),
		})
	}

	return r
}

func (c *diskCollector) addDeviceSample(device string, data DeviceIOData) {
	stat, ok := c.stat.Devices[device]

	if !ok {
		if data.Valid {
			c.stat.Devices[device] = &DeviceStat{
				lastIOData: data,
			}
		} else {
			//== TODO: log
		}
		return
	}

	sample := DeviceSample{
		Valid:     data.Valid,
		Timestamp: data.Timestamp,
	}

	if data.Valid {
		tDiff := data.Timestamp.Sub(stat.lastIOData.Timestamp).Seconds()
		sample.ReadThroughput = (float64(data.ReadBytes) - float64(stat.lastIOData.ReadBytes)) / tDiff
		sample.WriteThroughput = (float64(data.WriteBytes) - float64(stat.lastIOData.WriteBytes)) / tDiff
		sample.ReadIOPS = (float64(data.ReadCount) - float64(stat.lastIOData.ReadCount)) / tDiff
		sample.WriteIOPS = (float64(data.WriteCount) - float64(stat.lastIOData.WriteCount)) / tDiff
	}

	if len(stat.History) >= int(c.maxHistorySize) {
		stat.History = stat.History[1:]
	}
	stat.History = append(stat.History, sample)

}

func (c *diskCollector) updateDeviceStats() {
	timestamp := time.Now()
	counters, err := disk.IOCounters()
	if err != nil {
		//=== TODO: log

		for device := range maps.Keys(c.stat.Devices) {
			c.addDeviceSample(device,
				DeviceIOData{
					Valid: false,
				})
		}
		return
	}

	for device, data := range counters {
		c.addDeviceSample(device,
			DeviceIOData{
				Valid:      true,
				Timestamp:  timestamp,
				ReadBytes:  data.ReadBytes,
				WriteBytes: data.WriteBytes,
				ReadCount:  data.ReadCount,
				WriteCount: data.WriteCount,
			})
	}

}

func newDiskCollector() *diskCollector {
	return &diskCollector{
		interval:       time.Second * 2,
		maxHistorySize: 50,
	}
}

func (c *diskCollector) collect() {

	mpStats := collectMountpointStats()

	c.mx.Lock()
	defer c.mx.Unlock()
	c.stat.MountPoints = mpStats
	c.updateDeviceStats()
}

func (c *diskCollector) Run(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	c.collect()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.collect()
		}
	}
}

func (exp *DiskExport) DeepCopy() DiskExport {

	devs := make(map[string]*DeviceStat)
	for devName, devStat := range exp.Devices {
		devs[devName] = &DeviceStat{
			History:    slices.Clone(devStat.History),
			lastIOData: devStat.lastIOData,
		}
	}

	return DiskExport{
		MountPoints: slices.Clone(exp.MountPoints),
		Devices:     devs,
	}
}

func (c *diskCollector) Get() DiskExport {
	c.mx.RLock()
	defer c.mx.RUnlock()
	return c.stat.DeepCopy()
}
