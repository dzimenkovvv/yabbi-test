package partner

type Partner struct {
	UID       string `json:"uid"`
	Name      string `json:"name"`
	Endpoint  string `json:"endpoint"`
	IsEnabled bool   `json:"is_enabled"`

	Countries []string `json:"countries"`

	DeviceTypes []string `json:"device_types"`

	MinBidFloor float64 `json:"min_bid_floor"`

	BlockedCategories []string `json:"blocked_categories"`
}
