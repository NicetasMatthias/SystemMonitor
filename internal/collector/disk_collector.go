package collector

import (
	"time"

	"github.com/shirou/gopsutil/v3/disk"
)

type DiskStatus struct {
	Used  float64
	Total float64
}

func (c *Collector) diskCheckLoop() {
	ticker := time.NewTicker(c.diskCheckTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			tmpMap := make(map[string]DiskStatus)
			for _, path := range c.diskPaths {
				tmpMap[path] = checkDisk(path)
			}
			c.mu.Lock()
			for _, path := range c.diskPaths {
				c.diskStat[path] = tmpMap[path]
			}
			c.mu.Unlock()
		}
	}
}

func checkDisk(diskPath string) DiskStatus {
	diskStat, _ := disk.Usage(diskPath)
	const gb = 1024 * 1024 * 1024
	return DiskStatus{
		Used:  float64(diskStat.Used) / float64(gb),
		Total: float64(diskStat.Total) / float64(gb),
	}
}
