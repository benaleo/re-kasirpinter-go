package graph

import (
	"fmt"
	"os"

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

	// Create email message
	m := gomail.NewMessage()
	m.SetHeader("From", smtpUsername)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Password Reset Request")

	// HTML email template based on the provided design
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Password Reset Request</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            background-color: #f4f4f4;
            margin: 0;
            padding: 20px;
        }
        .container {
            max-width: 600px;
            margin: 0 auto;
            background-color: #ffffff;
            padding: 40px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .header {
            text-align: center;
            margin-bottom: 30px;
        }
        .header h1 {
            color: #333333;
            margin: 0;
            font-size: 24px;
        }
        .code-container {
            background-color: #f8f9fa;
            border: 2px dashed #dee2e6;
            padding: 20px;
            text-align: center;
            margin: 30px 0;
            border-radius: 8px;
        }
        .code {
            font-size: 36px;
            font-weight: bold;
            letter-spacing: 8px;
            color: #007bff;
            margin: 0;
        }
        .expiry {
            color: #6c757d;
            font-size: 14px;
            margin-top: 10px;
        }
        .security-notice {
            background-color: #fff3cd;
            border-left: 4px solid #ffc107;
            padding: 15px;
            margin: 30px 0;
            border-radius: 4px;
        }
        .security-notice p {
            margin: 0;
            color: #856404;
            font-size: 14px;
        }
        .footer {
            margin-top: 40px;
            padding-top: 20px;
            border-top: 1px solid #dee2e6;
            text-align: center;
            font-size: 12px;
            color: #6c757d;
        }
        .footer-links {
            margin: 10px 0;
        }
        .footer-links a {
            color: #007bff;
            text-decoration: none;
            margin: 0 10px;
        }
        .footer-links a:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Password Reset Request</h1>
        </div>
        
        <p>Hello,</p>
        
        <p>We received a request to reset your password. Use the verification code below to proceed with the password reset:</p>
        
        <div class="code-container">
            <p class="code">%s</p>
            <p class="expiry">This code will expire in 10 minutes</p>
        </div>
        
        <div class="security-notice">
            <p><strong>Security Notice:</strong> If you did not request this password reset, please ignore this email. Your account remains secure.</p>
        </div>
        
        <p>Thank you,<br>Kasir Pintar Team</p>
        
        <div class="footer">
            <div class="footer-links">
                <a href="#">Privacy Policy</a>
                <a href="#">Support</a>
                <a href="#">Security Center</a>
            </div>
            <p>&copy; 2024 Kasir Pintar. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, code)

	m.SetBody("text/html", htmlBody)

	// Send email
	d := gomail.NewDialer(smtpHost, port, smtpUsername, smtpPassword)
	d.TLSConfig = nil // Let gomail handle TLS automatically

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}
