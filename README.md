# Signal K radar

[Work in progress]

Beginning of a radar server as companion to Signal K and Freeboard-SK

![Screenshot](./doc/img/screenshot-1.png)

# Code
This repository contains 2 projects

## radar-server
A radar server implemented in golang and implementing the Json REST [API](signalk-radar-plugin/openApi.json) and the protocal buffers websocket [API](radar-server/radar/schema/RadarMessage.proto)
## signalk-radar-plugin
An Signal K plugin exposing (proxying) the radar Json REST API in signal K

# Freeboard SK support
In [this](https://github.com/wdantuma/freeboard-sk/tree/radar-support) Freeboard SK branch has support for using the repo's above