package main

import (
	"flag"
	"fmt"
	"os"

	"smtp-cli/internal/smtp"
	"smtp-cli/pkg/email"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println(getHelp())
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "send":
		handleSend(os.Args[2:])
	case "test":
		handleTest(os.Args[2:])
	case "help", "--help", "-h":
		fmt.Println(getHelp())
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println(getHelp())
		os.Exit(1)
	}
}

func handleSend(args []string) {
	// Create a new FlagSet for the send command
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)

	host := sendCmd.String("host", "smtp.maleon.run", "SMTP server host")
	port := sendCmd.Int("port", 5870, "SMTP server port")
	user := sendCmd.String("user", "elus54", "SMTP username")
	pass := sendCmd.String("pass", "", "SMTP password")
	from := sendCmd.String("from", "noreply@maleon.run", "From address")
	to := sendCmd.String("to", "", "To address")
	subject := sendCmd.String("subject", "", "Email subject")
	body := sendCmd.String("body", "", "Email body (plain text)")
	html := sendCmd.String("html", "", "Email body (HTML)")

	sendCmd.Parse(args)

	if *to == "" {
		fmt.Println("Error: --to is required")
		sendCmd.Usage()
		os.Exit(1)
	}

	config := &smtp.Config{
		Host:     *host,
		Port:     *port,
		Username: *user,
		Password: *pass,
	}

	e := &email.Email{
		From:    *from,
		To:      *to,
		Subject: *subject,
		Body:    *body,
		HTML:    *html,
	}

	var err error
	if *user != "" && *pass != "" {
		err = config.SendWithAuth(e.From, e.To, e.BuildMessage())
	} else {
		err = config.Send(e.BuildMessage())
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Email sent successfully!")
}

func handleTest(args []string) {
	// Create a new FlagSet for the test command
	testCmd := flag.NewFlagSet("test", flag.ExitOnError)

	host := testCmd.String("host", "smtp.maleon.run", "SMTP server host")
	port := testCmd.Int("port", 5870, "SMTP server port")
	user := testCmd.String("user", "elus54", "SMTP username")
	pass := testCmd.String("pass", "", "SMTP password")

	testCmd.Parse(args)

	config := &smtp.Config{
		Host:     *host,
		Port:     *port,
		Username: *user,
		Password: *pass,
	}

	err := config.TestConnection()
	if err != nil {
		fmt.Printf("Connection test failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Connection successful!")
}

func getHelp() string {
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
  smtp-cli send -to user@example.com -subject "Hello" -body "Message"
  smtp-cli send -to user@example.com -subject "Hello" -html "<h1>Title</h1>"
  smtp-cli test -host smtp.example.com -port 587

Flags:
  --host string       SMTP server hostname (default "smtp.maleon.run")
  --port int          SMTP server port (default 5870)
  --user string       SMTP username
  --pass string       SMTP password
  --from string       From address (default "noreply@maleon.run")
  --to string         To address (required for send)
  --subject string    Email subject
  --body string       Plain text body
  --html string       Email body (HTML)
  --help              Show help
`
}
