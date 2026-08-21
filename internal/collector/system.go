package collector

import (
	"context"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
)

type systemCollector struct {
	state SystemExport

	interval time.Duration
	mx       sync.RWMutex
}

type SystemExport struct {
	Host     HostInfo     `json:"host"`
	Activity ActivityInfo `json:"activity"`
}

type HostInfo struct {
	Hostname        string `json:"hostname"`
	OS              string `json:"os"`
	Platform        string `json:"platform"`
	PlatformFamily  string `json:"platform_family"`
	PlatformVersion string `json:"platform_version"`
	KernelVersion   string `json:"kernel_version"`
	Architecture    string `json:"architecture"`

	BootTime time.Time `json:"boot_time"`

	LogicalCPUs   int `json:"logical_cp_us"`
	PhysicalCores int `json:"physical_cores"`
}

type ActivityInfo struct {
	ProcessCount uint64 `json:"process_count"`
	SessionCount int    `json:"session_count"`
}

func collectHostInfo() HostInfo {
	r := HostInfo{}

	if hostInfo, err := host.Info(); err == nil {
		r.Hostname = hostInfo.Hostname
		r.OS = hostInfo.OS
		r.Platform = hostInfo.Platform
		r.PlatformFamily = hostInfo.PlatformFamily
		r.PlatformVersion = hostInfo.PlatformVersion
		r.KernelVersion = hostInfo.KernelVersion
		r.Architecture = hostInfo.KernelArch
		r.BootTime = time.Unix(int64(hostInfo.BootTime), 0)
	} else {
		//=== TODO: log
		r.BootTime = time.Unix(0, 0)
	}

	if logicalCPUs, err := cpu.Counts(true); err == nil {
		r.LogicalCPUs = logicalCPUs
	} else {
		//=== TODO: log
		r.LogicalCPUs = -1
	}

	if physicalCores, err := cpu.Counts(false); err == nil {
		r.PhysicalCores = physicalCores
	} else {
		//=== TODO: log
		r.PhysicalCores = -1
	}

	return r
}

func collectActivityInfo() ActivityInfo {
	r := ActivityInfo{}
	if hostInfo, err := host.Info(); err == nil {
		r.ProcessCount = hostInfo.Procs
	} else {
		r.ProcessCount = 0
	}

	if users, err := host.Users(); err == nil {
		r.SessionCount = len(users)
	} else {
		//=== TODO: log
		r.SessionCount = 0
	}

	return r
}

func newSystemCollector() *systemCollector {
	return &systemCollector{
		interval: time.Second * 10,
		state: SystemExport{
			Host: collectHostInfo(),
		},
	}
}

func (c *systemCollector) collect() {
	actInfo := collectActivityInfo()

	c.mx.Lock()
	defer c.mx.Unlock()
	c.state.Activity = actInfo
}

func (c *systemCollector) Run(ctx context.Context) {
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

func (exp *SystemExport) DeepCopy() SystemExport {

	return *exp
}

func (c *systemCollector) Get() SystemExport {
	c.mx.RLock()
	defer c.mx.RUnlock()
	return c.state.DeepCopy()
}
