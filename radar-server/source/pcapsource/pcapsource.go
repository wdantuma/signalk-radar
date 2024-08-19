package pcapsource

import (
	"errors"
	"fmt"
	"log"
	"os"
	"slices"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/ip4defrag"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/wdantuma/signalk-radar/radar-server/source"
)

type pcapFrameSource struct {
	label   string
	address source.Address
	source  chan []byte
}

func (fs *pcapFrameSource) Label() string {
	return fs.label
}

func (fs *pcapFrameSource) Source() chan []byte {
	return fs.source
}

func (fs *pcapFrameSource) Address() source.Address {
	return fs.address
}

func (fs *pcapFrameSource) Start() {

}

func (fs *pcapFrameSource) Stop() {
	close(fs.source)
}

type pcapSource struct {
	running bool
	loop    bool
	file    string
	sources []source.FrameSource
}

func NewPcapSource(file string, loop bool) (*pcapSource, error) {
	if _, err := os.Stat(file); err == nil {
		return &pcapSource{file: file, sources: make([]source.FrameSource, 0), loop: loop}, nil
	} else {
		return nil, errors.New(fmt.Sprintf("Pcap file %s not found\n", file))
	}
}

func (p *pcapSource) CreateFrameSource(label string, address source.Address) source.FrameSource {
	for _, s := range p.sources {
		if s.Label() == label && s.Address().String() == address.String() {
			return nil // reuse existing
		}
	}
	entry := &pcapFrameSource{address: address, label: label, source: make(chan []byte, 10)}
	p.sources = append(p.sources, entry)
	return entry
}

func (p *pcapSource) RemoveFrameSource(source source.FrameSource) {
	index := slices.Index(p.sources, source)
	p.sources = append(p.sources[:index], p.sources[index+1:]...)
}

func (p *pcapSource) Label() string {
	return fmt.Sprintf("Pcap source %s", p.file)
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
		n := 0
		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		defragger := ip4defrag.NewIPv4Defragmenter()
		var prevTimestamp time.Time = time.Time{}
		var startTimestamp = time.Now()
		for packet := range packetSource.Packets() {
			n++
			ip4Layer := packet.Layer(layers.LayerTypeIPv4)
			if ip4Layer == nil {
				continue
			}
			ip4 := ip4Layer.(*layers.IPv4)
			l := ip4.Length

			newip4, err := defragger.DefragIPv4(ip4)
			if err != nil {
				log.Fatalln("Error while de-fragmenting", err)
			} else if newip4 == nil {
				continue // packet fragment, we don't have whole packet yet.
			}
			if newip4.Length != l {
				pb, ok := packet.(gopacket.PacketBuilder)
				if !ok {
					panic("Not a PacketBuilder")
				}
				nextDecoder := newip4.NextLayerType()
				nextDecoder.Decode(newip4.Payload, pb)
			}
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
		totalDuration := time.Since(startTimestamp)
		fmt.Printf("Processed :%d packets in %s\n", n, totalDuration)
	}
}

func (p *pcapSource) Stop() {
	if p.running {
		p.running = false
		for _, e := range p.sources {
			e.Stop()
		}
	}
}

func (p *pcapSource) handlePacket(packet gopacket.Packet) {

	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	udpLayer := packet.Layer(layers.LayerTypeUDP)

	if ipLayer != nil && udpLayer != nil {
		dstPort := udpLayer.(*layers.UDP).DstPort
		dstIpAddr := ipLayer.(*layers.IPv4).DstIP
		dstAddr := source.NewAddress(dstIpAddr[0], dstIpAddr[1], dstIpAddr[2], dstIpAddr[3], uint16(dstPort))
		for _, e := range p.sources {
			if e.Address().IsMatch(dstAddr) {
				e.Source() <- udpLayer.LayerPayload()
			}
		}
	}
}
