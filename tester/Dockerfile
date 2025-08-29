FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN go mod init speedtest && \
    go get golang.org/x/net/proxy@latest

COPY speedtest.go .
RUN go build -o speedtest speedtest.go

FROM debian:12-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    curl unzip ca-certificates && \
    rm -rf /var/lib/apt/lists/*

ENV XRAY_VERSION 1.8.8
ENV XRAY_ARCH linux-64

RUN curl -L -o xray.zip "https://github.com/XTLS/Xray-core/releases/download/v${XRAY_VERSION}/Xray-${XRAY_ARCH}.zip" && \
    unzip xray.zip xray -d /usr/local/bin/ && \
    rm xray.zip && \
    chmod +x /usr/local/bin/xray

COPY --from=builder /app/speedtest /app/speedtest

RUN mkdir -p /usr/local/etc/xray
COPY client-config.json /usr/local/etc/xray/config.json

CMD ["sh", "-c", "/usr/local/bin/xray run -c /usr/local/etc/xray/config.json & sleep 2 && /app/speedtest"]