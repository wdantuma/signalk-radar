package radar

import (
	"encoding/json"
	"fmt"
)

const (
	BLOB_HISTORY_COLORS uint8 = 32
)

type DopplerMode int

const (
	None DopplerMode = iota
	Both
	Approaching
)

func (d DopplerMode) String() string {
	switch d {
	case None:
		return "None"
	case Both:
		return "Both"
	case Approaching:
		return "Approaching"
	default:
		return ""
	}
}

func (d DopplerMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

type PixelType int

const (
	History PixelType = iota + 1
	TargetBorder
	DopplerApproaching
	DopplerReceding
	Normal
)

func (p PixelType) String() string {
	switch p {
	case History:
		return "History"
	case TargetBorder:
		return "TargetBorder"
	case DopplerApproaching:
		return "DopplerApproaching"
	case DopplerReceding:
		return "DopplerReceding"
	case Normal:
		return "Normal"
	default:
		return ""
	}
}

func (p PixelType) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

type Color struct {
	R uint8
	G uint8
	B uint8
	A uint8
}

func (c Color) MarshalJSON() ([]byte, error) {

	return json.Marshal(fmt.Sprintf("#%02x%02x%02x%02x", c.R, c.G, c.B, c.A))
}

type LegendEntry struct {
	Type  PixelType `json:"type,omitempty"`
	Color Color     `json:"color,omitempty"`
}

type Legend struct {
	Pixels             []LegendEntry
	Border             uint8
	DopplerApproaching uint8
	DopplerReceding    uint8
	HistoryStart       uint8
}

func (legend Legend) MarshalJSON() ([]byte, error) {
	l := make(map[string]LegendEntry, 0)
	for i, e := range legend.Pixels {
		l[fmt.Sprintf("%d", i)] = e
	}
	return json.Marshal(l)
}

type RadarSource interface {
	Name() string
	Source() chan *RadarMessage
	Spokes() int
	MaxSpokeLen() int
	Legend() Legend
}

type Radar struct {
	Id          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Spokes      int    `json:"spokes,omitempty"`
	MaxSpokeLen int    `json:"maxSpokeLen,omitempty"`
	StreamUrl   string `json:"streamUrl,omitempty"`
	Legend      Legend `json:"legend,omitempty"`
}

func DefaultLegend(doppler bool, pixelValues int) Legend {
	legend := Legend{Pixels: make([]LegendEntry, 0), HistoryStart: 0, Border: 0, DopplerApproaching: 0, DopplerReceding: 0}

	if pixelValues > 255-32-2 {
		pixelValues = 255 - 32 - 2
	}

	const WHITE float32 = 255
	pixelsWithColor := pixelValues - 1
	start := WHITE / 3.0
	delta := WHITE * 2.0 / float32(pixelsWithColor)
	oneThird := pixelsWithColor / 3
	twoThirds := oneThird * 2

	legend.Pixels = append(legend.Pixels, LegendEntry{Type: Normal, Color: Color{R: 0, G: 0, B: 0, A: 0}})
	for v := 1; v <= int(pixelValues); v++ {
		var r uint8 = 0
		if v >= twoThirds {
			r = uint8(start + float32(v-twoThirds)*delta)
		}
		var g uint8 = 0
		if v >= oneThird && v < twoThirds {
			g = uint8(start + float32(v-oneThird)*delta)
		}
		var b uint8 = 0
		if v < oneThird {
			b = uint8(start + float32(v)*(WHITE/float32(pixelValues)))
		}
		legend.Pixels = append(legend.Pixels, LegendEntry{Type: Normal, Color: Color{R: r, G: g, B: b, A: 255}})
	}
	legend.Border = uint8(len(legend.Pixels))
	legend.Pixels = append(legend.Pixels, LegendEntry{Type: TargetBorder, Color: Color{R: 200, G: 200, B: 200, A: 255}})
	if doppler {
		legend.DopplerApproaching = uint8(len(legend.Pixels))
		legend.Pixels = append(legend.Pixels, LegendEntry{Type: DopplerApproaching, Color: Color{R: 0, G: 200, B: 200, A: 255}})
		legend.DopplerReceding = uint8(len(legend.Pixels))
		legend.Pixels = append(legend.Pixels, LegendEntry{Type: DopplerReceding, Color: Color{R: 0x90, G: 0xd0, B: 0xf0, A: 255}})
	}
	legend.HistoryStart = uint8(len(legend.Pixels))
	const START_DENSITY uint8 = 255
	const END_DENSITY uint8 = 63
	const DELTA_INTENSITY uint8 = (START_DENSITY - END_DENSITY) / BLOB_HISTORY_COLORS
	density := START_DENSITY
	for h := 0; h <= int(BLOB_HISTORY_COLORS); h++ {
		legend.Pixels = append(legend.Pixels, LegendEntry{Type: History, Color: Color{R: density, G: density, B: density, A: 255}})
		density -= DELTA_INTENSITY
	}

	return legend
}
