package radar

type RadarSource interface {
	Name() string
	Source() chan *RadarMessage
	Spokes() int
	MaxSpokeLen() int
}

type Radar struct {
	Id          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Spokes      int    `json:"spokes,omitempty"`
	MaxSpokeLen int    `json:"maxSpokeLen,omitempty"`
	StreamUrl   string `json:"streamUrl,omitempty"`
}
