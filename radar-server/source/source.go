package source

type FrameSourceFactory interface {
	CreateFrameSource(label string, port int) FrameSource
	Start()
	Stop()
}

type FrameSource interface {
	Label() string
	Source() chan []byte
}
