#!/bin/sh
tcpreplay -q -T select -l 0 -i lo halo_and_0183.pcap&
signalk-radar/radar-server --udp-source --type navico&
mayara/mayara -i lo -p 3002 --replay&
signalk-server/bin/signalk-server -c signalk/