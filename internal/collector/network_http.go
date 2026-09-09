package collector

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

type httpCheck struct {
	client *http.Client
}

func newHTTPCheck() *httpCheck {
	return &httpCheck{
		client: &http.Client{},
	}
}

func (c *httpCheck) Check(ctx context.Context, target NetworkTarget) NetworkStatus {
	start := time.Now()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		target.Type+"://"+target.Address+"/",
		nil,
	)
	status := NetworkStatus{
		LastCheck: start,
	}
	if err != nil {
		status.Reachable = false
		return status
	}

	resp, err := c.client.Do(req)
	if err != nil {
		status.Reachable = false
		return status
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("failed to close response body", slog.Any("error", err))
		}
	}()

	status.Reachable = true
	status.Latency = time.Since(start)

	return status
}
