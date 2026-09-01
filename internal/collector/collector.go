package collector

import (
	"context"
	"sync"

	"github.com/NicetasMatthias/SystemMonitor/internal/config"
)

type Collector struct {
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

func New(cfg config.Config) (*Collector, error) {
	//==== TODO: переделать корректнее с учетом что все модули будут читать конфиг
	var targets []NetworkTarget //=== FIXME: читать это из конфига или вообще отдаем его в подмодули
	return &Collector{
		cpu:     newCPUCollector(),
		disk:    newDiskCollector(),
		memory:  newMemoryCollector(),
		network: newNetworkCollector(targets),
		system:  newSystemCollector(),
	}, nil
}

func (c *Collector) Start(ctx context.Context) error {

	c.ctx, c.cancel = context.WithCancel(ctx)

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

	return nil //=== FIXME: придумать как тут и что проверить
}

func (c *Collector) Stop(ctx context.Context) error {
	c.cancel()
	c.wg.Wait()
	return nil
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
