package email

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// Attachment represents a file attachment
type Attachment struct {
	Filename string
	Data     []byte
}

// Email represents an email message
type Email struct {
	From       string
	To         string
	Subject    string
	Body       string
	HTML       string
	Attachments []Attachment
}

// AddAttachment adds a file attachment
func (e *Email) AddAttachment(filepath string) error {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %v", filepath, err)
	}
	e.Attachments = append(e.Attachments, Attachment{
		Filename: filepath,
		Data:     data,
	})
	return nil
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
	
	// If no attachments, use simple message
	if len(e.Attachments) == 0 {
		return e.buildSimpleMessage()
	}
	
	// Build multipart message
	boundary := "----=_Part_" + fmt.Sprintf("%d", len(e.Attachments))
	
	msg.WriteString(fmt.Sprintf("From: %s\r\n", e.From))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", e.To))
	if e.Subject != "" {
		msg.WriteString(fmt.Sprintf("Subject: %s\r\n", e.Subject))
	}
	msg.WriteString(fmt.Sprintf("MIME-Version: 1.0\r\n"))
	msg.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n", boundary))
	msg.WriteString("\r\n")
	
	// Body part
	msg.WriteString("--" + boundary + "\r\n")
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
	msg.WriteString("\r\n")
	
	// Attachment parts
	for _, att := range e.Attachments {
		filename := filepath.Base(att.Filename)
		msg.WriteString("--" + boundary + "\r\n")
		msg.WriteString("Content-Type: application/octet-stream\r\n")
		msg.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=%s\r\n", filename))
		msg.WriteString("Content-Transfer-Encoding: base64\r\n")
		msg.WriteString("\r\n")
		
		encoded := base64.StdEncoding.EncodeToString(att.Data)
		msg.WriteString(encoded)
		msg.WriteString("\r\n")
	}
	
	msg.WriteString("--" + boundary + "--\r\n")
	
	return msg.String()
}

func (e *Email) buildSimpleMessage() string {
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
