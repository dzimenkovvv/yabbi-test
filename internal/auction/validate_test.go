package auction

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	validReq := Request{
		RequestID:  "3f0a1c9e-2b1d-4a8f-9c11-7e6b2d0a55f1",
		Country:    "RU",
		DeviceType: DeviceMobile,
		BidFloor:   1.5,
		Categories: []string{"news", "sport"},
	}

	tests := []struct {
		name    string
		mutate  func(r Request) Request
		wantErr error
	}{
		{
			name:    "valid request passed",
			mutate:  func(r Request) Request { return r },
			wantErr: nil,
		},
		{
			name:    "missing request_id",
			mutate:  func(r Request) Request { r.RequestID = ""; return r },
			wantErr: ErrMissingRequestID,
		},
		{
			name:    "missing country",
			mutate:  func(r Request) Request { r.Country = ""; return r },
			wantErr: ErrMissingCountry,
		},
		{
			name:    "lowercase country code is invalid",
			mutate:  func(r Request) Request { r.Country = "ru"; return r },
			wantErr: ErrInvalidCountry,
		},
		{
			name:    "3-letter country code is invalid",
			mutate:  func(r Request) Request { r.Country = "RUS"; return r },
			wantErr: ErrInvalidCountry,
		},
		{
			name:    "missing device_type",
			mutate:  func(r Request) Request { r.DeviceType = ""; return r },
			wantErr: ErrMissingDeviceType,
		},
		{
			name:    "unsupported device_type",
			mutate:  func(r Request) Request { r.DeviceType = "smarttv"; return r },
			wantErr: ErrInvalidDeviceType,
		},
		{
			name:    "negative bid_floor",
			mutate:  func(r Request) Request { r.BidFloor = -0.1; return r },
			wantErr: ErrNegativeBidFloor,
		},
		{
			name:    "zero bid_floor is allowed",
			mutate:  func(r Request) Request { r.BidFloor = 0; return r },
			wantErr: nil,
		},
		{
			name:    "empty categories is allowed",
			mutate:  func(r Request) Request { r.Categories = nil; return r },
			wantErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := Validate(test.mutate(validReq))

			if test.wantErr == nil {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			require.True(t, errors.Is(err, test.wantErr), "expected error to wrap %v, got %v", test.wantErr, err)
		})
	}
}
