package state

const (
	SERVER_NAME = "signalk-radar"
	VERSION     = "0.0.1"
)

type ServerState interface {
	GetName() string
	GetVersion() string
}
