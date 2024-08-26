#!/bin/sh
tcpreplay -q -T select -l 0 -i lo navico_and_0183.pcap&
signalk-radar/radar-server --udp-source --type navico&
mayara/mayara&
signalk-server/bin/signalk-server -c signalk/