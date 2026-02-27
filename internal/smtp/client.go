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

// Send sends an email via SMTP
func (c *Config) Send(to, msg string) error {
	if err := c.Validate(); err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)

	// Skip TLS for ports that don't support it (25, 80, 2525)
	skipTLS := c.Port == 25 || c.Port == 80 || c.Port == 2525

	if skipTLS {
		// Direct send without TLS using plain connection
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}
		defer conn.Close()

		c, err := smtp.NewClient(conn, c.Host)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
		defer c.Quit()

		// Send from noreply@maleon.run
		if err := c.Mail("noreply@maleon.run"); err != nil {
			return fmt.Errorf("failed to set From: %w", err)
		}

		if err := c.Rcpt(to); err != nil {
			return fmt.Errorf("failed to set To: %w", err)
		}

		w, err := c.Data()
		if err != nil {
			return fmt.Errorf("failed to get data writer: %w", err)
		}
		_, err = w.Write([]byte(msg))
		if err != nil {
			return fmt.Errorf("failed to write email: %w", err)
		}
		w.Close()

		return nil
	}

	// For TLS ports (587, 465, etc.)
	auth := smtp.PlainAuth("", c.Username, c.Password, c.Host)

	conn, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	// Upgrade to TLS if STARTTLS is available
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

	if err := conn.Mail("noreply@maleon.run"); err != nil {
		return fmt.Errorf("failed to set From: %w", err)
	}
	if err := conn.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set To: %w", err)
	}

	w, err := conn.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}
	_, err = w.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("failed to write email: %w", err)
	}
	w.Close()

	return nil
}

// SendWithAuth sends an email via SMTP with authentication
func (c *Config) SendWithAuth(from, to, msg string) error {
	if err := c.Validate(); err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	auth := smtp.PlainAuth("", c.Username, c.Password, c.Host)

	conn, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	// Upgrade to TLS if STARTTLS is available
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

	if err := conn.Mail(from); err != nil {
		return fmt.Errorf("failed to set From: %w", err)
	}
	if err := conn.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set To: %w", err)
	}

	w, err := conn.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}
	_, err = w.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("failed to write email: %w", err)
	}
	w.Close()

	return nil
}
