package collector

import (
	"context"
	"log/slog"
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
	Stats map[string]NetworkStatus
}

type NetworkTarget struct {
	Name     string
	Address  string
	Interval time.Duration
	Timeout  time.Duration
}

type NetworkStatus struct {
	Reachable bool
	Latency   time.Duration
	LastCheck time.Time
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

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.collect(target)
		}
	}
}

func (c *Collector) networkCheckLoop(target NetworkTarget) {
	// ticker := time.NewTicker(target.Interval)
	// defer ticker.Stop()

	// for {
	// 	select {
	// 	case <-c.ctx.Done():
	// 		return
	// 	case <-ticker.C:
	// 		status := c.checkNetwork(target)

	// 		c.mu.Lock()
	// 		// if len(c.systemStats) > 0 {
	// 		// 	lastStats := &c.systemStats[len(c.systemStats)-1]
	// 		// 	lastStats.Networks[target.Name] = status
	// 		// }
	// 		c.networkStats[target.Name] = status
	// 		c.mu.Unlock()
	// 	}
	// }
}

func (c *Collector) checkNetwork(target NetworkTarget) NetworkStatus {
	// start := time.Now()
	// ctx, cancel := context.WithTimeout(context.Background(), target.Timeout)
	// defer cancel()

	// var dialer net.Dialer
	// conn, err := dialer.DialContext(ctx, "tcp", target.Address)

	// if err != nil {
	// 	return NetworkStatus{
	// 		Reachable: false,
	// 		LastCheck: time.Now(),
	// 	}
	// }
	// if err := conn.Close(); err != nil {
	// 	slog.Warn("Failed to close check connection",
	// 		slog.Any("name", target.Name),
	// 		slog.Any("host", target.Address),
	// 		slog.Any("error", err))
	// }

	// return NetworkStatus{
	// 	Reachable: true,
	// 	Latency:   time.Since(start),
	// 	LastCheck: time.Now(),
	// }
	return NetworkStatus{}
}

func (c *networkCollector) collect(target NetworkTarget) {
	status := checkNetwork(target)
	c.mx.Lock()
	defer c.mx.Unlock()
	c.state.Stats[target.Name] = status
}

func checkNetwork(target NetworkTarget) NetworkStatus {
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
