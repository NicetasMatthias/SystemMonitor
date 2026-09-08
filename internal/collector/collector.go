package collector

import (
	"context"
	"errors"
	"sync"
)

type Collector struct {
	cpu     *cpuCollector
	disk    *diskCollector
	memory  *memoryCollector
	network *networkCollector
	system  *systemCollector

	startOnce sync.Once
	startErr  error
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

type CollectorExport struct {
	CPU     CPUExport
	Disk    DiskExport
	Memory  MemoryExport
	Network NetworkExport
	System  SystemExport
}

func New(cfg Config) (*Collector, error) {

	cpuColl, cpuErr := newCPUCollector(cfg.Cpu)
	diskColl, diskErr := newDiskCollector(cfg.Disk)
	memColl, memErr := newMemoryCollector(cfg.Memory)
	netColl, netErr := newNetworkCollector(cfg.Network)
	sysColl, sysErr := newSystemCollector(cfg.System)

	return &Collector{
			cpu:     cpuColl,
			disk:    diskColl,
			memory:  memColl,
			network: netColl,
			system:  sysColl,
		},
		errors.Join(
			cpuErr,
			diskErr,
			memErr,
			netErr,
			sysErr,
		)
}

func (c *Collector) Start(ctx context.Context) error {

	c.startOnce.Do(func() {
		ctx, c.cancel = context.WithCancel(ctx)

		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			c.cpu.Run(ctx)
		}()

		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			c.disk.Run(ctx)
		}()

		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			c.memory.Run(ctx)
		}()

		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			c.network.Run(ctx)
		}()

		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			c.system.Run(ctx)
		}()

		//=== TODO: check Run`s via chan and set c.startErr
	})

	return c.startErr
}

func (c *Collector) Stop(ctx context.Context) error {
	if c.cancel == nil {
		return nil
	}
	c.cancel()

	done := make(chan struct{})

	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
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
