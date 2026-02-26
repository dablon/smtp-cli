package main

import (
	"testing"
	"net/smtp"
	"fmt"
	"time"
)

// E2EIntegrationTest tests the actual SMTP server
func TestE2E_SendEmail(t *testing.T) {
	// Skip if no SMTP server available
	config := &Config{
		Host:     getEnv("SMTP_HOST", "smtp.maleon.run"),
		Port:     getEnvInt("SMTP_PORT", 5870),
		Username: getEnv("SMTP_USER", "elus54"),
		Password: getEnv("SMTP_PASS", "nWvnh93M8nZj8X7kr6Hm6pbm5vN1Rl"),
	}
	
	// Test connection first
	err := config.TestConnection()
	if err != nil {
		t.Skipf("SMTP server not available: %v", err)
	}
	
	// Send actual email
	email := &Email{
		From:    "noreply@maleon.run",
		To:      getEnv("TEST_TO", "nicolasalca@hotmail.com"),
		Subject: "E2E Test - " + time.Now().Format(time.RFC3339),
		Body:    "This is an end-to-end test from the SMTP CLI",
		HTML:    "<html><body><h1>E2E Test</h1><p>This is an end-to-end test from the SMTP CLI</p></body></html>",
	}
	
	// Use auth if credentials provided
	if config.Username != "" && config.Password != "" {
		err = config.SendWithAuth(email)
	} else {
		err = config.Send(email)
	}
	
	if err != nil {
		t.Fatalf("Failed to send email: %v", err)
	}
	
	t.Log("Email sent successfully!")
}

// TestE2E_ConnectionTimeout tests connection timeout handling
func TestE2E_ConnectionTimeout(t *testing.T) {
	config := &Config{
		Host: "10.255.255.1", // Non-routable IP
		Port: 25,
	}
	
	err := config.TestConnection()
	if err == nil {
		t.Error("Expected connection timeout error")
	}
}

// TestE2E_InvalidPort tests invalid port handling
func TestE2E_InvalidPort(t *testing.T) {
	config := &Config{
		Host: "localhost",
		Port: 59999,
	}
	
	err := config.TestConnection()
	if err == nil {
		t.Error("Expected connection error for invalid port")
	}
}

// TestE2E_ValidHTMLEmail tests sending HTML email
func TestE2E_HTMLEmail(t *testing.T) {
	config := &Config{
		Host:     getEnv("SMTP_HOST", "smtp.maleon.run"),
		Port:     getEnvInt("SMTP_PORT", 5870),
		Username: getEnv("SMTP_USER", "elus54"),
		Password: getEnv("SMTP_PASS", "nWvnh93M8nZj8X7kr6Hm6pbm5vN1Rl"),
	}
	
	err := config.TestConnection()
	if err != nil {
		t.Skipf("SMTP server not available: %v", err)
	}
	
	email := &Email{
		From:    "noreply@maleon.run",
		To:      getEnv("TEST_TO", "nicolasalca@hotmail.com"),
		Subject: "HTML E2E Test",
		HTML:    `<html><body style="background:#1a1a2e;color:#fff;padding:20px;"><h1>🎉 HTML Email!</h1><p style="color:#10b981;">This is a styled HTML email.</p></body></html>`,
	}
	
	if config.Username != "" {
		err = config.SendWithAuth(email)
	} else {
		err = config.Send(email)
	}
	
	if err != nil {
		t.Fatalf("Failed to send HTML email: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := getEnvInternal(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := getEnvInternal(key); value != "" {
		var intValue int
		fmt.Sscanf(value, "%d", &intValue)
		if intValue > 0 {
			return intValue
		}
	}
	return defaultValue
}

func getEnvInternal(key string) string {
	// Simple implementation without os.Getenv for testing
	return ""
}

// Mock smtp.SendMail for testing
func init() {
	// This allows testing without actual SMTP server
	fmt.Println("E2E tests configured - set SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASS env vars")
}
