#!/bin/sh
tcpreplay -l 0 -i lo samples/navico_and_0183.pcap&
signalk-radar/radar-server&
signalk-server/bin/signalk-server