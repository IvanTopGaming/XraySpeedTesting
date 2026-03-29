# Xray Speed Testing

A self-contained benchmark for measuring Xray throughput using VLESS + REALITY + XTLS-Vision.

## How it works

Three Docker containers on an isolated bridge network:

- **xray-server** — Xray running as a VLESS/REALITY server
- **tester** — Xray client (SOCKS5) + Go program that sends data through the proxy and measures speed
- **sink** — TCP sink (`nc -lkp 8080 > /dev/null`) that discards all incoming data

The tester connects to the sink via the Xray proxy tunnel, sends a configurable amount of random data, and reports throughput.

## Usage

Pull the pre-built tester image and start all services:

```bash
docker compose pull && docker compose up
```

Results are printed to the `tester` container logs when the test finishes.

To follow the output in real time:

```bash
docker compose up && docker logs -f tester
```

## Configuration

Environment variables for the `tester` service in `docker-compose.yml`:

| Variable         | Default  | Description                        |
|------------------|----------|------------------------------------|
| `DATA_TO_SEND_MB` | `102400` | Total data to send, in megabytes  |
| `BUFFER_SIZE`     | `65536`  | Write buffer size, in bytes       |

## Important: use a separate machine for realistic results

This setup runs the Xray client and Xray server on the **same host**, sharing CPU resources. In production, client and server run on separate machines with dedicated CPUs.

Because VLESS/REALITY/XTLS-Vision is CPU-intensive on both sides, co-locating them causes CPU contention and **underestimates** real-world throughput.

For accurate absolute numbers, run `xray-server` (and `sink`) on a remote machine and point the tester's `TARGET_HOST` / `xray-server` address at it. The current setup is still useful for **comparing configurations** against each other, since the bias affects all runs equally.

## Project structure

```
.
├── docker-compose.yml
├── xray-server/
│   └── config.json       # Xray server config (VLESS inbound, REALITY)
└── tester/
    ├── Dockerfile         # Builds Go binary + downloads Xray client
    ├── client-config.json # Xray client config (SOCKS5 inbound, VLESS outbound)
    └── speedtest.go       # Go speed test program
```
