package main

import (
	"fmt"
	"net/smtp"
	"strings"
	"os"
	"flag"
	"bytes"
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

// Config holds SMTP configuration
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
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

// Validate checks if the config is valid
func (c *Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}
	if c.Port == 0 {
		return fmt.Errorf("port is required")
	}
	return nil
}

// BuildMessage builds the email message
func (e *Email) BuildMessage() string {
	var msg bytes.Buffer
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

// Send sends an email via SMTP
func (c *Config) Send(e *Email) error {
	if err := c.Validate(); err != nil {
		return err
	}
	if err := e.Validate(); err != nil {
		return err
	}
	
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	msg := e.BuildMessage()
	
	// For unauthenticated sending
	err := smtp.SendMail(addr, nil, e.From, []string{e.To}, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	
	return nil
}

// SendWithAuth sends an email via SMTP with authentication
func (c *Config) SendWithAuth(e *Email) error {
	if err := c.Validate(); err != nil {
		return err
	}
	if err := e.Validate(); err != nil {
		return err
	}
	
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	msg := e.BuildMessage()
	
	auth := smtp.PlainAuth("", c.Username, c.Password, c.Host)
	
	// Connect without TLS first
	conn, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()
	
	// Upgrade to TLS if needed (STARTTLS)
	if ok, _ := conn.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{
			ServerName: c.Host,
			InsecureSkipVerify: true,
		}
		if err := conn.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("failed to start TLS: %w", err)
		}
	}
	
	// Authenticate
	if err := conn.Auth(auth); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	
	// Set From and To
	if err := conn.Mail(e.From); err != nil {
		return fmt.Errorf("failed to set From: %w", err)
	}
	if err := conn.Rcpt(e.To); err != nil {
		return fmt.Errorf("failed to set To: %w", err)
	}
	
	// Send body
	w, err := conn.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}
	_, err = w.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("failed to write email: %w", err)
	}
	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}
	
	return nil
}

// TestConnection tests the SMTP connection
func (c *Config) TestConnection() error {
	if err := c.Validate(); err != nil {
		return err
	}
	
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	
	// Try to connect
	conn, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer conn.Close()
	
	return nil
}

// GetHelp returns help text
func GetHelp() string {
	return `SMTP CLI - Simple SMTP Email Client

Usage:
  smtp-cli send [flags]          Send an email
  smtp-cli test                   Test SMTP connection
  smtp-cli help                  Show this help

Commands:
  send      Send an email message
  test      Test SMTP server connection
  help      Show help information

Examples:
  smtp-cli send --to user@example.com --subject "Hello" --body "Message"
  smtp-cli send --to user@example.com --subject "HTML" --html "<h1>Title</h1>"
  smtp-cli test --host smtp.example.com --port 587

Flags:
  --host string       SMTP server hostname (default "smtp.maleon.run")
  --port int          SMTP server port (default 5870)
  --user string       SMTP username
  --pass string       SMTP password
  --from string       From address (default "noreply@maleon.run")
  --to string         To address (required for send)
  --subject string    Email subject
  --body string       Plain text body
  --html string       HTML body
  --help              Show help
`
}

func main() {
	// Define flags
	host := flag.String("host", "smtp.maleon.run", "SMTP server host")
	port := flag.Int("port", 5870, "SMTP server port")
	user := flag.String("user", "elus54", "SMTP username")
	pass := flag.String("pass", "", "SMTP password")
	from := flag.String("from", "noreply@maleon.run", "From address")
	to := flag.String("to", "", "To address")
	subject := flag.String("subject", "", "Email subject")
	body := flag.String("body", "", "Email body (plain text)")
	html := flag.String("html", "", "Email body (HTML)")
	
	flag.Parse()
	
	args := flag.Args()
	if len(args) == 0 {
		fmt.Println(GetHelp())
		os.Exit(1)
	}
	
	config := &Config{
		Host:     *host,
		Port:     *port,
		Username: *user,
		Password: *pass,
	}
	
	switch args[0] {
	case "send":
		if *to == "" {
			fmt.Println("Error: --to is required")
			os.Exit(1)
		}
		
		email := &Email{
			From:    *from,
			To:      *to,
			Subject: *subject,
			Body:    *body,
			HTML:    *html,
		}
		
		var err error
		if *user != "" && *pass != "" {
			err = config.SendWithAuth(email)
		} else {
			err = config.Send(email)
		}
		
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Println("Email sent successfully!")
		
	case "test":
		err := config.TestConnection()
		if err != nil {
			fmt.Printf("Connection test failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Connection successful!")
		
	case "help", "--help", "-h":
		fmt.Println(GetHelp())
		
	default:
		fmt.Printf("Unknown command: %s\n", args[0])
		fmt.Println(GetHelp())
		os.Exit(1)
	}
}
