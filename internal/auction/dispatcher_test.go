package auction

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"yabbi_test/internal/dsp"
	"yabbi_test/internal/partner"
)

func TestDispatch_AllSucceed(t *testing.T) {
	client := dsp.NewFakeClient()
	partners := []partner.Partner{
		{UID: "p1"}, {UID: "p2"}, {UID: "p3"},
	}
	req := Request{RequestID: "r1", Country: "RU", DeviceType: DeviceMobile, BidFloor: 1}

	outcomes := Dispatch(context.Background(), client, partners, req, 200*time.Millisecond)

	require.Len(t, outcomes, 3)
	for _, o := range outcomes {
		require.True(t, o.Success, "partner %s expected success, got err=%v", o.PartnerUID, o.Err)
		require.NoError(t, o.Err)
	}
}

func TestDispatch_MixedOutcomes(t *testing.T) {
	client := dsp.NewFakeClient()
	client.Delays = map[string]time.Duration{
		"slow": 500 * time.Millisecond, // дольше общего таймаута
		"fast": 5 * time.Millisecond,
	}
	client.Errors = map[string]error{
		"failing": errors.New("boom"),
	}
	partners := []partner.Partner{
		{UID: "fast"}, {UID: "slow"}, {UID: "failing"},
	}
	req := Request{RequestID: "r2", Country: "RU", DeviceType: DeviceMobile, BidFloor: 1}

	start := time.Now()
	outcomes := Dispatch(context.Background(), client, partners, req, 100*time.Millisecond)
	elapsed := time.Since(start)

	require.Len(t, outcomes, 3)
	require.Less(t, elapsed, 200*time.Millisecond,
		"Dispatch should respect the overall timeout, not wait for the slow partner")

	byUID := make(map[string]BidOutcome, len(outcomes))
	for _, o := range outcomes {
		byUID[o.PartnerUID] = o
	}

	require.True(t, byUID["fast"].Success)

	require.False(t, byUID["slow"].Success)
	require.ErrorIs(t, byUID["slow"].Err, context.DeadlineExceeded)

	require.False(t, byUID["failing"].Success)
	require.Error(t, byUID["failing"].Err)
}

func TestDispatch_EmptyPartnerList(t *testing.T) {
	client := dsp.NewFakeClient()
	req := Request{RequestID: "r3", Country: "RU", DeviceType: DeviceMobile, BidFloor: 1}

	outcomes := Dispatch(context.Background(), client, nil, req, 100*time.Millisecond)

	require.NotNil(t, outcomes)
	require.Empty(t, outcomes)
}

type panickingClient struct{}

func (panickingClient) Bid(_ context.Context, _ partner.Partner, _ dsp.BidRequest) error {
	panic("dsp client blew up")
}

func TestDispatch_PartnerPanicDoesNotCrashDispatch(t *testing.T) {
	partners := []partner.Partner{{UID: "p1"}, {UID: "p2"}}
	req := Request{RequestID: "r4", Country: "RU", DeviceType: DeviceMobile, BidFloor: 1}

	require.NotPanics(t, func() {
		outcomes := Dispatch(context.Background(), panickingClient{}, partners, req, 100*time.Millisecond)

		require.Len(t, outcomes, 2)
		for _, o := range outcomes {
			require.False(t, o.Success)
			require.ErrorContains(t, o.Err, "panicked")
		}
	})
}

func TestDispatch_NoGoroutineLeak(t *testing.T) {
	client := dsp.NewFakeClient()
	client.Delays = map[string]time.Duration{"slow": 300 * time.Millisecond}
	partners := []partner.Partner{{UID: "slow"}}
	req := Request{RequestID: "r5", Country: "RU", DeviceType: DeviceMobile, BidFloor: 1}

	for i := 0; i < 5; i++ {
		outcomes := Dispatch(context.Background(), client, partners, req, 20*time.Millisecond)
		require.Len(t, outcomes, 1)
		require.False(t, outcomes[0].Success)
	}
}
