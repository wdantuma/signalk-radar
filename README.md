# Signal K radar

[Work in progress]

Beginning of a radar server as companion to Signal K and Freeboard-SK

![Screenshot](./doc/img/screenshot-1.png)

# Code
This repository contains 3 projects

## radar-server
A radar server implemented in golang and implementing the Json REST [API](signalk-radar-plugin/openApi.json) and the protocal buffers websocket [API](radar-server/radar/schema/RadarMessage.proto)
## radar-client
An simple Angular/typescript client connecting directly to the radar server
## signalk-radar-plugin
An Signal K plugin exposing (proxying) the radar Json REST API in signal K

# Freeboard SK support
In [this](https://github.com/wdantuma/freeboard-sk/tree/radar-support) Freeboard SK branch there is a beginning of using all the above to provide radar support