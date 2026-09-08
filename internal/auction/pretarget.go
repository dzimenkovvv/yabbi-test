package auction

import "yabbi_test/internal/partner"

type FilterResult struct {
	Partner partner.Partner
	Passed  bool
	Reason  string
}

func Filter(req Request, partners []partner.Partner) []FilterResult {
	results := make([]FilterResult, 0, len(partners))

	for _, p := range partners {
		passed, reason := evaluate(req, p)
		results = append(results, FilterResult{
			Partner: p,
			Passed:  passed,
			Reason:  reason,
		})
	}

	return results
}

func Matched(results []FilterResult) []partner.Partner {
	matched := make([]partner.Partner, 0, len(results))
	for _, r := range results {
		if r.Passed {
			matched = append(matched, r.Partner)
		}
	}
	return matched
}

func evaluate(req Request, p partner.Partner) (passed bool, reason string) {
	if !isEnabled(p) {
		return false, "partner is disabled"
	}

	if !matchesCountry(p, req.Country) {
		return false, "country not served"
	}

	if !matchesDeviceType(p, req.DeviceType) {
		return false, "device type not served"
	}

	if !meetsBidFloor(p, req.BidFloor) {
		return false, "bid floor too low"
	}

	if blocked, category := hasBlockedCategory(p, req.Categories); blocked {
		return false, "blocked category: " + category
	}

	return true, ""
}

func isEnabled(p partner.Partner) bool {
	return p.IsEnabled
}

func matchesCountry(p partner.Partner, country string) bool {
	if len(p.Countries) == 0 {
		return true
	}

	return contains(p.Countries, country)
}

func matchesDeviceType(p partner.Partner, deviceType string) bool {
	if len(p.DeviceTypes) == 0 {
		return true
	}

	return contains(p.DeviceTypes, deviceType)
}

func meetsBidFloor(p partner.Partner, bidFloor float64) bool {
	return bidFloor >= p.MinBidFloor
}

func hasBlockedCategory(p partner.Partner, categories []string) (blocked bool, category string) {
	for _, c := range categories {
		if contains(p.BlockedCategories, c) {
			return true, c
		}
	}

	return false, ""
}

func contains(list []string, value string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}

	return false
}
