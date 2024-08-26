package navico

import (
	"bytes"
	"encoding/binary"
	"log/slog"
	"time"
	"unsafe"

	"github.com/wdantuma/signalk-radar/radar-server/radar"
	"github.com/wdantuma/signalk-radar/radar-server/source"
)

const NAVICO_MAX_SPOKE_LEN = 1024
const NAVICO_SPOKES = 2048

type RadarReportType uint8

const (
	REPORT_TYPE_BR24 RadarReportType = 0x0f
	REPORT_TYPE_3G   RadarReportType = 0x08
	REPORT_TYPE_4G   RadarReportType = 0x01
	REPORT_TYPE_HALO RadarReportType = 0x00
)

type RadarType int

const (
	TYPE_UNKOWN RadarType = iota
	TYPE_BR24
	TYPE_3G
	TYPE_4GA
	TYPE_4GB
	TYPE_HALOA
	TYPE_HALOB
)

type LookupSpoke int

const (
	LOOKUP_SPOKE_LOW_NORMAL LookupSpoke = iota
	LOOKUP_SPOKE_LOW_BOTH
	LOOKUP_SPOKE_LOW_APPROACHING
	LOOKUP_SPOKE_HIGH_NORMAL
	LOOKUP_SPOKE_HIGH_BOTH
	LOOKUP_SPOKE_HIGH_APPROACHING
)

type navico struct {
	label              string
	radarType          RadarType
	doppler            radar.DopplerMode
	legend             radar.Legend
	pixelToBlob        []uint8
	source             chan *radar.RadarMessage
	farmeSourceFactory source.FrameSourceFactory
	locatorSource      source.FrameSource
	reportSource       source.FrameSource
	dataSource         source.FrameSource
}

type RadarReport_01B2 struct {
	Id          uint16
	Serialno    [16]uint8
	Addr0       source.Address
	U1          [12]uint8
	Addr1       source.Address
	U2          [4]uint8
	Addr2       source.Address
	U3          [10]uint8
	Addr3       source.Address
	U4          [4]uint8
	Addr4       source.Address
	U5          [10]uint8
	AddrDataA   source.Address
	U6          [4]uint8
	AddrSendA   source.Address
	U7          [4]uint8
	AddrReportA source.Address
	U8          [10]uint8
	AddrDataB   source.Address
	U9          [4]uint8
	AddrSendB   source.Address
	U10         [4]uint8
	AddrReportB source.Address
	U11         [10]uint8
	Addr11      source.Address
	U12         [4]uint8
	Addr12      source.Address
	U13         [4]uint8
	Addr13      source.Address
	U14         [10]uint8
	Addr14      source.Address
	U15         [4]uint8
	Addr15      source.Address
	U16         [4]uint8
	Addr16      source.Address
}

type RadarReport_03C4_129 struct {
	What          uint8
	Command       uint8
	Radar_type    RadarReportType // I hope! 01 = 4G and new 3G, 08 = 3G, 0F = BR24, 00 = HALO
	U00           [31]uint8       // Lots of unknown
	Hours         uint32          // Hours of operation
	U01           [20]uint8       // Lots of unknown
	Firmware_date [16]uint16
	Firmware_time [16]uint16
	U02           [7]uint8
}

type Common_header struct {
	HeaderLen uint8 // 1 bytes
	Status    uint8 // 1 bytes
}

type Br24_header struct {
	Scan_number uint16   // 2 bytes, 0-4095
	Mark        [4]uint8 // 4 bytes 0x00, 0x44, 0x0d, 0x0e
	Angle       [2]uint8 // 2 bytes
	Heading     [2]uint8 // 2 bytes heading with RI-10/11. See bitmask explanation above.
	Range       [4]uint8 // 4 bytes
	U01         [2]uint8 // 2 bytes blank
	U02         [2]uint8 // 2 bytes
	U03         [4]uint8 // 4 bytes blank
}

type Br4g_header struct {
	Scan_number uint16   // 2 bytes, 0-4095
	U00         [2]uint8 // Always 0x4400 (integer)
	Largerange  uint16   // 2 bytes or -1
	Angle       uint16   // 2 bytes
	Heading     uint16   // 2 bytes heading with RI-10/11 or -1. See bitmask explanation above.
	Smallrange  uint16   // 2 bytes or -1
	Rotation    uint16   // 2 bytes, rotation/angle
	U02         [4]uint8 // 4 bytes signed integer, always -1
	U03         [4]uint8 // 4 bytes signed integer, mostly -1 (0x80 in last byte) or 0xa0 in last byte
}

func lookupIndex(t LookupSpoke, i int) int {
	return int(t)*256 + int(i)
}

func (r *navico) InitializeLookupData() {
	var lookupData []uint8 = make([]uint8, 6*256)
	for j := 0; j < 256; j++ {
		low := uint8(j) & 0x0f
		high := (uint8(j) >> 4) & 0x0f

		lookupData[lookupIndex(LOOKUP_SPOKE_LOW_NORMAL, j)] = low
		switch low {
		case 0x0f:
			lookupData[lookupIndex(LOOKUP_SPOKE_LOW_BOTH, j)] = r.legend.DopplerApproaching
		case 0x0e:
			lookupData[lookupIndex(LOOKUP_SPOKE_LOW_BOTH, j)] = r.legend.DopplerReceding
		default:
			lookupData[lookupIndex(LOOKUP_SPOKE_LOW_BOTH, j)] = low
		}
		switch low {
		case 0x0f:
			lookupData[lookupIndex(LOOKUP_SPOKE_LOW_APPROACHING, j)] = r.legend.DopplerApproaching
		default:
			lookupData[lookupIndex(LOOKUP_SPOKE_LOW_APPROACHING, j)] = low
		}

		lookupData[lookupIndex(LOOKUP_SPOKE_HIGH_NORMAL, j)] = high
		switch high {
		case 0x0f:
			lookupData[lookupIndex(LOOKUP_SPOKE_HIGH_BOTH, j)] = r.legend.DopplerApproaching
		case 0x0e:
			lookupData[lookupIndex(LOOKUP_SPOKE_HIGH_BOTH, j)] = r.legend.DopplerReceding
		default:
			lookupData[lookupIndex(LOOKUP_SPOKE_HIGH_BOTH, j)] = high
		}
		switch high {
		case 0x0f:
			lookupData[lookupIndex(LOOKUP_SPOKE_HIGH_APPROACHING, j)] = r.legend.DopplerApproaching
		default:
			lookupData[lookupIndex(LOOKUP_SPOKE_HIGH_APPROACHING, j)] = high
		}
	}
	r.pixelToBlob = lookupData
}

func NewNavico(frameSourceFactory source.FrameSourceFactory) *navico {

	locatorSource := frameSourceFactory.CreateFrameSource("Navico locator", source.NewAddress(236, 6, 7, 5, 6878))
	navico := &navico{radarType: TYPE_UNKOWN, label: "Navico", farmeSourceFactory: frameSourceFactory, locatorSource: locatorSource, source: make(chan *radar.RadarMessage), reportSource: nil, dataSource: nil, doppler: radar.Both, legend: radar.DefaultLegend(true, 16), pixelToBlob: make([]uint8, 0)}
	navico.InitializeLookupData()

	navico.start()
	locatorSource.Start()

	return navico
}

func (g *navico) Source() chan *radar.RadarMessage {
	return g.source
}

func (g *navico) Name() string {
	return g.label
}

func (g *navico) Spokes() int {
	return NAVICO_SPOKES
}

func (g *navico) MaxSpokeLen() int {
	return NAVICO_MAX_SPOKE_LEN
}

func (g *navico) Legend() radar.Legend {
	return g.legend
}

func (g *navico) start() {
	go func() {
		locatorSource := g.locatorSource.Source()
		var reportSource source.FrameSource
		var dataSource source.FrameSource
		for {
			var reportSourceChan chan []byte = nil
			if reportSource != nil {
				reportSourceChan = reportSource.Source()
			}
			var dataSourceChan chan []byte = nil
			if dataSource != nil {
				dataSourceChan = dataSource.Source()
			}
			select {
			case data, ok := <-reportSourceChan:
				if ok {
					g.processReport(data)
				} else {
					reportSource = nil
				}
			case data, ok := <-dataSourceChan:
				if ok {
					g.processData(data)
				} else {
					dataSource = nil
				}
			case data, ok := <-locatorSource:
				if ok {
					newDataSource, newReportSource := g.processLocator(data)
					if newDataSource != nil && newReportSource != nil {
						if dataSource != nil {
							g.farmeSourceFactory.RemoveFrameSource(dataSource)
							dataSource.Stop()
						}
						dataSource = newDataSource
						newDataSource.Start()
						if reportSource != nil {
							g.farmeSourceFactory.RemoveFrameSource(reportSource)
							reportSource.Stop()
						}
						reportSource = newReportSource
						newReportSource.Start()
					}
				} else {
					locatorSource = nil
				}
			}
			if dataSource == nil && reportSource == nil && locatorSource == nil {
				break
			}
		}
		close(g.source)
	}()
}

func (g *navico) processLocator(locatorBytes []byte) (data source.FrameSource, report source.FrameSource) {
	if len(locatorBytes) >= 222 {
		if locatorBytes[0] == 0x01 && locatorBytes[1] == 0xb2 {
			dataReader := bytes.NewReader(locatorBytes)
			var report RadarReport_01B2
			err := binary.Read(dataReader, binary.BigEndian, &report)

			if err == nil {
				reportSource := g.farmeSourceFactory.CreateFrameSource("Navico report", report.AddrReportA)
				dataSource := g.farmeSourceFactory.CreateFrameSource("Navico data", report.AddrDataA)
				return dataSource, reportSource
			}
		}
	}
	return nil, nil
}

func (g *navico) processReport(reportBytes []byte) {

	var len uint16 = uint16(len(reportBytes))

	if len > 1 && reportBytes[1] == 0xc4 {
		reportReader := bytes.NewReader(reportBytes)
		switch (len << 8) + uint16(reportBytes[0]) {
		case (18 << 8) + 0x01:
			slog.Debug("Report 01c4_18")
		case (99 << 8) + 0x02:
			slog.Debug("Report 02c4_99")
		case (129 << 8) + 0x03:
			slog.Debug("Report 03c4_129")
			var data RadarReport_03C4_129
			_ = binary.Read(reportReader, binary.LittleEndian, &data)
			switch data.Radar_type {
			case REPORT_TYPE_BR24:
				g.radarType = TYPE_BR24
				slog.Debug("BR24")
			case REPORT_TYPE_3G:
				g.radarType = TYPE_3G
				slog.Debug("3G")
			case REPORT_TYPE_4G:
				g.radarType = TYPE_4GA
				slog.Debug("4G")
			case REPORT_TYPE_HALO:
				g.radarType = TYPE_HALOA
				slog.Debug("HALO")
			default:
				slog.Debug("Unkown radar type")
			}
		case (66 << 8) + 0xc4:
			slog.Debug("Report 04c4_66")
		case (68 << 8) + 0x06:
			slog.Debug("Report 06c4_68")
		case (74 << 8) + 0x06:
			slog.Debug("Report 06c4_74")
		case (22 << 8) + 0x08, (21 << 8) + 0x08:
			slog.Debug("Report 08c4_21")
		case (18 << 8) + 0x08:
			slog.Debug("Report 08c4_18")
		case (66 << 8) + 0x12:
			slog.Debug("Report 12c4_66")
		default:
			slog.Debug("Received unkown radar report")
		}
	}

}

func modSpokes(angle uint32) uint32 {
	return angle + (2*NAVICO_SPOKES)%NAVICO_SPOKES
}

func (g *navico) processData(dataBytes []byte) {

	if len(dataBytes) < 9 {
		slog.Debug("Strange header length")
		return
	}

	dataReader := bytes.NewReader(dataBytes)
	dataReader.Seek(8, 0) // skip header

	var br4g Br4g_header
	len := len(dataBytes)
	spokes := (len - 8) / (int(unsafe.Sizeof(&br4g)) + (NAVICO_MAX_SPOKE_LEN / 2))
	if spokes != 32 {
	}

	message := radar.RadarMessage{
		Spokes: make([]*radar.RadarMessage_Spoke, spokes),
	}
	for scanline := 0; scanline < spokes; scanline++ {

		var common Common_header
		data := make([]byte, (NAVICO_MAX_SPOKE_LEN / 2))
		binary.Read(dataReader, binary.LittleEndian, &common)

		if common.HeaderLen != 0x18 {
			slog.Debug("Strange header length")
			return
		}

		if common.Status != 0x02 && common.Status != 0x12 {
			slog.Debug("Strange status")
			return
		}

		var range_meters int
		//var heading_raw uint16

		switch g.radarType {
		case TYPE_3G, TYPE_4GA, TYPE_4GB:
			_ = binary.Read(dataReader, binary.LittleEndian, &br4g)
			//heading_raw = br4g.Heading
			if br4g.Largerange == 0x80 {
				if br4g.Smallrange == 0xffff {
					range_meters = 0
				} else {
					range_meters = int(br4g.Smallrange) / 4
				}
			} else {
				range_meters = int(br4g.Largerange) * 64
			}
		case TYPE_HALOA, TYPE_HALOB:
			_ = binary.Read(dataReader, binary.LittleEndian, &br4g)
			//heading_raw = br4g.Heading
			if br4g.Largerange == 0x80 {
				if br4g.Smallrange == 0xffff {
					range_meters = 0
				} else {
					range_meters = int(br4g.Smallrange) / 4
				}
			} else {
				range_meters = int(br4g.Largerange) * (int(br4g.Smallrange) / 512)
			}
		}

		binary.Read(dataReader, binary.LittleEndian, &data)

		var lowNibbleIndex LookupSpoke
		switch g.doppler {
		case radar.None:
			lowNibbleIndex = LOOKUP_SPOKE_LOW_NORMAL
		case radar.Both:
			lowNibbleIndex = LOOKUP_SPOKE_LOW_BOTH
		case radar.Approaching:
			lowNibbleIndex = LOOKUP_SPOKE_LOW_APPROACHING
		}
		var highNibbleIndex LookupSpoke
		switch g.doppler {
		case radar.None:
			highNibbleIndex = LOOKUP_SPOKE_HIGH_NORMAL
		case radar.Both:
			highNibbleIndex = LOOKUP_SPOKE_HIGH_BOTH
		case radar.Approaching:
			highNibbleIndex = LOOKUP_SPOKE_HIGH_APPROACHING
		}

		var data_highres []uint8 = make([]uint8, NAVICO_MAX_SPOKE_LEN)

		for i := 0; i < NAVICO_MAX_SPOKE_LEN/2; i++ {
			data_highres[2*i] = g.pixelToBlob[lookupIndex(lowNibbleIndex, int(data[i]))]
			data_highres[2*i+1] = g.pixelToBlob[lookupIndex(highNibbleIndex, int(data[i]))]
		}

		time := uint64(time.Now().UnixMilli())
		message.Spokes[scanline] = &radar.RadarMessage_Spoke{
			Angle: uint32(modSpokes(uint32(br4g.Angle / 2))),
			Range: uint32(range_meters),
			Data:  data_highres,
			Time:  &time,
		}
	}
	g.source <- &message
}
