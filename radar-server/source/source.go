package source

type FrameSourceFactory interface {
	CreateFrameSource(label string, address Address) FrameSource
	RemoveFrameSource(source FrameSource)
	Start()
	Stop()
}

type FrameSource interface {
	Label() string
	Source() chan []byte
	Address() Address
}
