package collector

import (
	"context"
	"net"
	"time"
)

func (c *Collector) networkCheckLoop(target NetworkTarget) {
	ticker := time.NewTicker(target.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			status := c.checkNetwork(target)

			c.mu.Lock()
			// if len(c.systemStats) > 0 {
			// 	lastStats := &c.systemStats[len(c.systemStats)-1]
			// 	lastStats.Networks[target.Name] = status
			// }
			c.networkStats[target.Name] = status
			c.mu.Unlock()
		}
	}
}

func (c *Collector) checkNetwork(target NetworkTarget) NetworkStatus {
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
