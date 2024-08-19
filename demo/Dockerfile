FROM node:20-alpine AS build
RUN apk add --no-cache git make musl-dev go python3 py3-pip g++ libpcap-dev
WORKDIR /src
RUN git clone --branch v2.8.3 https://github.com/SignalK/signalk-server.git
RUN git clone https://github.com/wdantuma/signalk-radar.git
RUN git clone --branch radar-support https://github.com/wdantuma/freeboard-sk.git
WORKDIR /src/signalk-server
RUN npm i
RUN sed -ri 's/ && npm run build:docs//' package.json 
RUN npm run build
WORKDIR /src/freeboard-sk
RUN sed -ri 's/"declaration"/"skipLibCheck": true,\n"declaration"/' tsconfig.json
RUN npm i
RUN NG_CLI_ANALYTICS=ci npm run build
WORKDIR /src/signalk-radar/signalk-radar-plugin
RUN npm i
RUN npm run build
WORKDIR /src/signalk-radar/radar-server
RUN GOARCH=amd64 GOOS=linux go build -o build/radar-server -ldflags="-X 'github.com/wdantuma/signalk-radar/radar-server/radarserver.Version=0.0.1'" ./cmd/signalk-radar
FROM node:20-alpine
RUN apk add tcpreplay
WORKDIR /app
COPY samples .
COPY --from=build /src/signalk-server /app/signalk-server
COPY --from=build /src/freeboard-sk/public /app/signalk-server/node_modules/@signalk/freeboard-sk/public
COPY --from=build /src/signalk-radar/radar-server/build /app/signalk-radar
CMD [ "/app/start.sh"]