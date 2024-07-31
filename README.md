# Signal K radar

[WIP]

Beginning of a radar server as companion to Signal K and Freeboard-SK

<video loop src="doc/img/signalk-radar.webm" width="320" height="240" controls></video>

The radar server listens on a network source ( or pcap source ) and streams the converted data as protobuf records over a websocket

A webworker translates this stream to a image on a map