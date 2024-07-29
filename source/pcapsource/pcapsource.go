package pcapsource

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/wdantuma/signalk-radar/source"
)

type pcapFrameSource struct {
	label  string
	port   int
	source chan []byte
}

func (fs *pcapFrameSource) Label() string {
	return fs.label
}

func (fs *pcapFrameSource) Source() chan []byte {
	return fs.source
}

type pcapSource struct {
	running bool
	loop    bool
	file    string
	sources []pcapFrameSource
}

func NewPcapSource(file string, loop bool) (*pcapSource, error) {
	if _, err := os.Stat(file); err == nil {
		return &pcapSource{file: file, sources: make([]pcapFrameSource, 0), loop: loop}, nil
	} else {
		return nil, errors.New(fmt.Sprintf("Pcap file %s not found\n", file))
	}
}

func (p *pcapSource) CreateFrameSource(label string, port int) source.FrameSource {
	entry := pcapFrameSource{port: port, label: label, source: make(chan []byte)}
	p.sources = append(p.sources, entry)
	return &entry
}

func (p *pcapSource) Start() {
	p.running = true
	go func() {
		if p.loop {
			for p.running && p.loop {
				p.processFile()
			}
		} else {
			p.processFile()
		}
		p.Stop()
	}()
}

func (p *pcapSource) processFile() {
	if handle, err := pcap.OpenOffline(p.file); err != nil {
		panic(err)
	} else {
		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		var prevTimestamp time.Time = time.Time{}
		for packet := range packetSource.Packets() {
			if !prevTimestamp.Equal(time.Time{}) {
				duration := packet.Metadata().Timestamp.Sub(prevTimestamp)
				time.Sleep(duration)
			}
			prevTimestamp = packet.Metadata().Timestamp
			if !p.running {
				break
			}
			p.handlePacket(packet)
		}
	}
}

func (p *pcapSource) Stop() {
	if p.running {
		p.running = false
		for _, e := range p.sources {
			close(e.source)
		}
	}
}

func (p *pcapSource) handlePacket(packet gopacket.Packet) {
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	udpLayer := packet.Layer(layers.LayerTypeUDP)
	if ipLayer != nil && udpLayer != nil {
		dtPort := udpLayer.(*layers.UDP).DstPort

		for _, e := range p.sources {
			if e.port == int(dtPort) {
				e.source <- udpLayer.LayerPayload()
			}
		}
	}
}
