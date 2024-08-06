package radar

type RadarSource interface {
	Label() string
	Source() chan *RadarMessage
	Spokes() int
	MaxSpokeLen() int
}

type Radar struct {
	Label       string `json:"label,omitempty"`
	Spokes      int    `json:"spokes,omitempty"`
	MaxSpokeLen int    `json:"maxSpokeLen,omitempty"`
	StreamUrl   string `json:"streamUrl,omitempty"`
}
