package garminxhd

import (
	"fmt"
	"testing"

	"github.com/wdantuma/signalk-radar/radar-server/radar"
	"github.com/wdantuma/signalk-radar/radar-server/source/pcapsource"
)

func TestGarmin(t *testing.T) {

	var garmin radar.RadarSource

	source, _ := pcapsource.NewPcapSource("../../samples/garmin_xhd.pcap", false)

	reportFarmeSource := source.CreateFrameSource("garminReport", 50100)
	dataFrameSource := source.CreateFrameSource("garminData", 50102)

	garmin = NewGarminXhd(reportFarmeSource, dataFrameSource)

	source.Start()

	for m := range garmin.Source() {
		fmt.Printf("%d\n", m.Spoke.Angle)
	}
}
