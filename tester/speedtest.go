package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"

	"golang.org/x/net/proxy"
)

var (
	socksHost       = getEnv("SOCKS_HOST", "127.0.0.1")
	socksPort       = getEnv("SOCKS_PORT", "1080")
	targetHost      = getEnv("TARGET_HOST", "sink")
	targetPort      = getEnv("TARGET_PORT", "8080")
	dataToSendMB, _ = strconv.Atoi(getEnv("DATA_TO_SEND_MB", "100"))
	bufferSize, _   = strconv.Atoi(getEnv("BUFFER_SIZE", "8192"))
	dialTimeout, _  = time.ParseDuration(getEnv("DIAL_TIMEOUT", "60s"))
)

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func main() {
	time.Sleep(2 * time.Second)
	runSpeedTest()
}

// transfer writes totalBytes to conn and reports how many bytes went out and how
// long it took.
func transfer(conn net.Conn, randomData []byte, totalBytes int64) (int64, time.Duration, error) {
	startTime := time.Now()
	bytesSent := int64(0)
	chunks := totalBytes / int64(len(randomData))
	remainder := totalBytes % int64(len(randomData))

	for i := int64(0); i < chunks; i++ {
		n, err := conn.Write(randomData)
		bytesSent += int64(n)
		if err != nil {
			return bytesSent, time.Since(startTime), err
		}
	}

	if remainder > 0 {
		n, err := conn.Write(randomData[:remainder])
		bytesSent += int64(n)
		if err != nil {
			return bytesSent, time.Since(startTime), err
		}
	}

	return bytesSent, time.Since(startTime), nil
}

func runSpeedTest() {
	totalBytes := int64(dataToSendMB) * 1024 * 1024

	fmt.Println("--- Preparing test ---")
	fmt.Printf("Proxy: %s:%s\n", socksHost, socksPort)
	fmt.Printf("Target: %s:%s\n", targetHost, targetPort)
	fmt.Printf("Data to send: %d MB\n", dataToSendMB)
	fmt.Printf("Buffer size: %d KB\n", bufferSize/1024)
	fmt.Println("----------------------")

	dialer, err := proxy.SOCKS5("tcp", net.JoinHostPort(socksHost, socksPort), nil, proxy.Direct)
	if err != nil {
		log.Fatalf("Failed to create SOCKS5 dialer: %v", err)
	}

	randomData := make([]byte, bufferSize)
	if _, err := rand.Read(randomData); err != nil {
		log.Fatalf("Failed to generate payload: %v", err)
	}

	targetAddr := net.JoinHostPort(targetHost, targetPort)

	// Xray answers the SOCKS5 CONNECT lazily, before the upstream connection
	// exists, so neither the dial nor the first writes prove the tunnel works: a
	// broken upstream only shows up as a reset once the socket buffers fill. The
	// whole transfer is therefore retried, and a partial run is discarded rather
	// than reported, since its throughput number would be meaningless.
	deadline := time.Now().Add(dialTimeout)
	for attempt := 1; ; attempt++ {
		conn, err := dialer.Dial("tcp", targetAddr)
		if err == nil {
			fmt.Printf("\nConnection established (attempt %d). Sending data...\n", attempt)
			bytesSent, duration, err := transfer(conn, randomData, totalBytes)
			conn.Close()
			if err == nil {
				report(bytesSent, duration)
				return
			}
			fmt.Printf("Transfer failed after %.2f MB (%v)\n", float64(bytesSent)/(1024*1024), err)
		} else {
			fmt.Printf("Connect failed (%v)\n", err)
		}

		if !time.Now().Before(deadline) {
			log.Fatalf("Giving up after %d attempts in %s", attempt, dialTimeout)
		}
		time.Sleep(2 * time.Second)
	}
}

func report(bytesSent int64, duration time.Duration) {
	mbSent := float64(bytesSent) / (1024 * 1024)
	durationSec := duration.Seconds()
	speedMbps := (float64(bytesSent) * 8) / (durationSec * 1024 * 1024)

	fmt.Println("\n\n--- Final results ---")
	fmt.Printf("Sent: %.2f MB\n", mbSent)
	fmt.Printf("Duration: %.2f seconds\n", durationSec)
	fmt.Printf("Average speed: %.2f Mbit/s\n", speedMbps)
	fmt.Println("---------------------")
}
