package collector

import (
	"context"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type ExportStats struct {
	System       []SystemStats
	NetworkStats map[string]NetworkStatus
	DiskStats    map[string]DiskStatus
}

type SystemStats struct {
	Timestamp   time.Time
	CPUUsage    float64
	MemoryUsed  uint64
	MemoryTotal uint64
}

type NetworkStatus struct {
	Reachable bool
	Latency   time.Duration
	LastCheck time.Time
}

type DiskStatus struct {
	Used  float64
	Total float64
}

type NetworkTarget struct {
	Name     string
	Address  string
	Interval time.Duration
	Timeout  time.Duration
}

type Collector struct {
	mu               sync.RWMutex
	systemStats      []SystemStats
	maxHistory       int
	networkStats     map[string]NetworkStatus
	targets          []NetworkTarget
	diskPaths        []string
	diskStat         map[string]DiskStatus
	diskCheckTimeout time.Duration
	ctx              context.Context
	cancel           context.CancelFunc
}

func New(maxHistory int, targets []NetworkTarget, diskPaths []string, diskCheckInterval time.Duration) *Collector {
	ctx, cancel := context.WithCancel(context.Background())
	return &Collector{
		systemStats:      make([]SystemStats, 0, maxHistory),
		maxHistory:       maxHistory,
		networkStats:     make(map[string]NetworkStatus),
		diskStat:         make(map[string]DiskStatus),
		diskCheckTimeout: diskCheckInterval,
		targets:          targets,
		diskPaths:        diskPaths,
		ctx:              ctx,
		cancel:           cancel,
	}
}

func (c *Collector) Start(interval time.Duration) {
	go c.collectLoop(interval)
	go c.diskCheckLoop()
	for _, target := range c.targets {
		go c.networkCheckLoop(target)
	}

}

func (c *Collector) Stop() {
	c.cancel()
}

func (c *Collector) GetStats() ExportStats {
	c.mu.Lock()
	defer c.mu.Unlock()

	result := ExportStats{
		System:       make([]SystemStats, len(c.systemStats)),
		NetworkStats: c.networkStats,
		DiskStats:    c.diskStat,
	}

	copy(result.System, c.systemStats)
	return result
}

func (c *Collector) collectLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	c.collect()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.collect()
		}
	}
}

func (c *Collector) collect() {

	cpuPercent, _ := cpu.Percent(0, false)
	cpuUsage := 0.0
	if len(cpuPercent) > 0 {
		cpuUsage = cpuPercent[0]
	}

	memStat, _ := mem.VirtualMemory()

	stats := SystemStats{
		Timestamp:   time.Now(),
		CPUUsage:    cpuUsage,
		MemoryUsed:  memStat.Used,
		MemoryTotal: memStat.Total,
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.systemStats = append(c.systemStats, stats)
	if len(c.systemStats) > c.maxHistory {
		c.systemStats = c.systemStats[1:]
	}
}
