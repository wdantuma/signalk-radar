package navico

import (
	"fmt"
	"testing"

	"github.com/wdantuma/signalk-radar/radar-server/radar"
	"github.com/wdantuma/signalk-radar/radar-server/source/pcapsource"
)

func TestNavico(t *testing.T) {

	var garmin radar.RadarSource

	source, _ := pcapsource.NewPcapSource("../../../samples/navico_and_0183.pcap", false)

	garmin = NewNavico(source)

	source.Start()

	for m := range garmin.Source() {
		for _, s := range m.Spokes {
			fmt.Printf("Angle %d\n", s.Angle)
		}
	}
}
