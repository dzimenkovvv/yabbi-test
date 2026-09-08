package auction

type Request struct {
	RequestID  string   `json:"request_id"`
	Country    string   `json:"country"`
	DeviceType string   `json:"device_type"`
	BidFloor   float64  `json:"bid_floor"`
	Categories []string `json:"categories"`
}

type Response struct {
	RequestID  string   `json:"request_id"`
	MatchedDSP []string `json:"matched_dsps"`
	Sent       int      `json:"sent"`
	Succeeded  int      `json:"succeeded"`
	DurationMs int64    `json:"duration_ms"`
}

const (
	DeviceMobile  = "mobile"
	DeviceDesktop = "desktop"
	DeviceTV      = "tv"
)

var ValidDeviceTypes = map[string]bool{
	DeviceMobile:  true,
	DeviceDesktop: true,
	DeviceTV:      true,
}
