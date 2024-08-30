#docker buildx build --platform linux/arm64,linux/amd64 --tag wdantuma/signalk-radar-demo:latest --push .
FROM node:20-alpine AS build
RUN apk add --no-cache git make musl-dev go python3 py3-pip g++ libpcap-dev gcc curl bash && \
    curl –proto ‘=https’ –tlsv1.2 -sSf https://sh.rustup.rs | bash -s -- -y 
ENV PATH="$PATH:/root/.cargo/bin"
WORKDIR /src
RUN git clone --branch v2.9.0 https://github.com/SignalK/signalk-server.git && \
    cd signalk-server && \
    npm i && sed -ri 's/ && npm run build:docs//' package.json && \
    npm run build:workspaces && \
    npm run build
RUN git clone --branch radar-support https://github.com/wdantuma/freeboard-sk.git && \
    cd freeboard-sk && \
    git checkout 2c4860976887ffbcd313211c9da6c6642f5b6f23 && \
    sed -ri 's/"declaration"/"skipLibCheck": true,\n"declaration"/' tsconfig.json && \
    npm i && \
    NG_CLI_ANALYTICS=ci npm run build
RUN git clone https://github.com/wdantuma/signalk-radar.git && \
    cd signalk-radar && \
    git checkout a0fe82bf76472cf4864aff743037de2c1da02fd6 && \
    cd signalk-radar-plugin && npm i && npm run build && \
    cd ../radar-server && \
    go build -o build/radar-server -ldflags="-X 'github.com/wdantuma/signalk-radar/radar-server/radarserver.Version=0.0.1'" ./cmd/signalk-radar
RUN git clone https://github.com/keesverruijt/mayara.git && \
    cd mayara && \
    git checkout 5952e073b6657995531a6f69f77b178a2a90703a && \
    . "$HOME/.cargo/env" && \
    cargo build --release
FROM node:20-alpine
RUN apk add tcpreplay curl unzip
WORKDIR /app
RUN mkdir signalk
RUN cd signalk && \
    mkdir charts && \
    cd charts && \
    curl -s http://dobs.nl/s57/charts.zip > charts.zip && \
    unzip -q charts.zip && \
    rm charts.zip
COPY signalk signalk
COPY samples/halo_and_0183.pcap .
COPY start.sh .
RUN chmod +x start.sh
COPY --from=build /src/signalk-server /app/signalk-server
COPY --from=build /src/freeboard-sk/public /app/signalk-server/node_modules/@signalk/freeboard-sk/public
COPY --from=build /src/signalk-radar/radar-server/build /app/signalk-radar
COPY --from=build /src/mayara/target/release /app/mayara
RUN cd signalk && \
    npm i @signalk/charts-plugin
COPY --from=build /src/signalk-radar/signalk-radar-plugin /app/signalk/node_modules/signalk-radar-plugin
CMD [ "/app/start.sh"]