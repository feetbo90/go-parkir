package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go_parkir/models"
	"go_parkir/utils"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tarm/serial"
)

var (
	mu           sync.Mutex
	cameraAPIURL = "https://1017-175-158-36-199.ngrok-free.app/api/Photo" // URL konfigurasi kamera
	httpClient   = &http.Client{}                                         // Gunakan satu instance HTTP client
)

func init() {
	// Configure and open serial port
	config := &serial.Config{Name: "/dev/cu.usbmodem141301", Baud: 9600}
	s, err := serial.OpenPort(config)
	if err != nil {
		log.Println("Error opening serial port: ", err)
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

	if len(models.DataModel) > 0 {
		// Ambil elemen pertama dan konversi ke integer
		value, err := strconv.Atoi(models.DataModel[0])
		if err != nil {
			fmt.Println("Error converting DataModel[0] to integer:", err)
			return
		}

		// Bandingkan nilai integer dengan 2
		if value == 2 {
			// Panggil API aktivasi kamera secara sinkron
			if err := ActivateCameraAPI(); err != nil {
				log.Printf("Failed to activate camera: %v", err)
				return
			}
			log.Println("Camera activated successfully")
		} else if value == 1 {
			fmt.Println("models.DataModel[0] == 2 is false")
			apiURL := "http://localhost:3000/api/device"

			// Panggil fungsi untuk cek API
			deviceResp, err := utils.CheckDeviceAPI(apiURL)
			if err != nil {
				log.Println("Error checking device API: ", err)
			} else {
				// Tampilkan hasil
				fmt.Printf("Status: %d\n", deviceResp.Status)
				posResp, errPlat := utils.CheckPos("http://localhost:3000/api/get_plat")
				if errPlat != nil {
					log.Println("Error checking plat API: ", errPlat)
				} else {
					// pos := posResp.Data[0].plat
					fmt.Printf("plat response: %s\n", posResp.Plat)

					for _, device := range deviceResp.Data {
						fmt.Printf("Device ID: %s, Created At: %s, Updated At: %s\n", device.DeviceID, device.CreatedAt, device.UpdatedAt)
						// Timezone Asia/Jakarta
						location, err := time.LoadLocation("Asia/Jakarta")
						if err != nil {
							fmt.Println("Error:", err)
							return
						}

						// Mendapatkan waktu saat ini dalam timezone tersebut
						now := time.Now().In(location)
						fmt.Println("Current time in Asia/Jakarta:", now)

						schema := "https"
						host := "api.logicparking.id"
						apikey := "S92AWBxpvxmbY320mf7o7nCFe5OwQhaJ"
						postID := device.PosId
						deviceID := device.DeviceID
						timezone := "Asia/Jakarta"
						plateNo := posResp.Plat

						response, err := utils.CreateInvoice(schema, host, apikey, postID, deviceID, timezone, "", plateNo)
						if err != nil {
							log.Println("Error creating invoice: ", err)
						} else {
							// Output hasil unmarshal
							fmt.Printf("Invoice ID: %s\n", response.Data.ID)
							fmt.Printf("Checked In: %s\n", response.Data.CheckedIn.Format(time.RFC3339))
							fmt.Printf("Plate No: %s\n", response.Data.PlateNo)

							// URL untuk endpoint Express
							printURL := "http://192.168.100.53:3001/api/print"

							// Payload untuk dikirim ke server Node.js
							printPayload := map[string]interface{}{
								"data": response.Data, // Pastikan struct `Data` sesuai dengan kebutuhan
							}

							// Encode payload ke JSON
							payloadBytes, err := json.Marshal(printPayload)
							if err != nil {
								log.Println("Failed to marshal print payload: ", err)
							}

							// Buat request HTTP POST
							req, err := http.NewRequest("POST", printURL, bytes.NewBuffer(payloadBytes))
							if err != nil {
								log.Println("Failed to create request for printing: ", err)
							}

							// Tambahkan header
							req.Header.Set("Content-Type", "application/json")

							client := &http.Client{Timeout: 10 * time.Second}
							resp, err := client.Do(req)
							if err != nil {
								log.Println("Failed to send print request: ", err)
							} else {
								// Periksa respons dari server Node.js
								if resp.StatusCode != http.StatusOK {
									log.Println("Print request failed with status: ", resp.StatusCode)
								}
								defer resp.Body.Close()
								log.Println("Print request sent successfully!")
							}

						}
					}
				}
			}
		}
	} else {
		fmt.Println("DataModel is empty")
	}

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
	req, err := http.NewRequest("GET", cameraAPIURL, bytes.NewBuffer([]byte{}))
	if err != nil {
		log.Printf("Failed to create request: %v\n", err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	log.Println("Sending request to camera API...") // Tambahkan log di sini
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("Failed to call : %v\n", err)
		return fmt.Errorf("failed to call : %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Camera activation failed, HTTP status code: %d", resp.StatusCode)
		return fmt.Errorf("camera activation failed with HTTP status code: %d", resp.StatusCode)
	}

	var cameraResp CameraResponse
	err = json.NewDecoder(resp.Body).Decode(&cameraResp)
	if err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Periksa status respons
	if cameraResp.Status != 200 {
		return fmt.Errorf("camera activation failed: status %d, message: %s", cameraResp.Status, cameraResp.Message)
	}

	// Log pesan sukses
	log.Printf("Camera activated successfully: %s", cameraResp.Message)
	return nil
}
