package garminxhd

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/wdantuma/signalk-radar/radar"
	"github.com/wdantuma/signalk-radar/source"
)

const GARMIN_XHD_MAX_SPOKE_LEN = 705

type garminxhd struct {
	label        string
	source       chan radar.RadarMessage
	reportSource source.FrameSource
	dataSource   source.FrameSource
}

type RadarLine struct {
	PacketType       uint32
	Len1             uint32
	Fill_1           uint16
	ScanLength       uint16
	Angle            uint16
	Fill_2           uint16
	RangeMeters      uint32
	DisplayMeters    uint32
	Fill_3           uint16
	ScanLengthBytesS uint16
	Fill_4           uint16
	ScanLengthBytesI uint32
	Fill_5           uint16
	//line_data        [GARMIN_XHD_MAX_SPOKE_LEN]uint8
}

func NewGarminXhd(reportSource source.FrameSource, dataSource source.FrameSource) *garminxhd {

	garminxhd := &garminxhd{label: "GarminXHD", reportSource: reportSource, dataSource: dataSource, source: make(chan radar.RadarMessage)}
	garminxhd.start()

	return garminxhd
}

func (g *garminxhd) Source() chan radar.RadarMessage {
	return g.source
}

func (g *garminxhd) Label() string {
	return g.label
}

func (g *garminxhd) start() {
	go func() {
		reportSource := g.reportSource.Source()
		dataSource := g.dataSource.Source()
		for {
			select {
			case data, ok := <-reportSource:
				if ok {
					g.processReport(data)
				} else {
					reportSource = nil
				}
			case data, ok := <-dataSource:
				if ok {
					g.processData(data)
				} else {
					dataSource = nil
				}
			}
			if dataSource == nil && reportSource == nil {
				break
			}
		}
		close(g.source)
	}()
}

func (g *garminxhd) processReport(reportBytes []byte) {
	reportReader := bytes.NewReader(reportBytes)
	var packetType uint32
	err := binary.Read(reportReader, binary.LittleEndian, &packetType)
	if err == nil {
		fmt.Printf("0x%X\n", packetType)
	}
}

func (g *garminxhd) processData(dataBytes []byte) {
	dataReader := bytes.NewReader(dataBytes)
	var line RadarLine
	err := binary.Read(dataReader, binary.LittleEndian, &line)
	if err == nil {
		//data := io.ReadFull(dataReader,10)
		fmt.Printf("%d,%d,%d\n", line.Angle/8, line.DisplayMeters, line.ScanLength)
	} else {
		fmt.Printf("%s\n", err)
	}
}
