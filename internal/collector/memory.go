package collector

import (
	"context"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/mem"
)

type memoryCollector struct {
	history []MemorySample

	maxHistorySize int
	interval       time.Duration
	mx             sync.Mutex
}

type MemorySample struct {
	Valid     bool
	Timestamp time.Time

	Total      uint64
	Used       uint64
	Available  uint64
	Percentage float64
}

func newMemoryCollector() *memoryCollector {
	return &memoryCollector{
		interval:       time.Second * 2,
		maxHistorySize: 50,
	}
}

func collectMemorySample() MemorySample {

	if stat, err := mem.VirtualMemory(); err != nil {
		//=== TODO: log
		return MemorySample{
			Timestamp: time.Now(),
			Valid:     false,
		}
	} else {
		return MemorySample{
			Timestamp:  time.Now(),
			Valid:      true,
			Total:      stat.Total,
			Used:       stat.Used,
			Available:  stat.Available,
			Percentage: stat.UsedPercent,
		}
	}

}

func (c *memoryCollector) collect() {
	//=== FIXME: WIP
	sample := collectMemorySample()

	c.mx.Lock()
	defer c.mx.Unlock()

	if len(c.history) >= c.maxHistorySize {
		c.history = c.history[1:]
	}
	c.history = append(c.history, sample)
}

func (c *memoryCollector) Run(ctx context.Context) {
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
