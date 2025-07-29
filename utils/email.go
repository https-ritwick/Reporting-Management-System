package utils

import (
	"fmt"
	"net/smtp"
)

var (
	SMTP_HOST = "smtp.gmail.com"
	SMTP_PORT = "587"
	SMTP_USER = "ritwickbackup10@gmail.com"
	SMTP_PASS = "xtug cgkb lgbo fihn"
)

func SendHTMLEmail(to string, subject string, htmlBody string) error {
	from := SMTP_USER
	password := SMTP_PASS
	host := SMTP_HOST
	port := SMTP_PORT

	auth := smtp.PlainAuth("", from, password, host)

	headers := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n"
	headers += fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n\n", from, to, subject)

	message := headers + htmlBody

	return smtp.SendMail(host+":"+port, auth, from, []string{to}, []byte(message))
}
