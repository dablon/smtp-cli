package main

import (
	"flag"
	"fmt"
	"os"

	"smtp-cli/internal/smtp"
	"smtp-cli/pkg/email"
)

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
		fmt.Println(getHelp())
		os.Exit(1)
	}

	config := &smtp.Config{
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

	case "test":
		err := config.TestConnection()
		if err != nil {
			fmt.Printf("Connection test failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Connection successful!")

	case "help", "--help", "-h":
		fmt.Println(getHelp())

	default:
		fmt.Printf("Unknown command: %s\n", args[0])
		fmt.Println(getHelp())
		os.Exit(1)
	}
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
