package graph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/template"

	"gopkg.in/gomail.v2"
)

// SendPasswordResetEmail sends a password reset email with the OTP code
func SendPasswordResetEmail(email string, code string) error {
	// Try Resend API first (for production/Render)
	resendAPIKey := os.Getenv("RESEND_API_KEY")
	if resendAPIKey != "" {
		err := sendViaResendAPI(email, code, resendAPIKey)
		if err == nil {
			return nil
		}
		// Log Resend error but fall back to SMTP
		fmt.Printf("Resend API failed, falling back to SMTP: %v\n", err)
	}

	// Get SMTP configuration from environment
	smtpHost := os.Getenv("EMAIL_HOST")
	smtpPort := os.Getenv("EMAIL_PORT")
	smtpUsername := os.Getenv("EMAIL_USERNAME")
	smtpPassword := os.Getenv("EMAIL_PASSWORD")

	if smtpHost == "" || smtpPort == "" || smtpUsername == "" || smtpPassword == "" {
		return fmt.Errorf("SMTP configuration is incomplete")
	}

	// Parse SMTP port
	port := 587
	if smtpPort != "" {
		_, err := fmt.Sscanf(smtpPort, "%d", &port)
		if err != nil {
			return fmt.Errorf("invalid SMTP port: %v", err)
		}
	}

	// Read and parse HTML template
	tmpl, err := template.ParseFiles("templates/password_reset.html")
	if err != nil {
		return fmt.Errorf("failed to parse email template: %v", err)
	}

	// Execute template with code
	var body bytes.Buffer
	data := struct {
		Code string
	}{
		Code: code,
	}
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute email template: %v", err)
	}

	// Create email message
	m := gomail.NewMessage()
	m.SetHeader("From", smtpUsername)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Password Reset Request")
	m.SetBody("text/html", body.String())

	// Send email
	d := gomail.NewDialer(smtpHost, port, smtpUsername, smtpPassword)
	d.TLSConfig = nil // Let gomail handle TLS automatically

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

// sendViaResendAPI sends email using Resend HTTP API
func sendViaResendAPI(email string, code string, apiKey string) error {
	// Read and parse HTML template
	tmpl, err := template.ParseFiles("templates/password_reset.html")
	if err != nil {
		return fmt.Errorf("failed to parse email template: %v", err)
	}

	// Execute template with code
	var body bytes.Buffer
	data := struct {
		Code string
	}{
		Code: code,
	}
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute email template: %v", err)
	}

	// Prepare request payload
	payload := map[string]interface{}{
		"from":    os.Getenv("RESEND_FROM_EMAIL"),
		"to":      []string{email},
		"subject": "Password Reset Request",
		"html":    body.String(),
	}

	if payload["from"] == "" {
		payload["from"] = "onboarding@resend.dev" // default Resend dev email
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	// Send HTTP request
	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Resend API returned status %d", resp.StatusCode)
	}

	return nil
}
