package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go_parkir/models"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/tarm/serial"
)

var (
	mu           sync.Mutex
	cameraAPIURL = "http://localhost:8081/camera/activate" // URL konfigurasi kamera
	httpClient   = &http.Client{}                          // Gunakan satu instance HTTP client
)

func init() {
	// Configure and open serial port
	config := &serial.Config{Name: "/dev/cu.usbmodem141201", Baud: 9600}
	s, err := serial.OpenPort(config)
	if err != nil {
		log.Fatalf("Error opening serial port: %v", err)
	}

	// Start reading serial data in a separate goroutine
	go func() {
		defer s.Close()
		for {
			data, err := readSerialData(s)
			if err != nil {
				fmt.Println("Error reading from serial port:", err)
				time.Sleep(1 * time.Second)
				continue
			}
			processData(data)
		}
	}()
}

// readSerialData reads data from the serial port
func readSerialData(s *serial.Port) (string, error) {
	buf := make([]byte, 123)
	n, err := s.Read(buf)
	if err != nil {
		return "", fmt.Errorf("error reading from serial port: %v", err)
	}
	return string(buf[:n]), nil
}

// processData splits the data and stores it in the DataModel
func processData(data string) {
	parts := strings.Split(data, ",")
	mu.Lock()
	models.DataModel = parts
	mu.Unlock()
	fmt.Println("Data processed:", models.DataModel)

	// Panggil API aktivasi kamera secara sinkron
	if err := ActivateCameraAPI(); err != nil {
		log.Printf("Failed to activate camera: %v", err)
		return
	}
	log.Println("Camera activated successfully")
}

// GetLatestData returns the latest data stored in DataModel
func GetLatestData() []string {
	mu.Lock()
	defer mu.Unlock()
	return models.DataModel
}

type CameraResponse struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func ActivateCameraAPI() error {
	log.Println("Attempting to activate camera...") // Tambahkan log di sini
	req, err := http.NewRequest("POST", cameraAPIURL, bytes.NewBuffer([]byte{}))
	if err != nil {
		log.Printf("Failed to create request: %v\n", err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	log.Println("Sending request to camera API...") // Tambahkan log di sini
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("Failed to call /camera/activate: %v\n", err)
		return fmt.Errorf("failed to call /camera/activate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Camera activation failed, HTTP status code: %d\n", resp.StatusCode)
		return fmt.Errorf("camera activation failed with HTTP status code: %d", resp.StatusCode)
	}

	var cameraResp CameraResponse
	err = json.NewDecoder(resp.Body).Decode(&cameraResp)
	if err != nil {
		log.Printf("Failed to decode response: %v\n", err)
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if cameraResp.Status == 200 {
		log.Printf("Camera activated successfully: %s\n", cameraResp.Message)
	} else {
		log.Printf("Camera activation failed: %s\n", cameraResp.Message)
		return fmt.Errorf("camera activation failed: %s", cameraResp.Message)
	}

	return nil
}
