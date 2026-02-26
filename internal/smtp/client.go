package smtp

import (
	"fmt"
	"net/smtp"
	"crypto/tls"
)

// Config holds SMTP configuration
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
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

// Send sends an email via SMTP
func (c *Config) Send(msg string) error {
	if err := c.Validate(); err != nil {
		return err
	}
	
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	
	// For unauthenticated sending
	err := smtp.SendMail(addr, nil, "", []string{""}, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	
	return nil
}

// SendWithAuth sends an email via SMTP with authentication
func (c *Config) SendWithAuth(from, to, msg string) error {
	if err := c.Validate(); err != nil {
		return err
	}
	
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	
	// If no username/password, try without auth
	if c.Username == "" || c.Password == "" {
		return c.Send(msg)
	}
	
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
			ServerName:         c.Host,
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
	if err := conn.Mail(from); err != nil {
		return fmt.Errorf("failed to set From: %w", err)
	}
	if err := conn.Rcpt(to); err != nil {
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
