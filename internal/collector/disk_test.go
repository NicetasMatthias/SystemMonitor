package collector

import (
	"testing"
	"time"
)

func TestDiskExport_DeepCopy(t *testing.T) {
	original := DiskExport{
		MountPoints: []MountpointStat{
			{
				Mountpoint: "/",
				Device:     "/dev/sda1",
				FS:         "ext4",
				Capacity: MountpointCapacity{
					Valid:      true,
					UsedBytes:  100,
					TotalBytes: 1000,
				},
			},
		},
		Devices: map[string]*DeviceStat{
			"sda": {
				History: []DeviceSample{
					{
						Valid:           true,
						Timestamp:       time.Unix(100, 0),
						ReadThroughput:  100,
						WriteThroughput: 200,
					},
				},
			},
		},
	}

	copyExport := original.DeepCopy()

	copyExport.MountPoints[0].Capacity.UsedBytes = 999
	copyExport.Devices["sda"].History[0].ReadThroughput = 888

	if got := original.MountPoints[0].Capacity.UsedBytes; got != 100 {
		t.Errorf("original UsedBytes = %d, want 100", got)
	}

	if got := original.Devices["sda"].History[0].ReadThroughput; got != 100 {
		t.Errorf("original ReadThroughput = %v, want 100", got)
	}
}

func TestDiskCollector_AddDeviceSample_FirstSample(t *testing.T) {
	c := &diskCollector{
		maxHistorySize: 3,
		stat: DiskExport{
			Devices: make(map[string]*DeviceStat),
		},
	}

	sample := DeviceIOData{
		Valid:      true,
		Timestamp:  time.Unix(100, 0),
		ReadBytes:  1000,
		WriteBytes: 2000,
		ReadCount:  10,
		WriteCount: 20,
	}

	c.addDeviceSample("sda", sample)

	device, ok := c.stat.Devices["sda"]
	if !ok {
		t.Fatal("device was not added")
	}

	if len(device.History) != 0 {
		t.Errorf("history length = %d, want 0", len(device.History))
	}
}
