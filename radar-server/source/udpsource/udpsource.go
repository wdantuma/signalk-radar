package udpsource

import (
	"fmt"
	"log"
	"net"
	"slices"

	"github.com/wdantuma/signalk-radar/radar-server/source"
)

const (
	MAX_DGRAM_SIZE = 64 * 1024
)

type udpFrameSource struct {
	running bool
	label   string
	address source.Address
	source  chan []byte
}

func (fs *udpFrameSource) Label() string {
	return fs.label
}

func (fs *udpFrameSource) Source() chan []byte {
	return fs.source
}

func (fs *udpFrameSource) Address() source.Address {
	return fs.address
}

func (fs *udpFrameSource) Start() {
	fs.running = true
	go func() {
		addr, err := net.ResolveUDPAddr("udp", fs.address.String())
		fmt.Println(fs.address)
		if err != nil {
			log.Fatal(err)
		}
		ifs, _ := net.Interfaces() // for now only listen on loopback
		l, err := net.ListenMulticastUDP("udp4", &ifs[0], addr)
		if err != nil {
			panic("Error listing UDP socket")
		}
		l.SetReadBuffer(MAX_DGRAM_SIZE)
		defer l.Close()
		buf := make([]byte, MAX_DGRAM_SIZE)
		for fs.running {
			n, _, _ := l.ReadFromUDP(buf)
			if fs.running {
				fs.source <- buf[:n]
			}
		}
	}()
}

func (fs *udpFrameSource) Stop() {
	fs.running = false
	close(fs.source)
}

type udpSource struct {
	running bool
	sources []source.FrameSource
}

func NewUdpSource() (*udpSource, error) {
	return &udpSource{sources: make([]source.FrameSource, 0)}, nil
}

func (p *udpSource) CreateFrameSource(label string, address source.Address) source.FrameSource {
	for _, s := range p.sources {
		if s.Label() == label && s.Address().String() == address.String() {
			return nil // reuse existing
		}
	}
	entry := &udpFrameSource{address: address, label: label, source: make(chan []byte, 10)}
	p.sources = append(p.sources, entry)
	return entry
}

func (p *udpSource) RemoveFrameSource(source source.FrameSource) {
	index := slices.Index(p.sources, source)
	p.sources = append(p.sources[:index], p.sources[index+1:]...)
}

func (p *udpSource) Start() {
	p.running = true
}

func (p *udpSource) Stop() {
	if p.running {
		p.running = false
		for _, e := range p.sources {
			e.Stop()
		}
	}
}
