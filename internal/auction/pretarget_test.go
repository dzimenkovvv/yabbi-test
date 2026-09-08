package auction

import (
	"testing"
	"yabbi_test/internal/partner"

	"github.com/stretchr/testify/require"
)

func TestFilter(t *testing.T) {
	req := Request{
		RequestID:  "req-1",
		Country:    "RU",
		DeviceType: DeviceMobile,
		BidFloor:   1.0,
		Categories: []string{"news"},
	}

	tests := []struct {
		name       string
		partner    partner.Partner
		wantPassed bool
		wantReason string
	}{
		{
			name: "passes all conditions",
			partner: partner.Partner{
				UID:               "p-ok",
				IsEnabled:         true,
				Countries:         []string{"RU"},
				DeviceTypes:       []string{DeviceMobile},
				MinBidFloor:       0.5,
				BlockedCategories: []string{"gambling"},
			},
			wantPassed: true,
		},
		{
			name: "empty countries/device_types means any",
			partner: partner.Partner{
				UID:         "p-any",
				IsEnabled:   true,
				Countries:   nil,
				DeviceTypes: nil,
				MinBidFloor: 0,
			},
			wantPassed: true,
		},
		{
			name: "disabled partner is rejected regardless of other fields",
			partner: partner.Partner{
				UID:         "p-disabled",
				IsEnabled:   false,
				Countries:   nil,
				DeviceTypes: nil,
				MinBidFloor: 0,
			},
			wantPassed: false,
			wantReason: "partner is disabled",
		},
		{
			name: "country not served",
			partner: partner.Partner{
				UID:         "p-country",
				IsEnabled:   true,
				Countries:   []string{"US", "DE"},
				DeviceTypes: nil,
				MinBidFloor: 0,
			},
			wantPassed: false,
			wantReason: "country not served",
		},
		{
			name: "device type not served",
			partner: partner.Partner{
				UID:         "p-device",
				IsEnabled:   true,
				Countries:   nil,
				DeviceTypes: []string{DeviceDesktop, DeviceTV},
				MinBidFloor: 0,
			},
			wantPassed: false,
			wantReason: "device type not served",
		},
		{
			name: "bid floor too low",
			partner: partner.Partner{
				UID:         "p-floor",
				IsEnabled:   true,
				Countries:   nil,
				DeviceTypes: nil,
				MinBidFloor: 5.0,
			},
			wantPassed: false,
			wantReason: "bid floor too low",
		},
		{
			name: "bid floor exactly at minimum passes",
			partner: partner.Partner{
				UID:         "p-floor-eq",
				IsEnabled:   true,
				Countries:   nil,
				DeviceTypes: nil,
				MinBidFloor: 1.0,
			},
			wantPassed: true,
		},
		{
			name: "blocked category rejects",
			partner: partner.Partner{
				UID:               "p-blocked",
				IsEnabled:         true,
				Countries:         nil,
				DeviceTypes:       nil,
				MinBidFloor:       0,
				BlockedCategories: []string{"news", "gambling"},
			},
			wantPassed: false,
			wantReason: "blocked category: news",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := Filter(req, []partner.Partner{test.partner})
			require.Len(t, result, 1)

			got := result[0]
			require.Equal(t, test.wantPassed, got.Passed)
			if !test.wantPassed {
				require.Equal(t, test.wantReason, got.Reason)
			}
		})
	}
}

func TestFilter_Matched(t *testing.T) {
	req := Request{
		RequestID:  "req-1",
		Country:    "RU",
		DeviceType: DeviceMobile,
		BidFloor:   1.0,
	}

	partner := []partner.Partner{
		{UID: "p-pass-1", IsEnabled: true},
		{UID: "p-fail", IsEnabled: false},
		{UID: "p-pass-2", IsEnabled: true},
	}

	results := Filter(req, partner)
	matched := Matched(results)

	require.Len(t, matched, 2)
	require.Equal(t, "p-pass-1", matched[0].UID)
	require.Equal(t, "p-pass-2", matched[1].UID)
}

func TestFilter_EmptyPartnerList(t *testing.T) {
	req := Request{
		RequestID:  "req-1",
		Country:    "RU",
		DeviceType: DeviceMobile,
		BidFloor:   1.0,
	}

	results := Filter(req, []partner.Partner{})
	require.NotNil(t, results)
	require.Empty(t, results)

}

func TestFilter_RealPartnerSet(t *testing.T) {
	store := partner.NewInMemoryStore()

	req := Request{
		RequestID:  "req-1",
		Country:    "RU",
		DeviceType: DeviceMobile,
		BidFloor:   1.0,
	}

	results := Filter(req, store.All())
	matched := Matched(results)

	matchedUIDs := make([]string, 0, len(matched))
	for _, p := range matched {
		matchedUIDs = append(matchedUIDs, p.UID)
	}

	require.ElementsMatch(t, []string{"dsp-alpha", "dsp-gamma"}, matchedUIDs)
}
