package auction

import (
	"errors"
	"fmt"
)

var (
	ErrMissingRequestID  = errors.New("request_id is required")
	ErrMissingCountry    = errors.New("country is required")
	ErrInvalidCountry    = errors.New("country must be a 2-letter ISO 3166-1 alpha-2 code")
	ErrMissingDeviceType = errors.New("device_type is required")
	ErrInvalidDeviceType = errors.New("device must be one of: mobile, desktop, tv")
	ErrNegativeBidFloor  = errors.New("bid_floor must be >= 0")
)

func Validate(req Request) error {
	if req.RequestID == "" {
		return fmt.Errorf("validate request: %w", ErrMissingRequestID)
	}

	if req.Country == "" {
		return fmt.Errorf("validate request: %w", ErrMissingCountry)
	}
	if !isAlpha2CountryCode(req.Country) {
		return fmt.Errorf("validate request: %w: got %q", ErrInvalidCountry, req.Country)
	}

	if req.DeviceType == "" {
		return fmt.Errorf("validate request: %w", ErrMissingDeviceType)
	}
	if _, ok := ValidDeviceTypes[req.DeviceType]; !ok {
		return fmt.Errorf("validate request: %w: got %q", ErrInvalidDeviceType, req.DeviceType)
	}

	if req.BidFloor < 0 {
		return fmt.Errorf("validate request: %w: got %v", ErrNegativeBidFloor, req.BidFloor)
	}

	return nil
}

func isAlpha2CountryCode(code string) bool {
	if len(code) != 2 {
		return false
	}

	for _, r := range code {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}
