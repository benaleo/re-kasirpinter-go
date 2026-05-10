package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type MidtransService struct {
	ServerKey  string
	MerchantID string
	BaseURL    string
}

type MidtransQrisRequest struct {
	PaymentType       string                    `json:"payment_type"`
	TransactionDetails MidtransTransactionDetails `json:"transaction_details"`
	CustomExpiry      *MidtransCustomExpiry      `json:"custom_expiry,omitempty"`
}

type MidtransTransactionDetails struct {
	OrderID     string  `json:"order_id"`
	GrossAmount float64 `json:"gross_amount"`
}

type MidtransCustomExpiry struct {
	ExpiryDuration int    `json:"expiry_duration"`
	Unit           string `json:"unit"`
}

type MidtransQrisResponse struct {
	StatusCode       string              `json:"status_code"`
	StatusMessage    string              `json:"status_message"`
	TransactionID    string              `json:"transaction_id"`
	OrderID          string              `json:"order_id"`
	MerchantID       string              `json:"merchant_id"`
	GrossAmount      string              `json:"gross_amount"`
	Currency         string              `json:"currency"`
	PaymentType      string              `json:"payment_type"`
	TransactionTime  string              `json:"transaction_time"`
	TransactionStatus string             `json:"transaction_status"`
	FraudStatus      string              `json:"fraud_status"`
	Actions          []MidtransAction    `json:"actions"`
	QrString         string              `json:"qr_string"`
	Acquirer         string              `json:"acquirer"`
}

type MidtransAction struct {
	Name   string `json:"name"`
	Method string `json:"method"`
	URL    string `json:"url"`
}

func NewMidtransService() *MidtransService {
	baseURL := "https://api.midtrans.com"
	if os.Getenv("ENV") == "production" {
		baseURL = "https://api.midtrans.com"
	} else {
		baseURL = "https://api.sandbox.midtrans.com"
	}

	return &MidtransService{
		ServerKey:  os.Getenv("MERCHANT_SERVER_KEY"),
		MerchantID: os.Getenv("MERCHANT_ID"),
		BaseURL:    baseURL,
	}
}

// CreateQrisTransaction creates a QRIS transaction using Midtrans MPM API
func (s *MidtransService) CreateQrisTransaction(orderID string, grossAmount float64, expiryMinutes int) (*MidtransQrisResponse, error) {
	url := fmt.Sprintf("%s/v2/charge", s.BaseURL)

	request := MidtransQrisRequest{
		PaymentType: "qris",
		TransactionDetails: MidtransTransactionDetails{
			OrderID:     orderID,
			GrossAmount: grossAmount,
		},
	}

	// Add custom expiry if specified
	if expiryMinutes > 0 {
		request.CustomExpiry = &MidtransCustomExpiry{
			ExpiryDuration: expiryMinutes,
			Unit:           "minutes",
		}
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	authString := base64.StdEncoding.EncodeToString([]byte(s.ServerKey + ":"))
	req.Header.Set("Authorization", "Basic "+authString)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Midtrans API error: %s - %s", resp.Status, string(body))
	}

	var response MidtransQrisResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// GetQrCodeURL returns the URL to generate QR code image
func (s *MidtransService) GetQrCodeURL(transactionID string) string {
	return fmt.Sprintf("%s/v2/qris/%s/qr-code", s.BaseURL, transactionID)
}
