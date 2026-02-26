package email

import (
	"fmt"
	"net/smtp"
	"strings"
	"crypto/tls"
)

// Email represents an email message
type Email struct {
	From    string
	To      string
	Subject string
	Body    string
	HTML    string
}

// Validate checks if the email is valid
func (e *Email) Validate() error {
	if e.From == "" {
		return fmt.Errorf("from address is required")
	}
	if e.To == "" {
		return fmt.Errorf("to address is required")
	}
	if !strings.Contains(e.To, "@") {
		return fmt.Errorf("invalid to address: %s", e.To)
	}
	return nil
}

// BuildMessage builds the email message
func (e *Email) BuildMessage() string {
	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("From: %s\r\n", e.From))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", e.To))
	
	if e.Subject != "" {
		msg.WriteString(fmt.Sprintf("Subject: %s\r\n", e.Subject))
	}
	
	msg.WriteString("MIME-Version: 1.0\r\n")
	
	if e.HTML != "" {
		msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	} else {
		msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	}
	
	msg.WriteString("\r\n")
	
	if e.HTML != "" {
		msg.WriteString(e.HTML)
	} else {
		msg.WriteString(e.Body)
	}
	
	return msg.String()
}
