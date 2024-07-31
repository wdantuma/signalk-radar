# Signal K radar

[WIP]

Beginning of a radar server as companion to Signal K and Freeboard-SK

[![Video](https://img.youtube.com/vi/5oTSLtVAKFs/0.jpg)](https://www.youtube.com/watch?v=5oTSLtVAKFs)

The radar server listens on a network source ( or pcap source ) and streams the converted data as protobuf records over a websocket

A webworker translates this stream to a image on a map