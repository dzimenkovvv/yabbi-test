package dsp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"yabbi_test/internal/partner"
)

type FakeClient struct {
	Delays map[string]time.Duration

	Errors map[string]error

	mu    sync.Mutex
	calls []string
}

func NewFakeClient() *FakeClient {
	return &FakeClient{
		Delays: make(map[string]time.Duration),
		Errors: make(map[string]error),
	}
}

func (c *FakeClient) Bid(ctx context.Context, p partner.Partner, req BidRequest) error {
	c.recordCall(p.UID)

	delay, hasDelay := c.Delays[p.UID]
	if hasDelay && delay > 0 {
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return fmt.Errorf("bid to %s: %w", p.UID, ctx.Err())
		}
	}

	if err, failed := c.Errors[p.UID]; failed {
		return fmt.Errorf("bid to %s: %w", p.UID, err)
	}

	return nil
}

func (c *FakeClient) recordCall(uid string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls = append(c.calls, uid)
}

func (c *FakeClient) Calls() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]string, len(c.calls))
	copy(out, c.calls)
	return out
}
