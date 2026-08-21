package collector

import (
	"context"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/load"
)

type cpuCollector struct {
	history []CPUSample

	interval       time.Duration
	maxHistorySize int
	mx             sync.RWMutex
}

type CPUSample struct {
	Timestamp time.Time
	Data      CPUData
}

type CPUData struct {
	Usage CPUUsage
	Cores CoresStat
	Load  LoadAverage
}

type CPUUsage struct {
	Valid bool
	Data  float64
}

type CoreStat struct {
	ID    int
	Usage float64
}

type CoresStat struct {
	Valid bool
	Data  []CoreStat
}

type LoadAverage struct {
	Valid bool
	Data  load.AvgStat
}

func collectCPUSample() CPUSample {
	return CPUSample{
		Timestamp: time.Now(),
		Data: CPUData{

			Usage: collectCPUUsage(),
			Cores: collectCoresStat(),
			Load:  collectLoadAverage(),
		},
	}
}

func collectCPUUsage() CPUUsage {
	usage, err := cpu.Percent(0, false)
	if err != nil {

		//=== TODO: log

		return CPUUsage{
			Valid: false,
		}

	}

	return CPUUsage{
		Valid: true,
		Data:  usage[0],
	}

}

func collectCoresStat() CoresStat {

	cores, err := cpu.Percent(0, true)
	if err != nil {
		return CoresStat{
			Valid: false,
		}
	}
	r := CoresStat{
		Valid: true,
	}
	for id, stat := range cores {
		r.Data = append(r.Data,
			CoreStat{
				ID:    id,
				Usage: stat,
			})
	}
	return r
}

func collectLoadAverage() LoadAverage {
	avg, err := load.Avg()
	if err != nil {
		//=== TODO: log
		return LoadAverage{
			Valid: false,
		}
	}

	return LoadAverage{
		Valid: true,
		Data:  *avg,
	}
}

func newCPUCollector() *cpuCollector {
	return &cpuCollector{
		interval:       time.Second * 2,
		maxHistorySize: 50,
	}
}

func (c *cpuCollector) collect() {

	sample := collectCPUSample()

	c.mx.Lock()
	defer c.mx.Unlock()

	if len(c.history) >= c.maxHistorySize {
		c.history = c.history[1:]
	}

	c.history = append(c.history, sample)
}

func (c *cpuCollector) Run(ctx context.Context) {
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

func cloneCPUSample(sample CPUSample) CPUSample {
	sample.Data.Cores.Data = append([]CoreStat(nil),
		sample.Data.Cores.Data...,
	)
	return sample
}

func (c *cpuCollector) Get() []CPUSample {
	c.mx.RLock()
	defer c.mx.RUnlock()
	history := make([]CPUSample, len(c.history))
	for i, sample := range c.history {
		history[i] = cloneCPUSample(sample)
	}
	return history
}
