package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Device represents the structure of a single device object in the "data" array
type Device struct {
	ID        string `json:"_id"`
	DeviceID  string `json:"device_id"`
	PosId     string `json:"pos_id"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	Version   int    `json:"__v"`
}

// InvoiceRequest adalah struktur untuk payload permintaan
type InvoiceRequest struct {
	DeviceID   string `json:"device_id"`
	Timezone   string `json:"timezone"`
	CardNumber string `json:"card_number"`
	PlateNo    string `json:"plate_no,omitempty"` // "omitempty" jika tidak diperlukan
}

// DeviceAPIResponse represents the full response from the API
type DeviceAPIResponse struct {
	Data   []Device `json:"data"`
	Status int      `json:"status"`
}

// InvoiceData represents the "data" object in the response
type InvoiceData struct {
	ID              string     `json:"id"`
	MemberID        *string    `json:"member_id"`       // Nullable field
	InCardNumber    *string    `json:"in_card_number"`  // Nullable field
	OutCardNumber   *string    `json:"out_card_number"` // Nullable field
	PlateNo         string     `json:"plate_no"`
	Amount          int        `json:"amount"`
	CheckedIn       time.Time  `json:"checked_in"`
	CheckedOut      *time.Time `json:"checked_out"` // Nullable field
	DurationMinutes int        `json:"duration_minutes"`
}

// InvoiceResponse represents the full response structure
type InvoiceResponse struct {
	Data InvoiceData `json:"data"`
}

// CheckDeviceAPI sends a GET request to the API and parses the response
func CheckDeviceAPI(apiURL string) (*DeviceAPIResponse, error) {
	client := &http.Client{
		Timeout: 10 * time.Second, // Set timeout untuk request
	}

	// Buat request ke API
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make GET request: %w", err)
	}
	defer resp.Body.Close()

	// Periksa status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Decode response JSON
	var deviceResp DeviceAPIResponse
	err = json.NewDecoder(resp.Body).Decode(&deviceResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %w", err)
	}

	return &deviceResp, nil
}

// CreateInvoice sends a request to create an invoice
func CreateInvoice(schema, host, apikey, postID, deviceID, timezone, cardNumber, plateNo string) (*InvoiceResponse, error) {
	url := fmt.Sprintf("%s://%s/posts/%s/invoices", schema, host, postID)
	fmt.Printf("URL: %s\n", url)

	// Debug: Tampilkan detail payload sebelum pengiriman
	requestPayload := InvoiceRequest{
		DeviceID:   deviceID,
		Timezone:   timezone,
		CardNumber: cardNumber,
		PlateNo:    plateNo,
	}
	if plateNo != "" {
		requestPayload.PlateNo = plateNo
	}

	// Debug: Cetak payload sebelum encoding
	fmt.Printf("Request Payload: %+v\n", requestPayload)

	// Encode payload ke JSON
	payloadBytes, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	// Debug: Tampilkan JSON payload yang akan dikirim
	fmt.Printf("JSON Payload: %s\n", string(payloadBytes))

	// Buat permintaan HTTP
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Tambahkan header
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", apikey)

	// Debug: Tampilkan header permintaan
	fmt.Printf("Headers: %v\n", req.Header)

	// Debug: Lakukan pemeriksaan akhir pada seluruh permintaan
	fmt.Printf("HTTP Request: Method=%s, URL=%s, Body=%s\n", req.Method, req.URL, string(payloadBytes))

	// Kirim permintaan
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var invoiceResponse InvoiceResponse
	err = json.NewDecoder(resp.Body).Decode(&invoiceResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &invoiceResponse, nil
}
