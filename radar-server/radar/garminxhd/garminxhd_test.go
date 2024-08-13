package garminxhd

import (
	"fmt"
	"testing"

	"github.com/wdantuma/signalk-radar/radar-server/radar"
	"github.com/wdantuma/signalk-radar/radar-server/source/pcapsource"
)

func TestGarmin(t *testing.T) {

	var garmin radar.RadarSource

	source, _ := pcapsource.NewPcapSource("../../../samples/garmin_xhd.pcap", false)

	garmin = NewGarminXhd(source)

	source.Start()

	for m := range garmin.Source() {
		fmt.Printf("%d\n", m.Spoke.Angle)
	}
}
