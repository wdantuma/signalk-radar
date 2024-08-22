package radar

type LegendEntry struct {
	Type   string
	Colour string
}

type RadarSource interface {
	Name() string
	Source() chan *RadarMessage
	Spokes() int
	MaxSpokeLen() int
	Legend() map[string]LegendEntry
}

type Radar struct {
	Id          string                 `json:"id,omitempty"`
	Name        string                 `json:"name,omitempty"`
	Spokes      int                    `json:"spokes,omitempty"`
	MaxSpokeLen int                    `json:"maxSpokeLen,omitempty"`
	StreamUrl   string                 `json:"streamUrl,omitempty"`
	Legend      map[string]LegendEntry `json:"legend,omitempty"`
}
