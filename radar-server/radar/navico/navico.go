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

var lookupData []uint8 = make([]uint8, 6*256)

var lookupNibbleToByte = [...]uint8{
	0,    // 0
	0x32, // 1
	0x40, // 2
	0x4e, // 3
	0x5c, // 4
	0x6a, // 5
	0x78, // 6
	0x86, // 7
	0x94, // 8
	0xa2, // 9
	0xb0, // a
	0xbe, // b
	0xcc, // c
	0xda, // d
	0xe8, // e
	0xf4, // f
}

type navico struct {
	label              string
	radarType          RadarType
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

func InitializeLookupData() {
	if lookupData[lookupIndex(5, 255)] == 0 {
		for j := 0; j <= 255; j++ {
			low := lookupNibbleToByte[(j & 0x0f)]
			high := lookupNibbleToByte[(j&0xf0)>>4]

			lookupData[lookupIndex(LOOKUP_SPOKE_LOW_NORMAL, j)] = low
			lookupData[lookupIndex(LOOKUP_SPOKE_HIGH_NORMAL, j)] = high

			switch low {
			case 0xf4:
				lookupData[lookupIndex(LOOKUP_SPOKE_LOW_BOTH, j)] = 0xff
				lookupData[lookupIndex(LOOKUP_SPOKE_LOW_APPROACHING, j)] = 0xff
			case 0xe8:
				lookupData[lookupIndex(LOOKUP_SPOKE_LOW_BOTH, j)] = 0xfe
				lookupData[lookupIndex(LOOKUP_SPOKE_LOW_APPROACHING, j)] = low
			default:
				lookupData[lookupIndex(LOOKUP_SPOKE_LOW_BOTH, j)] = low
				lookupData[lookupIndex(LOOKUP_SPOKE_LOW_APPROACHING, j)] = low
			}
			switch high {
			case 0xf4:
				lookupData[lookupIndex(LOOKUP_SPOKE_HIGH_BOTH, j)] = 0xff
				lookupData[lookupIndex(LOOKUP_SPOKE_HIGH_APPROACHING, j)] = 0xff
			case 0xe8:
				lookupData[lookupIndex(LOOKUP_SPOKE_HIGH_BOTH, j)] = 0xfe
				lookupData[lookupIndex(LOOKUP_SPOKE_HIGH_APPROACHING, j)] = high
			default:
				lookupData[lookupIndex(LOOKUP_SPOKE_HIGH_BOTH, j)] = high
				lookupData[lookupIndex(LOOKUP_SPOKE_HIGH_APPROACHING, j)] = high
			}
		}
	}
}

func NewNavico(frameSourceFactory source.FrameSourceFactory) *navico {

	InitializeLookupData()
	locatorSource := frameSourceFactory.CreateFrameSource("Navico locator", source.NewAddress(236, 6, 7, 5, 6878))
	navico := &navico{radarType: TYPE_UNKOWN, label: "Navico", farmeSourceFactory: frameSourceFactory, locatorSource: locatorSource, source: make(chan *radar.RadarMessage), reportSource: nil, dataSource: nil}
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

func (g *navico) Legend() map[string]radar.LegendEntry {
	legend := map[string]radar.LegendEntry{
		"43": {Type: "Normal", Colour: "#a93d1eff"},
		"44": {Type: "Normal", Colour: "#b84c0fff"},
		"32": {Type: "Normal", Colour: "#006bc8ff"},
		"36": {Type: "Normal", Colour: "#3d2e8aff"},
		"41": {Type: "Normal", Colour: "#8a1e3dff"},
		"33": {Type: "Normal", Colour: "#0f5cb8ff"},
		"39": {Type: "Normal", Colour: "#6b005cff"},
		"34": {Type: "Normal", Colour: "#1e4ca9ff"},
		"37": {Type: "Normal", Colour: "#4c1e7bff"},
		"42": {Type: "Normal", Colour: "#992e2eff"},
		"35": {Type: "Normal", Colour: "#2e3d99ff"},
		"38": {Type: "Normal", Colour: "#5c0f6bff"},
		"40": {Type: "Normal", Colour: "#7b0f4cff"},
		"13": {Type: "History", Colour: "#ffffff34"},
		"14": {Type: "History", Colour: "#ffffff38"},
		"7":  {Type: "History", Colour: "#ffffff1c"},
		"17": {Type: "History", Colour: "#ffffff44"},
		"18": {Type: "History", Colour: "#ffffff48"},
		"12": {Type: "History", Colour: "#ffffff30"},
		"5":  {Type: "History", Colour: "#ffffff14"},
		"0":  {Type: "History", Colour: "#ffffff00"},
		"26": {Type: "History", Colour: "#ffffff68"},
		"3":  {Type: "History", Colour: "#ffffff0c"},
		"30": {Type: "History", Colour: "#ffffff78"},
		"6":  {Type: "History", Colour: "#ffffff18"},
		"4":  {Type: "History", Colour: "#ffffff10"},
		"22": {Type: "History", Colour: "#ffffff58"},
		"15": {Type: "History", Colour: "#ffffff3c"},
		"10": {Type: "History", Colour: "#ffffff28"},
		"1":  {Type: "History", Colour: "#ffffff04"},
		"28": {Type: "History", Colour: "#ffffff70"},
		"16": {Type: "History", Colour: "#ffffff40"},
		"25": {Type: "History", Colour: "#ffffff64"},
		"9":  {Type: "History", Colour: "#ffffff24"},
		"27": {Type: "History", Colour: "#ffffff6c"},
		"31": {Type: "History", Colour: "#ffffff7c"},
		"21": {Type: "History", Colour: "#ffffff54"},
		"24": {Type: "History", Colour: "#ffffff60"},
		"29": {Type: "History", Colour: "#ffffff74"},
		"20": {Type: "History", Colour: "#ffffff50"},
		"23": {Type: "History", Colour: "#ffffff5c"},
		"2":  {Type: "History", Colour: "#ffffff08"},
		"8":  {Type: "History", Colour: "#ffffff20"},
		"11": {Type: "History", Colour: "#ffffff2c"},
		"19": {Type: "History", Colour: "#ffffff4c"},
		"46": {Type: "DopplerApproaching", Colour: "#00c8c8ff"},
		"47": {Type: "DopplerReceding", Colour: "#90d0f0ff"},
		"45": {Type: "TargetBorder", Colour: "#c8c8c8ff"},
	}

	return legend
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

		var data_highres []uint8 = make([]uint8, NAVICO_MAX_SPOKE_LEN)

		doppler := 0 //TODO set this
		for i := 0; i < NAVICO_MAX_SPOKE_LEN/2; i++ {
			data_highres[2*i] = lookupData[lookupIndex(LOOKUP_SPOKE_LOW_NORMAL+LookupSpoke(doppler), int(data[i]))]
			data_highres[2*i+1] = lookupData[lookupIndex(LOOKUP_SPOKE_HIGH_NORMAL+LookupSpoke(doppler), int(data[i]))]
		}

		range_meters = int(float32(range_meters) * 1.66) // strange factor needed to display correctly on map

		message.Spokes[scanline] = &radar.RadarMessage_Spoke{
			Angle:   uint32(modSpokes(uint32(br4g.Angle / 2))),
			Bearing: 0,
			Range:   uint32(range_meters),
			Data:    data_highres,
			Time:    uint64(time.Now().UnixMilli()),
		}
	}
	g.source <- &message
}
