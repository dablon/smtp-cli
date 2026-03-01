package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/user"

	"smtp-cli/internal/smtp"
	"smtp-cli/pkg/email"
)

type Profile struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Pass     string `json:"pass"`
	From     string `json:"from"`
	TLS      bool   `json:"tls"`
}

type Profiles map[string]Profile

func getConfigDir() string {
	usr, _ := user.Current()
	homeDir := usr.HomeDir
	return homeDir + "/.smtp-cli"
}

func loadProfiles() Profiles {
	configDir := getConfigDir()
	configFile := configDir + "/profiles.json"
	
	data, err := os.ReadFile(configFile)
	if err != nil {
		return make(Profiles)
	}
	
	var profiles Profiles
	if err := json.Unmarshal(data, &profiles); err != nil {
		return make(Profiles)
	}
	return profiles
}

func saveProfiles(profiles Profiles) error {
	configDir := getConfigDir()
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}
	
	data, err := json.MarshalIndent(profiles, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(configDir+"/profiles.json", data, 0600)
}

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
	case "profile":
		handleProfile(os.Args[2:])
	case "profiles", "profile-list":
		handleListProfiles(os.Args[2:])
	case "help", "--help", "-h":
		fmt.Println(getHelp())
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println(getHelp())
		os.Exit(1)
	}
}

func handleProfile(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: smtp-cli profile <add|remove> [flags]")
		os.Exit(1)
	}

	action := args[0]

	switch action {
	case "add":
		handleAddProfile(args[1:])
	case "remove", "delete":
		handleRemoveProfile(args[1:])
	default:
		fmt.Printf("Unknown profile action: %s\n", action)
		os.Exit(1)
	}
}

func handleAddProfile(args []string) {
	addCmd := flag.NewFlagSet("profile add", flag.ExitOnError)
	name := addCmd.String("name", "", "Profile name (required)")
	host := addCmd.String("host", "smtp.maleon.run", "SMTP server host")
	port := addCmd.Int("port", 587, "SMTP server port")
	user := addCmd.String("user", "", "SMTP username")
	pass := addCmd.String("pass", "", "SMTP password")
	from := addCmd.String("from", "", "From address")
	tls := addCmd.Bool("tls", true, "Use TLS")

	addCmd.Parse(args)

	if *name == "" {
		fmt.Println("Error: --name is required")
		addCmd.Usage()
		os.Exit(1)
	}

	profiles := loadProfiles()
	
	fromAddr := *from
	if fromAddr == "" {
		fromAddr = *user + "@maleon.run"
	}

	profiles[*name] = Profile{
		Name: *name,
		Host: *host,
		Port: *port,
		User: *user,
		Pass: *pass,
		From: fromAddr,
		TLS:  *tls,
	}

	if err := saveProfiles(profiles); err != nil {
		fmt.Printf("Error saving profile: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Profile '%s' saved successfully!\n", *name)
}

func handleRemoveProfile(args []string) {
	removeCmd := flag.NewFlagSet("profile remove", flag.ExitOnError)
	name := removeCmd.String("name", "", "Profile name to remove")

	removeCmd.Parse(args)

	if *name == "" {
		fmt.Println("Error: --name is required")
		removeCmd.Usage()
		os.Exit(1)
	}

	profiles := loadProfiles()
	
	if _, ok := profiles[*name]; !ok {
		fmt.Printf("Profile '%s' not found\n", *name)
		os.Exit(1)
	}

	delete(profiles, *name)

	if err := saveProfiles(profiles); err != nil {
		fmt.Printf("Error saving profiles: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Profile '%s' removed successfully!\n", *name)
}

func handleListProfiles(_ []string) {
	profiles := loadProfiles()
	
	if len(profiles) == 0 {
		fmt.Println("No profiles found. Use 'smtp-cli profile add --name myprofile' to create one.")
		return
	}

	fmt.Println("Saved profiles:")
	fmt.Println("--------------")
	for name, p := range profiles {
		fmt.Printf("  %s -> %s:%d (%s)\n", name, p.Host, p.Port, p.From)
	}
}

func handleSend(args []string) {
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)

	host := sendCmd.String("host", "", "SMTP server host")
	port := sendCmd.Int("port", 0, "SMTP server port")
	user := sendCmd.String("user", "", "SMTP username")
	pass := sendCmd.String("pass", "", "SMTP password")
	from := sendCmd.String("from", "", "From address")
	to := sendCmd.String("to", "", "To address")
	subject := sendCmd.String("subject", "", "Email subject")
	body := sendCmd.String("body", "", "Email body (plain text)")
	html := sendCmd.String("html", "", "Email body (HTML)")
	profile := sendCmd.String("profile", "", "Use saved profile")

	sendCmd.Parse(args)

	var config *smtp.Config
	var fromAddr string

	if *profile != "" {
		profiles := loadProfiles()
		p, ok := profiles[*profile]
		if !ok {
			fmt.Printf("Profile '%s' not found\n", *profile)
			os.Exit(1)
		}
		
		config = &smtp.Config{
			Host:     p.Host,
			Port:     p.Port,
			Username: p.User,
			Password: p.Pass,
		}
		fromAddr = p.From
		
		if *host != "" {
			config.Host = *host
		}
		if *port != 0 {
			config.Port = *port
		}
		if *user != "" {
			config.Username = *user
		}
		if *pass != "" {
			config.Password = *pass
		}
		if *from != "" {
			fromAddr = *from
		}
	} else {
		hostVal := *host
		if hostVal == "" {
			hostVal = "smtp.maleon.run"
		}
		portVal := *port
		if portVal == 0 {
			portVal = 587
		}
		
		config = &smtp.Config{
			Host:     hostVal,
			Port:     portVal,
			Username: *user,
			Password: *pass,
		}
		fromAddr = *from
		if fromAddr == "" {
			fromAddr = *user + "@maleon.run"
		}
	}

	if *to == "" {
		fmt.Println("Error: --to is required")
		sendCmd.Usage()
		os.Exit(1)
	}

	e := &email.Email{
		From:    fromAddr,
		To:      *to,
		Subject: *subject,
		Body:    *body,
		HTML:    *html,
	}

	var err error
	if config.Username != "" && config.Password != "" {
		err = config.SendWithAuth(e.From, e.To, e.BuildMessage())
	} else {
		err = config.Send(e.To, e.BuildMessage())
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Email sent successfully!")
}

func handleTest(args []string) {
	testCmd := flag.NewFlagSet("test", flag.ExitOnError)

	host := testCmd.String("host", "", "SMTP server host")
	port := testCmd.Int("port", 0, "SMTP server port")
	user := testCmd.String("user", "", "SMTP username")
	pass := testCmd.String("pass", "", "SMTP password")
	profile := testCmd.String("profile", "", "Use saved profile")

	testCmd.Parse(args)

	var config *smtp.Config

	if *profile != "" {
		profiles := loadProfiles()
		p, ok := profiles[*profile]
		if !ok {
			fmt.Printf("Profile '%s' not found\n", *profile)
			os.Exit(1)
		}
		
		config = &smtp.Config{
			Host:     p.Host,
			Port:     p.Port,
			Username: p.User,
			Password: p.Pass,
		}
		
		if *host != "" {
			config.Host = *host
		}
		if *port != 0 {
			config.Port = *port
		}
		if *user != "" {
			config.Username = *user
		}
		if *pass != "" {
			config.Password = *pass
		}
	} else {
		hostVal := *host
		if hostVal == "" {
			hostVal = "smtp.maleon.run"
		}
		portVal := *port
		if portVal == 0 {
			portVal = 587
		}
		
		config = &smtp.Config{
			Host:     hostVal,
			Port:     portVal,
			Username: *user,
			Password: *pass,
		}
	}

	err := config.Validate()
	if err != nil {
		fmt.Printf("Connection test failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Connection successful!")
}

func getHelp() string {
	return `SMTP CLI - Simple SMTP Email Client with Profile Support

Usage:
  smtp-cli send [flags]                Send an email
  smtp-cli test [flags]                Test SMTP connection
  smtp-cli profile add [flags]         Add a new profile
  smtp-cli profile remove [flags]       Remove a profile
  smtp-cli profiles                    List all saved profiles
  smtp-cli help                        Show this help

Commands:
  send          Send an email message
  test          Test SMTP server connection
  profile       Manage email profiles (add/remove)
  profiles      List all saved profiles
  help          Show help information

Examples:
  # Add a profile
  smtp-cli profile add --name work --host smtp.maleon.run --user myuser --pass mypass

  # Send using profile
  smtp-cli send -profile work -to user@example.com -subject "Hello" -body "Message"

  # Send without profile (direct)
  smtp-cli send -to user@example.com -subject "Hello" -body "Message"

  # Test connection with profile
  smtp-cli test -profile work

  # List profiles
  smtp-cli profiles

  # Remove profile
  smtp-cli profile remove --name work

Global Flags:
  --profile string    Use saved profile (overrides other flags if provided)

Send Flags:
  --host string       SMTP server hostname
  --port int          SMTP server port
  --user string       SMTP username
  --pass string       SMTP password
  --from string       From address
  --to string         To address (required)
  --subject string    Email subject
  --body string       Plain text body
  --html string       Email body (HTML)

Profile Add Flags:
  --name string       Profile name (required)
  --host string       SMTP server hostname
  --port int          SMTP server port
  --user string       SMTP username
  --pass string       SMTP password
  --from string       From address
  --tls               Use TLS (default true)

Profile Remove Flags:
  --name string       Profile name to remove (required)
`
}
