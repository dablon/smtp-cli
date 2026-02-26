package email

import (
	"testing"
	"strings"
)

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
				} else if !strings.Contains(err.Error(), tt.errMsg) {
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
				if !strings.Contains(msg, c) {
					t.Errorf("expected message to contain %q, got:\n%s", c, msg)
				}
			}

			for _, c := range tt.notContain {
				if strings.Contains(msg, c) {
					t.Errorf("expected message NOT to contain %q, got:\n%s", c, msg)
				}
			}
		})
	}
}

func TestEmailWithAllFields(t *testing.T) {
	e := &Email{
		From:    "noreply@maleon.run",
		To:      "test@example.com",
		Subject: "Complete Test",
		Body:    "Plain text body",
		HTML:    "<p>HTML body</p>",
	}

	err := e.Validate()
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}

	msg := e.BuildMessage()

	// HTML should take precedence
	if !strings.Contains(msg, "<p>HTML body</p>") {
		t.Error("HTML body not found in message")
	}
	if strings.Contains(msg, "Plain text body") {
		t.Error("Plain text should not be in message when HTML is set")
	}
}

func TestEmailEmptyBody(t *testing.T) {
	e := &Email{
		From: "from@test.com",
		To:   "to@test.com",
	}

	err := e.Validate()
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}

	msg := e.BuildMessage()
	if !strings.Contains(msg, "From: from@test.com") {
		t.Error("From not in message")
	}
	if !strings.Contains(msg, "To: to@test.com") {
		t.Error("To not in message")
	}
}

func TestMultipleRecipients(t *testing.T) {
	e := &Email{
		From: "from@test.com",
		To:   "user1@test.com,user2@test.com",
	}

	err := e.Validate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg := e.BuildMessage()
	if !strings.Contains(msg, "To: user1@test.com,user2@test.com") {
		t.Error("Multiple recipients not in message")
	}
}

func TestSpecialCharactersInBody(t *testing.T) {
	e := &Email{
		From: "from@test.com",
		To:   "to@test.com",
		Body: "Special chars: ñ áéíóú @ # $ % & * ()",
	}

	msg := e.BuildMessage()
	if !strings.Contains(msg, "ñ áéíóú") {
		t.Error("Special characters not preserved")
	}
}

func TestHTMLWithStyles(t *testing.T) {
	html := `<html><body style="background: #000; color: #fff;"><h1>Test</h1></body></html>`
	e := &Email{
		From: "from@test.com",
		To:   "to@test.com",
		HTML: html,
	}

	msg := e.BuildMessage()
	if !strings.Contains(msg, html) {
		t.Error("HTML content not preserved")
	}
	if !strings.Contains(msg, "text/html") {
		t.Error("Content-Type should be text/html")
	}
}

func BenchmarkEmailBuildMessage(b *testing.B) {
	e := &Email{
		From:    "from@test.com",
		To:      "to@test.com",
		Subject: "Benchmark Test",
		Body:    "This is a test body with some content",
		HTML:    "<p>Benchmark HTML</p>",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.BuildMessage()
	}
}

func BenchmarkEmailValidate(b *testing.B) {
	e := &Email{
		From: "test@test.com",
		To:   "dest@dest.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Validate()
	}
}
