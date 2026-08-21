package collector

import (
	"context"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/load"
)

type cpuCollector struct {
	state CPUExport

	interval       time.Duration
	maxHistorySize int
	mx             sync.RWMutex
}
type CPUExport struct {
	History []CPUSample `json:"history"`
}

type CPUSample struct {
	Timestamp time.Time `json:"timestamp"`
	Data      CPUData   `json:"data"`
}

type CPUData struct {
	Usage CPUUsage    `json:"usage"`
	Cores CoresStat   `json:"cores"`
	Load  LoadAverage `json:"load"`
}

type CPUUsage struct {
	Valid bool    `json:"valid"`
	Data  float64 `json:"data"`
}

type CoreStat struct {
	ID    int     `json:"id"`
	Usage float64 `json:"usage"`
}

type CoresStat struct {
	Valid bool       `json:"valid"`
	Data  []CoreStat `json:"data"`
}

type LoadAverage struct {
	Valid bool         `json:"valid"`
	Data  load.AvgStat `json:"data"`
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

	if len(c.state.History) >= c.maxHistorySize {
		c.state.History = c.state.History[1:]
	}

	c.state.History = append(c.state.History, sample)
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

func (exp *CPUExport) DeepCopy() CPUExport {
	history := make([]CPUSample, len(exp.History))
	for i, sample := range exp.History {
		history[i] = cloneCPUSample(sample)
	}
	return CPUExport{
		History: history,
	}
}

func (c *cpuCollector) Get() CPUExport {
	c.mx.RLock()
	defer c.mx.RUnlock()
	return c.state.DeepCopy()
}
