package collector

import (
	"context"
	"sync"
	"time"
)

// type ExportStats struct {
// 	System       []SystemStats_OLD
// 	NetworkStats map[string]NetworkStatus
// 	DiskStats    map[string]DiskStatus
// }

// type SystemStats_OLD struct {
// 	Timestamp   time.Time
// 	CPUUsage    float64
// 	MemoryUsed  uint64
// 	MemoryTotal uint64
// }

type Collector struct {
	// mu               sync.RWMutex
	// systemStats      []SystemStats_OLD
	// maxHistory       int
	// networkStats     map[string]NetworkStatus
	// targets          []NetworkTarget
	// diskPaths        []string
	// diskStat         map[string]DiskStatus
	// diskCheckTimeout time.Duration
	cpu     *cpuCollector
	disk    *diskCollector
	memory  *memoryCollector
	network *networkCollector
	system  *systemCollector

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

type CollectorExport struct {
	CPU     CPUExport
	Disk    DiskExport
	Memory  MemoryExport
	Network NetworkExport
	System  SystemExport
}

func New(maxHistory int, targets []NetworkTarget, diskPaths []string, diskCheckInterval time.Duration) *Collector {
	ctx, cancel := context.WithCancel(context.Background())
	return &Collector{
		cpu:     newCPUCollector(),
		disk:    newDiskCollector(),
		memory:  newMemoryCollector(),
		network: newNetworkCollector(targets),
		system:  newSystemCollector(),

		ctx:    ctx,
		cancel: cancel,
	}
}

func (c *Collector) Start(interval time.Duration) {

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.cpu.Run(c.ctx)
	}()

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.disk.Run(c.ctx)
	}()

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.memory.Run(c.ctx)
	}()

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.network.Run(c.ctx)
	}()

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.system.Run(c.ctx)
	}()

}

func (c *Collector) Stop() {
	c.cancel()
	c.wg.Wait()
}

func (c *Collector) Get() CollectorExport {
	return CollectorExport{
		CPU:     c.cpu.Get(),
		Disk:    c.disk.Get(),
		Memory:  c.memory.Get(),
		Network: c.network.Get(),
		System:  c.system.Get(),
	}
}

func (c *Collector) GetCPU() CPUExport {
	return c.cpu.Get()
}

func (c *Collector) GetDisk() DiskExport {
	return c.disk.Get()
}

func (c *Collector) GetMemory() MemoryExport {
	return c.memory.Get()
}

func (c *Collector) GetNetwork() NetworkExport {
	return c.network.Get()
}

func (c *Collector) GetSystem() SystemExport {
	return c.system.Get()
}
