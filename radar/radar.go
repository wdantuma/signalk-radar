package radar

type RadarSource interface {
	Label() string
	Source() chan *RadarMessage
}
