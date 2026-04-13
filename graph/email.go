package graph

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"gopkg.in/gomail.v2"
)

// SendPasswordResetEmail sends a password reset email with the OTP code
func SendPasswordResetEmail(email string, code string) error {
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
