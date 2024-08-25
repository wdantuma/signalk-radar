VERSION=0.0.1
BINARY_NAME=signalk-radar

build: radar-server radar-client signalk-radar-plugin

radar-server/build/${BINARY_NAME}: 
	GOARCH=amd64 GOOS=linux cd radar-server;go build -o build/${BINARY_NAME} -ldflags="-X 'github.com/wdantuma/signalk-radar/radar-server/radarserver.Version=${VERSION}'" ./cmd/signalk-radar

signalk-radar-plugin/dist/index.js: 
	cd signalk-radar-plugin;tsc	

radar-server: radar-server/build/${BINARY_NAME}

signalk-radar-plugin: signalk-radar-plugin/dist/index.js

buildarm:
	GOARCH=arm GOOS=linux go build -o build/${BINARY_NAME}-arm -ldflags="-X 'github.com/wdantuma/signalk-radar/radar-server/radarserver.Version=${VERSION}'" ./cmd/signalk-radar

run: build
	./radar-server/build/${BINARY_NAME} --port 3001 --file-source  samples/garmin_xhd.pcap  --type garminxhd

debug: build
	./radar-server/build/${BINARY_NAME} --port 3001 --debug --file-source  samples/garmin_xhd.pcap  --type garminxhd

clean:
	cd radar-server;go clean
	rm -rf radar-server/build/*
	rm -rf signalk-radar-plugin/dist/*
