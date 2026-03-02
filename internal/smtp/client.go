package smtp

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
)

// Config holds SMTP configuration
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	TLS      bool // Force implicit TLS (port 465 style)
	STARTTLS bool // Force STARTTLS upgrade
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}
	if c.Port == 0 {
		return fmt.Errorf("port is required")
	}
	return nil
}

// Send sends an email via SMTP without authentication
func (c *Config) Send(to, msg string) error {
	if err := c.Validate(); err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	tlsConfig := &tls.Config{ServerName: c.Host, InsecureSkipVerify: true}

	// Implicit TLS (port 465 style): connect over TLS directly
	if c.TLS {
		tlsConn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("TLS connection failed: %w", err)
		}
		defer tlsConn.Close()

		client, err := smtp.NewClient(tlsConn, c.Host)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
		defer client.Quit()

		return sendMessage(client, "noreply@maleon.run", to, msg)
	}

	// Plain or STARTTLS connection
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, c.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	// STARTTLS upgrade
	if c.STARTTLS {
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("STARTTLS failed: %w", err)
		}
	} else if ok, _ := client.Extension("STARTTLS"); ok {
		// Auto-upgrade if server supports it
		client.StartTLS(tlsConfig)
	}

	return sendMessage(client, "noreply@maleon.run", to, msg)
}

// SendWithAuth sends an email via SMTP with authentication
func (c *Config) SendWithAuth(from, to, msg string) error {
	if err := c.Validate(); err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	auth := smtp.PlainAuth("", c.Username, c.Password, c.Host)
	tlsConfig := &tls.Config{ServerName: c.Host, InsecureSkipVerify: true}

	// Implicit TLS (port 465 style): connect over TLS directly
	if c.TLS {
		tlsConn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("TLS connection failed: %w", err)
		}
		defer tlsConn.Close()

		client, err := smtp.NewClient(tlsConn, c.Host)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
		defer client.Quit()

		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}
		return sendMessage(client, from, to, msg)
	}

	// Plain or STARTTLS connection
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, c.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	// STARTTLS upgrade
	if c.STARTTLS {
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("STARTTLS failed: %w", err)
		}
	} else if ok, _ := client.Extension("STARTTLS"); ok {
		client.StartTLS(tlsConfig)
	}

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	return sendMessage(client, from, to, msg)
}

// sendMessage sends the email envelope and body through an SMTP client
func sendMessage(client *smtp.Client, from, to, msg string) error {
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("failed to set From: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set To: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}
	if _, err = w.Write([]byte(msg)); err != nil {
		return fmt.Errorf("failed to write email: %w", err)
	}
	return w.Close()
}
