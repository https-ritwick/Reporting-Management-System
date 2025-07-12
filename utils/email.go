package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendHTMLEmail(to string, subject string, htmlBody string) error {
	from := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASS")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")

	auth := smtp.PlainAuth("", from, password, host)

	headers := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n"
	headers += fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n\n", from, to, subject)

	message := headers + htmlBody

	return smtp.SendMail(host+":"+port, auth, from, []string{to}, []byte(message))
}
