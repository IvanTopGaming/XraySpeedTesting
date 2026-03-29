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
)

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func main() {
	time.Sleep(5 * time.Second)
	runSpeedTest()
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

	conn, err := dialer.Dial("tcp", net.JoinHostPort(targetHost, targetPort))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	fmt.Println("Connection established. Sending data...")

	randomData := make([]byte, bufferSize)
	rand.New(rand.NewSource(time.Now().UnixNano())).Read(randomData)

	startTime := time.Now()
	bytesSent := int64(0)
	chunks := totalBytes / int64(bufferSize)
	remainder := totalBytes % int64(bufferSize)

	for i := int64(0); i < chunks; i++ {
		n, err := conn.Write(randomData)
		if err != nil {
			log.Fatalf("Failed to send data: %v", err)
		}
		bytesSent += int64(n)
	}

	if remainder > 0 {
		n, err := conn.Write(randomData[:remainder])
		if err != nil {
			log.Fatalf("Failed to send data: %v", err)
		}
		bytesSent += int64(n)
	}

	duration := time.Since(startTime)
	conn.Close()

	mbSent := float64(bytesSent) / (1024 * 1024)
	durationSec := duration.Seconds()
	speedMbps := (float64(bytesSent) * 8) / (durationSec * 1024 * 1024)

	fmt.Println("\n\n--- Final results ---")
	fmt.Printf("Sent: %.2f MB\n", mbSent)
	fmt.Printf("Duration: %.2f seconds\n", durationSec)
	fmt.Printf("Average speed: %.2f Mbit/s\n", speedMbps)
	fmt.Println("---------------------")
}