package collector

import (
	"context"
	"time"

	"github.com/miekg/dns"
)

type dnsCheck struct {
	client *dns.Client
}

func newDNSCheck() *dnsCheck {
	return &dnsCheck{
		client: &dns.Client{},
	}
}

func (c *dnsCheck) Check(ctx context.Context, target NetworkTarget) NetworkStatus {

	msg := new(dns.Msg)
	msg.SetQuestion(".", dns.TypeNS)

	client := *c.client
	client.Net = target.Transport

	status := NetworkStatus{
		LastCheck: time.Now(),
	}

	response, rtt, err := client.ExchangeContext(
		ctx,
		msg,
		target.Address,
	)

	if err != nil || response == nil {
		status.Reachable = false
		return status
	}

	status.Reachable = true
	status.Latency = rtt
	return status
}
