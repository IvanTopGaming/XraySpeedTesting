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

// Настройки теста
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

	fmt.Println("--- Подготовка к тесту ---")
	fmt.Printf("Прокси: %s:%s\n", socksHost, socksPort)
	fmt.Printf("Цель: %s:%s\n", targetHost, targetPort)
	fmt.Printf("Объем данных для отправки: %d MB\n", dataToSendMB)
	fmt.Printf("Размер буфера: %d KB\n", bufferSize/1024)
	fmt.Println("--------------------------")

	dialer, err := proxy.SOCKS5("tcp", net.JoinHostPort(socksHost, socksPort), nil, proxy.Direct)
	if err != nil {
		log.Fatalf("Ошибка создания SOCKS5 диалектора: %v", err)
	}

	conn, err := dialer.Dial("tcp", net.JoinHostPort(targetHost, targetPort))
	if err != nil {
		log.Fatalf("Ошибка соединения: %v", err)
	}
	defer conn.Close()

	fmt.Println("Соединение установлено. Начинаю отправку данных...")

	randomData := make([]byte, bufferSize)
	rand.New(rand.NewSource(time.Now().UnixNano())).Read(randomData)

	startTime := time.Now()
	bytesSent := int64(0)
	chunks := totalBytes / int64(bufferSize)
	remainder := totalBytes % int64(bufferSize)

	for i := int64(0); i < chunks; i++ {
		n, err := conn.Write(randomData)
		if err != nil {
			log.Fatalf("Ошибка отправки данных: %v", err)
		}
		bytesSent += int64(n)
	}

	if remainder > 0 {
		n, err := conn.Write(randomData[:remainder])
		if err != nil {
			log.Fatalf("Ошибка отправки данных: %v", err)
		}
		bytesSent += int64(n)
	}

	duration := time.Since(startTime)
	conn.Close()

	mbSent := float64(bytesSent) / (1024 * 1024)
	durationSec := duration.Seconds()
	speedMbps := (float64(bytesSent) * 8) / (durationSec * 1024 * 1024)

	fmt.Println("\n\n--- Финальные результаты ---")
	fmt.Printf("Отправлено: %.2f MB\n", mbSent)
	fmt.Printf("Затраченное время: %.2f секунд\n", durationSec)
	fmt.Printf("Средняя скорость: %.2f Мбит/с\n", speedMbps)
	fmt.Println("---------------------------")
}