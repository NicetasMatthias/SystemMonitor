package collector

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type SystemStats struct {
	Timestamp   time.Time
	CPUUsage    float64
	MemoryUsed  uint64
	MemoryTotal uint64
	Networks    map[string]NetworkStatus
}

type NetworkStatus struct {
	Reachable bool
	Latency   time.Duration
	LastCheck time.Time
}

type NetworkTarget struct {
	Name     string
	Address  string
	Interval time.Duration
	Timeout  time.Duration
}

type Collecor struct {
	mu         sync.RWMutex
	stats      []SystemStats
	maxHistory int
	targets    []NetworkTarget
	ctx        context.Context
	cancel     context.CancelFunc
}

func New(maxHistory int, targets []NetworkTarget) *Collecor {
	ctx, cancel := context.WithCancel(context.Background())
	return &Collecor{
		stats:      make([]SystemStats, 0, maxHistory),
		maxHistory: maxHistory,
		targets:    targets,
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (c *Collecor) Start(interval time.Duration) {
	go c.collectLoop(interval)
	for _, target := range c.targets {
		go c.networkCheckLoop(target)
	}
}

func (c *Collecor) Stop() {
	c.cancel()
}

func (c *Collecor) GetStats() []SystemStats {
	c.mu.Lock()
	defer c.mu.Unlock()

	result := make([]SystemStats, len(c.stats))
	copy(result, c.stats)
	return result
}

func (c *Collecor) collectLoop(interval time.Duration) {
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

func (c *Collecor) collect() {

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
		Networks:    make(map[string]NetworkStatus),
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.stats = append(c.stats, stats)
	if len(c.stats) > c.maxHistory {
		c.stats = c.stats[1:]
	}
}

func (c *Collecor) networkCheckLoop(target NetworkTarget) {
	ticker := time.NewTicker(target.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			status := c.checkNetwork(target)

			c.mu.Lock()
			if len(c.stats) > 0 {
				lastStats := &c.stats[len(c.stats)-1]
				lastStats.Networks[target.Name] = status
			}
			c.mu.Unlock()
		}
	}
}

func (c *Collecor) checkNetwork(target NetworkTarget) NetworkStatus {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), target.Timeout)
	defer cancel()

	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "tcp", target.Address)

	if err != nil {
		return NetworkStatus{
			Reachable: false,
			LastCheck: time.Now(),
		}
	}
	conn.Close()

	return NetworkStatus{
		Reachable: true,
		Latency:   time.Since(start),
		LastCheck: time.Now(),
	}
}
