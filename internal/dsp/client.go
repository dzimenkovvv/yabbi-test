package dsp

import (
	"context"

	"yabbi_test/internal/partner"
)

type BidRequest struct {
	RequestID  string
	Country    string
	DeviceType string
	BidFloor   float64
}

type Client interface {
	Bid(ctx context.Context, p partner.Partner, req BidRequest) error
}
