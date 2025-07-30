package utils

import (
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

var (
	SMTP_HOST = "smtp.gmail.com"
	SMTP_PORT = "587"
	SMTP_USER = "ritwickbackup10@gmail.com"
	SMTP_PASS = "xtug cgkb lgbo fihn"
)

// SendHTMLEmail sends a styled HTML email to the specified recipient.
func SendHTMLEmail(to string, subject string, htmlBody string) error {
	log.Println("📧 Preparing to send HTML email")

	from := SMTP_USER
	password := SMTP_PASS
	host := SMTP_HOST
	port := SMTP_PORT

	// Step 1: Setup Auth
	log.Println("🔐 Setting up SMTP authentication")
	auth := smtp.PlainAuth("", from, password, host)

	// Step 2: Construct email headers
	log.Println("🧩 Constructing message headers and body")
	var msg strings.Builder
	msg.WriteString("MIME-Version: 1.0;\n")
	msg.WriteString("Content-Type: text/html; charset=\"UTF-8\";\n")
	msg.WriteString(fmt.Sprintf("From: %s\n", from))
	msg.WriteString(fmt.Sprintf("To: %s\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\n\n", subject))
	msg.WriteString(htmlBody)

	// Step 3: Send email
	log.Println("🚀 Attempting to send email...")
	err := smtp.SendMail(host+":"+port, auth, from, []string{to}, []byte(msg.String()))
	if err != nil {
		log.Println("❌ Failed to send email:", err)
		return err
	}

	log.Println("✅ Email sent successfully to:", to)
	return nil
}
