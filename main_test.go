package main

import (
	"testing"
)

// Mock SMTP server for testing
type mockSMTP struct {
	connected bool
	sent      []string
	err       error
}

func (m *mockSMTP) Dial(addr string) (interface{}, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.connected = true
	return &mockClient{m: m}, nil
}

type mockClient struct {
	m *mockSMTP
}

func (c *mockClient) Close() error {
	c.m.connected = false
	return nil
}

func TestEmailValidate(t *testing.T) {
	tests := []struct {
		name    string
		email   *Email
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid email",
			email:   &Email{From: "test@test.com", To: "dest@dest.com"},
			wantErr: false,
		},
		{
			name:    "missing from",
			email:   &Email{To: "dest@dest.com"},
			wantErr: true,
			errMsg:  "from address is required",
		},
		{
			name:    "missing to",
			email:   &Email{From: "test@test.com"},
			wantErr: true,
			errMsg:  "to address is required",
		},
		{
			name:    "invalid to address",
			email:   &Email{From: "test@test.com", To: "invalid"},
			wantErr: true,
			errMsg:  "invalid to address",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.email.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
				} else if !contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid config",
			config:  &Config{Host: "smtp.test.com", Port: 587},
			wantErr: false,
		},
		{
			name:    "missing host",
			config:  &Config{Port: 587},
			wantErr: true,
			errMsg:  "host is required",
		},
		{
			name:    "missing port",
			config:  &Config{Host: "smtp.test.com"},
			wantErr: true,
			errMsg:  "port is required",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.errMsg)
				} else if !contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestEmailBuildMessage(t *testing.T) {
	tests := []struct {
		name      string
		email     *Email
		contains  []string
		notContain []string
	}{
		{
			name:  "plain text email",
			email: &Email{
				From:    "from@test.com",
				To:      "to@test.com",
				Subject: "Test Subject",
				Body:    "Hello World",
			},
			contains: []string{
				"From: from@test.com",
				"To: to@test.com",
				"Subject: Test Subject",
				"Hello World",
				"text/plain",
			},
			notContain: []string{"text/html"},
		},
		{
			name:  "HTML email",
			email: &Email{
				From:    "from@test.com",
				To:      "to@test.com",
				Subject: "HTML Subject",
				HTML:    "<h1>Hello</h1>",
			},
			contains: []string{
				"From: from@test.com",
				"To: to@test.com",
				"Subject: HTML Subject",
				"<h1>Hello</h1>",
				"text/html",
			},
			notContain: []string{"text/plain"},
		},
		{
			name:  "email without subject",
			email: &Email{
				From: "from@test.com",
				To:   "to@test.com",
				Body: "No subject",
			},
			contains: []string{
				"From: from@test.com",
				"To: to@test.com",
				"No subject",
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.email.BuildMessage()
			
			for _, c := range tt.contains {
				if !contains(msg, c) {
					t.Errorf("expected message to contain %q, got:\n%s", c, msg)
				}
			}
			
			for _, c := range tt.notContain {
				if contains(msg, c) {
					t.Errorf("expected message NOT to contain %q, got:\n%s", c, msg)
				}
			}
		})
	}
}

func TestGetHelp(t *testing.T) {
	help := GetHelp()
	
	tests := []string{
		"SMTP CLI",
		"send",
		"test",
		"--host",
		"--port",
		"--to",
		"--subject",
		"--body",
		"--html",
		"Examples",
	}
	
	for _, s := range tests {
		if !contains(help, s) {
			t.Errorf("expected help to contain %q", s)
		}
	}
}

func TestEmailWithAllFields(t *testing.T) {
	email := &Email{
		From:    "noreply@maleon.run",
		To:      "test@example.com",
		Subject: "Complete Test",
		Body:    "Plain text body",
		HTML:    "<p>HTML body</p>",
	}
	
	err := email.Validate()
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
	
	msg := email.BuildMessage()
	
	// HTML should take precedence
	if !contains(msg, "<p>HTML body</p>") {
		t.Error("HTML body not found in message")
	}
	if contains(msg, "Plain text body") {
		t.Error("Plain text should not be in message when HTML is set")
	}
}

func TestConfigWithAuth(t *testing.T) {
	config := &Config{
		Host:     "smtp.maleon.run",
		Port:     5870,
		Username: "elus54",
		Password: "testpass",
	}
	
	err := config.Validate()
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestConfigDefaults(t *testing.T) {
	config := &Config{
		Host: "smtp.maleon.run",
		Port: 5870,
	}
	
	if config.Host != "smtp.maleon.run" {
		t.Errorf("expected host smtp.maleon.run, got %s", config.Host)
	}
	if config.Port != 5870 {
		t.Errorf("expected port 5870, got %d", config.Port)
	}
}

func TestEmailEmptyBody(t *testing.T) {
	email := &Email{
		From: "from@test.com",
		To:   "to@test.com",
	}
	
	err := email.Validate()
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
	
	msg := email.BuildMessage()
	if !contains(msg, "From: from@test.com") {
		t.Error("From not in message")
	}
	if !contains(msg, "To: to@test.com") {
		t.Error("To not in message")
	}
}

func TestMultipleRecipients(t *testing.T) {
	email := &Email{
		From: "from@test.com",
		To:   "user1@test.com,user2@test.com",
	}
	
	err := email.Validate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	msg := email.BuildMessage()
	if !contains(msg, "To: user1@test.com,user2@test.com") {
		t.Error("Multiple recipients not in message")
	}
}

func TestSpecialCharactersInBody(t *testing.T) {
	email := &Email{
		From: "from@test.com",
		To:   "to@test.com",
		Body: "Special chars: ñ áéíóú @ # $ % & * ()",
	}
	
	msg := email.BuildMessage()
	if !contains(msg, "ñ áéíóú") {
		t.Error("Special characters not preserved")
	}
}

func TestHTMLWithStyles(t *testing.T) {
	html := `<html><body style="background: #000; color: #fff;"><h1>Test</h1></body></html>`
	email := &Email{
		From: "from@test.com",
		To:   "to@test.com",
		HTML: html,
	}
	
	msg := email.BuildMessage()
	if !contains(msg, html) {
		t.Error("HTML content not preserved")
	}
	if !contains(msg, "text/html") {
		t.Error("Content-Type should be text/html")
	}
}

func BenchmarkEmailBuildMessage(b *testing.B) {
	email := &Email{
		From:    "from@test.com",
		To:      "to@test.com",
		Subject: "Benchmark Test",
		Body:    "This is a test body with some content",
		HTML:    "<p>Benchmark HTML</p>",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		email.BuildMessage()
	}
}

func BenchmarkEmailValidate(b *testing.B) {
	email := &Email{
		From: "test@test.com",
		To:   "dest@dest.com",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		email.Validate()
	}
}
