package garminxhd

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"
	"unsafe"

	"github.com/wdantuma/signalk-radar/radar-server/radar"
	"github.com/wdantuma/signalk-radar/radar-server/source"
)

const GARMIN_XHD_MAX_SPOKE_LEN = 705
const GARMIN_XHD_SPOKES = 1440

type garminxhd struct {
	label        string
	source       chan *radar.RadarMessage
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

func NewGarminXhd(frameSourceFactory source.FrameSourceFactory) *garminxhd {
	reportSource := frameSourceFactory.CreateFrameSource("garminReport", source.NewAddress(0, 0, 0, 0, 50100))
	dataSource := frameSourceFactory.CreateFrameSource("garminData", source.NewAddress(0, 0, 0, 0, 50102))
	garminxhd := &garminxhd{label: "GarminXHD", reportSource: reportSource, dataSource: dataSource, source: make(chan *radar.RadarMessage)}
	garminxhd.start()

	return garminxhd
}

func (g *garminxhd) Source() chan *radar.RadarMessage {
	return g.source
}

func (g *garminxhd) Name() string {
	return g.label
}

func (g *garminxhd) Spokes() int {
	return GARMIN_XHD_SPOKES
}

func (g *garminxhd) MaxSpokeLen() int {
	return GARMIN_XHD_MAX_SPOKE_LEN
}

func (g *garminxhd) Legend() map[string]radar.LegendEntry {
	legend := map[string]radar.LegendEntry{
		"0": {Type: "normal", Colour: "#ffffffff"},
	}
	return legend
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
	_ = binary.Read(reportReader, binary.LittleEndian, &packetType)
	//if err == nil {
	//fmt.Printf("0x%X\n", packetType)
	//}
}

func (g *garminxhd) processData(dataBytes []byte) {
	dataReader := bytes.NewReader(dataBytes)
	var line RadarLine
	err := binary.Read(dataReader, binary.LittleEndian, &line)

	if err == nil {
		data := make([]byte, int(len(dataBytes)-int(unsafe.Sizeof(line))))
		err = binary.Read(dataReader, binary.LittleEndian, data)
		if err == nil {
			message := radar.RadarMessage{
				Spokes: make([]*radar.RadarMessage_Spoke, 1),
			}
			message.Spokes[0] = &radar.RadarMessage_Spoke{
				Angle:   uint32(line.Angle / 8),
				Bearing: 0,
				Range:   uint32(line.RangeMeters),
				Data:    data,
				Time:    uint64(time.Now().UnixMilli()),
			}
			g.source <- &message
		}
	} else {
		fmt.Printf("%s\n", err)
	}
}
