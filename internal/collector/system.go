package collector

import (
	"context"
	"log/slog"
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
		slog.Warn("error collecting host info",
			slog.Any("error", err),
		)
		r.BootTime = time.Unix(0, 0)
	}

	if logicalCPUs, err := cpu.Counts(true); err == nil {
		r.LogicalCPUs = logicalCPUs
	} else {
		slog.Warn("error collecting logical cpu counts",
			slog.Any("error", err),
		)
		r.LogicalCPUs = -1
	}

	if physicalCores, err := cpu.Counts(false); err == nil {
		r.PhysicalCores = physicalCores
	} else {
		slog.Warn("error collecting physical cpu counts",
			slog.Any("error", err),
		)
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
		slog.Warn("error collecting sessions count",
			slog.Any("error", err),
		)
		r.SessionCount = 0
	}

	return r
}

func newSystemCollector(cfg SystemConfig) (*systemCollector, error) {

	coll := &systemCollector{
		interval: time.Second * time.Duration(cfg.Interval),
		state: SystemExport{
			Host: collectHostInfo(),
		},
	}
	return coll, nil
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
