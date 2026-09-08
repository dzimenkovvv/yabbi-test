package partner

type Store interface {
	All() []Partner
}

type InMemoryStore struct {
	partners []Partner
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		partners: []Partner{
			{
				UID:               "dsp-alpha",
				Name:              "DSP Alpha",
				Endpoint:          "http://localhost:9001/bid",
				IsEnabled:         true,
				Countries:         []string{"RU", "KZ"},
				DeviceTypes:       []string{"mobile", "desktop"},
				MinBidFloor:       0.5,
				BlockedCategories: []string{"gambling"},
			},
			{
				UID:               "dsp-beta",
				Name:              "DSP Beta",
				Endpoint:          "http://localhost:9002/bid",
				IsEnabled:         true,
				Countries:         nil, // работает со всеми странами
				DeviceTypes:       []string{"mobile"},
				MinBidFloor:       2.0,
				BlockedCategories: []string{"gambling", "adult"},
			},
			{
				UID:               "dsp-gamma",
				Name:              "DSP Gamma",
				Endpoint:          "http://localhost:9003/bid",
				IsEnabled:         true,
				Countries:         []string{"RU", "US", "DE"},
				DeviceTypes:       nil, // любое устройство
				MinBidFloor:       0,
				BlockedCategories: nil,
			},
			{
				UID:               "dsp-delta",
				Name:              "DSP Delta",
				Endpoint:          "http://localhost:9004/bid",
				IsEnabled:         false, // отключён — должен отсекаться всегда
				Countries:         []string{"RU"},
				DeviceTypes:       []string{"mobile", "desktop", "tv"},
				MinBidFloor:       0,
				BlockedCategories: nil,
			},
		},
	}
}

func (s *InMemoryStore) All() []Partner {
	return s.partners
}
