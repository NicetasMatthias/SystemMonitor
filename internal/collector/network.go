package collector

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"sync"
	"time"
)

type networkCheck interface {
	Check(ctx context.Context, target NetworkTarget) NetworkStatus
}

type endpoint struct {
	target NetworkTarget
	check  networkCheck
}

type networkCollector struct {
	state NetworkExport

	endpoints []endpoint

	mx sync.RWMutex
	wg sync.WaitGroup
}

type NetworkExport struct {
	Stats map[string]NetworkStatus `json:"stats"`
}

type NetworkStatus struct {
	Reachable bool          `json:"reachable"`
	Latency   time.Duration `json:"latency"`
	LastCheck time.Time     `json:"last_check"`
}

func newNetworkCollector(cfg NetworkConfig) (*networkCollector, error) {

	var errs []error
	var endpoints []endpoint

	for _, target := range cfg.Targets {
		check, err := newNetworkCheck(target)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to build endpoint %q: %v", target.Name, err))
		} else {
			endpoints = append(endpoints, endpoint{
				target: target,
				check:  check,
			})
		}
	}

	coll := &networkCollector{
		endpoints: endpoints,
		state: NetworkExport{
			Stats: make(map[string]NetworkStatus),
		},
	}
	return coll, errors.Join(errs...)
}

func newNetworkCheck(target NetworkTarget) (networkCheck, error) {
	switch target.Type {
	case "dns":
		return newDNSCheck(), nil
	case "http", "https":
		return newHTTPCheck(), nil
	default:
		return nil, fmt.Errorf("unsupported network check type %q", target.Type)
	}
}

func (c *networkCollector) collect(e endpoint, ctx context.Context) {
	ctx, cancel := context.WithTimeout(
		ctx,
		time.Duration(e.target.Timeout)*time.Second,
	)
	defer cancel()

	status := e.check.Check(ctx, e.target)

	c.mx.Lock()
	defer c.mx.Unlock()
	c.state.Stats[e.target.Name] = status
}

func (c *networkCollector) Run(ctx context.Context) {
	for _, endpoint := range c.endpoints {
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			c.checkLoop(endpoint, ctx)
		}()
	}
	c.wg.Wait()
}

func (c *networkCollector) checkLoop(e endpoint, ctx context.Context) {
	ticker := time.NewTicker(time.Duration(e.target.Interval))
	defer ticker.Stop()

	c.collect(e, ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.collect(e, ctx)
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
