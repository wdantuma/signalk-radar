VERSION=0.0.1
BINARY_NAME=signalk-radar


build/${BINARY_NAME}: 
	GOARCH=amd64 GOOS=linux go build -o radar-server/build/${BINARY_NAME} -ldflags="-X 'github.com/wdantuma/signalk-radar/radar-server/radarserver.Version=${VERSION}'" ./radar-server/cmd/signalk-radar


build: build/${BINARY_NAME}

buildarm:
	GOARCH=arm GOOS=linux go build -o build/${BINARY_NAME}-arm -ldflags="-X 'github.com/wdantuma/signalk-radar/radar-server/radarserver.Version=${VERSION}'" ./cmd/signalk-radar

run: build
	./build/${BINARY_NAME} --port 3001 --file-source  samples/garmin_xhd.pcap  --type garminxhd

debug: build
	./build/${BINARY_NAME} --port 3001 --debug --file-source  samples/garmin_xhd.pcap  --type garminxhd

clean:
	go clean
	rm build/*
