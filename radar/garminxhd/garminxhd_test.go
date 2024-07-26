package garminxhd

import (
	"testing"

	"github.com/wdantuma/signalk-radar/radar"
	"github.com/wdantuma/signalk-radar/source/pcapsource"
)

func TestGarmin(t *testing.T) {

	var garmin radar.RadarSource

	source := pcapsource.NewPcapSource("../../examples/garmin_xhd.pcap", false)

	reportFarmeSource := source.CreateFrameSource("garminReport", 50100)
	dataFrameSource := source.CreateFrameSource("garminData", 50102)

	garmin = NewGarminXhd(reportFarmeSource, dataFrameSource)

	source.Start()

	<-garmin.Source()
}
