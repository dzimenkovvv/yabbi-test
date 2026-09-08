package auction

import (
	"context"
	"fmt"
	"sync"
	"time"

	"yabbi_test/internal/dsp"
	"yabbi_test/internal/partner"
)

// BidOutcome — результат отправки запроса одному партнёру.
type BidOutcome struct {
	PartnerUID string
	Success    bool
	Err        error
}

func Dispatch(ctx context.Context, client dsp.Client, partners []partner.Partner, req Request, timeout time.Duration) []BidOutcome {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	outcomes := make(chan BidOutcome, len(partners))
	var wg sync.WaitGroup

	for _, p := range partners {
		wg.Add(1)
		go func(p partner.Partner) {
			defer wg.Done()
			outcomes <- bidOne(ctx, client, p, req)
		}(p)
	}

	go func() {
		wg.Wait()
		close(outcomes)
	}()

	results := make([]BidOutcome, 0, len(partners))
	for o := range outcomes {
		results = append(results, o)
	}
	return results
}

func bidOne(ctx context.Context, client dsp.Client, p partner.Partner, req Request) (outcome BidOutcome) {
	outcome = BidOutcome{PartnerUID: p.UID}

	defer func() {
		if r := recover(); r != nil {
			outcome.Success = false
			outcome.Err = fmt.Errorf("dsp client panicked: %v", r)
		}
	}()

	bidReq := dsp.BidRequest{
		RequestID:  req.RequestID,
		Country:    req.Country,
		DeviceType: req.DeviceType,
		BidFloor:   req.BidFloor,
	}

	if err := client.Bid(ctx, p, bidReq); err != nil {
		outcome.Err = err
		return outcome
	}

	outcome.Success = true
	return outcome
}
