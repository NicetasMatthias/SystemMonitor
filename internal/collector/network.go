package collector

import (
	"context"
	"log/slog"
	"maps"
	"net"
	"sync"
	"time"
)

type networkCollector struct {
	state NetworkExport

	targets []NetworkTarget

	mx sync.RWMutex
	wg sync.WaitGroup
}

type NetworkExport struct {
	Stats map[string]NetworkStatus `json:"stats"`
}

type NetworkTarget struct {
	Name     string
	Address  string
	Interval time.Duration
	Timeout  time.Duration
}

type NetworkStatus struct {
	Reachable bool          `json:"reachable"`
	Latency   time.Duration `json:"latency"`
	LastCheck time.Time     `json:"last_check"`
}

func collectNetworkStatus(target NetworkTarget) NetworkStatus {
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
	if err := conn.Close(); err != nil {
		slog.Warn("Failed to close check connection",
			slog.Any("name", target.Name),
			slog.Any("host", target.Address),
			slog.Any("error", err))
	}

	return NetworkStatus{
		Reachable: true,
		Latency:   time.Since(start),
		LastCheck: time.Now(),
	}
}

func newNetworkCollector(netTargets []NetworkTarget) *networkCollector {
	return &networkCollector{
		targets: netTargets,
		state: NetworkExport{
			Stats: make(map[string]NetworkStatus),
		},
	}
}

func (c *networkCollector) collect(target NetworkTarget) {
	status := collectNetworkStatus(target)
	c.mx.Lock()
	defer c.mx.Unlock()
	c.state.Stats[target.Name] = status
}

func (c *networkCollector) Run(ctx context.Context) {
	for _, target := range c.targets {
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			c.checkLoop(target, ctx)
		}()
	}
	c.wg.Wait()
}

func (c *networkCollector) checkLoop(target NetworkTarget, ctx context.Context) {
	ticker := time.NewTicker(target.Interval)
	defer ticker.Stop()

	c.collect(target)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.collect(target)
		}
	}
}

func (exp *NetworkExport) DeepCopy() NetworkExport {

	stats := make(map[string]NetworkStatus, len(exp.Stats))
	maps.Copy(stats, exp.Stats)
	return NetworkExport{
		Stats: stats,
	}
}

func (c *networkCollector) Get() NetworkExport {
	c.mx.RLock()
	defer c.mx.RUnlock()
	return c.state.DeepCopy()
}
