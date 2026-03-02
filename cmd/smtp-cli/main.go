package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/user"
	"strings"

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
	APIURL   string `json:"api_url"`
}

type Profiles map[string]Profile

func getConfigDir() string {
	usr, _ := user.Current()
	return usr.HomeDir + "/.smtp-cli"
}

func loadProfiles() Profiles {
	configFile := getConfigDir() + "/profiles.json"
	data, err := os.ReadFile(configFile)
	if err != nil {
		return make(Profiles)
	}
	var profiles Profiles
	json.Unmarshal(data, &profiles)
	return profiles
}

func saveProfiles(profiles Profiles) error {
	os.MkdirAll(getConfigDir(), 0700)
	data, _ := json.MarshalIndent(profiles, "", "  ")
	return os.WriteFile(getConfigDir()+"/profiles.json", data, 0600)
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
		handleListProfiles()
	case "inbox":
		handleInbox(os.Args[2:])
	case "sent":
		handleSent(os.Args[2:])
	case "list":
		handleList(os.Args[2:])
	case "read":
		handleRead(os.Args[2:])
	case "delete":
		handleDelete(os.Args[2:])
	case "login":
		handleLogin(os.Args[2:])
	case "help", "--help", "-h":
		fmt.Println(getHelp())
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println(getHelp())
		os.Exit(1)
	}
}

// ============ PROFILE COMMANDS ============

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
		fmt.Printf("Unknown action: %s\n", action)
	}
}

func handleAddProfile(args []string) {
	cmd := flag.NewFlagSet("profile add", flag.ExitOnError)
	name := cmd.String("name", "", "Profile name (required)")
	host := cmd.String("host", "smtp.maleon.run", "SMTP server host")
	port := cmd.Int("port", 587, "SMTP server port")
	user := cmd.String("user", "", "SMTP username")
	pass := cmd.String("pass", "", "SMTP password")
	from := cmd.String("from", "", "From address")
	api := cmd.String("api", "http://localhost:3002", "Mail Manager API URL")
	cmd.Parse(args)

	if *name == "" {
		fmt.Println("Error: --name is required")
		os.Exit(1)
	}

	profiles := loadProfiles()
	fromAddr := *from
	if fromAddr == "" {
		fromAddr = *user + "@maleon.run"
	}

	profiles[*name] = Profile{
		Name:   *name,
		Host:   *host,
		Port:   *port,
		User:   *user,
		Pass:   *pass,
		From:   fromAddr,
		TLS:    true,
		APIURL: *api,
	}
	saveProfiles(profiles)
	fmt.Printf("Profile '%s' saved!\n", *name)
}

func handleRemoveProfile(args []string) {
	cmd := flag.NewFlagSet("profile remove", flag.ExitOnError)
	name := cmd.String("name", "", "Profile name")
	cmd.Parse(args)

	if *name == "" {
		fmt.Println("Error: --name required")
		os.Exit(1)
	}

	profiles := loadProfiles()
	delete(profiles, *name)
	saveProfiles(profiles)
	fmt.Printf("Profile '%s' removed!\n", *name)
}

func handleListProfiles() {
	profiles := loadProfiles()
	if len(profiles) == 0 {
		fmt.Println("No profiles. Use: smtp-cli profile add --name myprofile")
		return
	}
	fmt.Println("Saved profiles:")
	fmt.Println("-------------")
	for name, p := range profiles {
		fmt.Printf("  %s -> %s:%d (%s)\n", name, p.Host, p.Port, p.From)
	}
}

// ============ MAIL COMMANDS ============

func getAPIURL(profile string) string {
	profiles := loadProfiles()
	if p, ok := profiles[profile]; ok {
		return p.APIURL
	}
	return "http://localhost:3002"
}

func makeAuthHeader(profile string) string {
	profiles := loadProfiles()
	if _, ok := profiles[profile]; ok {
		// Try to get token from cache or login
		tokenFile := getConfigDir() + "/" + profile + ".token"
		if data, err := os.ReadFile(tokenFile); err == nil {
			return strings.TrimSpace(string(data))
		}
	}
	return ""
}

func handleLogin(args []string) {
	cmd := flag.NewFlagSet("login", flag.ExitOnError)
	profile := cmd.String("profile", "default", "Profile name")
	email := cmd.String("email", "", "Email")
	password := cmd.String("pass", "", "Password")
	cmd.Parse(args)

	if *email == "" || *password == "" {
		fmt.Println("Usage: smtp-cli login --profile default --email user@domain --pass password")
		os.Exit(1)
	}

	apiURL := getAPIURL(*profile)
	
	// Login to API
	resp, err := http.Post(apiURL+"/api/auth/login", "application/json", 
		strings.NewReader(fmt.Sprintf(`{"email":"%s","password":"%s"}`, *email, *password)))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		fmt.Printf("Login failed: %s\n", string(body))
		os.Exit(1)
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)
	
	token := result["token"].(string)
	
	// Save token
	os.MkdirAll(getConfigDir(), 0700)
	os.WriteFile(getConfigDir()+"/"+*profile+".token", []byte(token), 0600)
	
	fmt.Println("Logged in successfully!")
}

func handleInbox(args []string) {
	cmd := flag.NewFlagSet("inbox", flag.ExitOnError)
	profile := cmd.String("profile", "default", "Profile name")
	cmd.Parse(args)

	apiURL := getAPIURL(*profile)
	token := makeAuthHeader(*profile)

	if token == "" {
		fmt.Println("Not logged in. Run: smtp-cli login --profile default --email user@domain --pass password")
		os.Exit(1)
	}

	req, _ := http.NewRequest("GET", apiURL+"/api/emails?folder=inbox", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func handleSent(args []string) {
	cmd := flag.NewFlagSet("sent", flag.ExitOnError)
	profile := cmd.String("profile", "default", "Profile name")
	cmd.Parse(args)

	apiURL := getAPIURL(*profile)
	token := makeAuthHeader(*profile)

	if token == "" {
		fmt.Println("Not logged in. Run: smtp-cli login --profile default --email user@domain --pass password")
		os.Exit(1)
	}

	req, _ := http.NewRequest("GET", apiURL+"/api/emails?folder=sent", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func handleList(args []string) {
	// Alias for inbox
	handleInbox(args)
}

func handleRead(args []string) {
	cmd := flag.NewFlagSet("read", flag.ExitOnError)
	profile := cmd.String("profile", "default", "Profile name")
	id := cmd.Int("id", 0, "Email ID")
	cmd.Parse(args)

	if *id == 0 {
		fmt.Println("Usage: smtp-cli read --id 123")
		os.Exit(1)
	}

	apiURL := getAPIURL(*profile)
	token := makeAuthHeader(*profile)

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/emails/%d", apiURL, *id), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func handleDelete(args []string) {
	cmd := flag.NewFlagSet("delete", flag.ExitOnError)
	profile := cmd.String("profile", "default", "Profile name")
	id := cmd.Int("id", 0, "Email ID")
	cmd.Parse(args)

	if *id == 0 {
		fmt.Println("Usage: smtp-cli delete --id 123")
		os.Exit(1)
	}

	apiURL := getAPIURL(*profile)
	token := makeAuthHeader(*profile)

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/api/emails/%d", apiURL, *id), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	
	if resp.StatusCode == 200 {
		fmt.Println("Email deleted!")
	} else {
		fmt.Printf("Error: %d\n", resp.StatusCode)
	}
}

// ============ SEND & TEST ============

func handleSend(args []string) {
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	host := sendCmd.String("host", "", "SMTP server")
	port := sendCmd.Int("port", 0, "SMTP port")
	user := sendCmd.String("user", "", "SMTP user")
	pass := sendCmd.String("pass", "", "SMTP pass")
	from := sendCmd.String("from", "", "From address")
	to := sendCmd.String("to", "", "To address")
	subject := sendCmd.String("subject", "", "Subject")
	body := sendCmd.String("body", "", "Body")
	html := sendCmd.String("html", "", "HTML body")
	profile := sendCmd.String("profile", "", "Use profile")

	sendCmd.Parse(args)

	var config *smtp.Config
	var fromAddr string

	if *profile != "" {
		profiles := loadProfiles()
		if p, ok := profiles[*profile]; ok {
			config = &smtp.Config{Host: p.Host, Port: p.Port, Username: p.User, Password: p.Pass}
			fromAddr = p.From
			if *host != "" { config.Host = *host }
			if *port != 0 { config.Port = *port }
			if *user != "" { config.Username = *user }
			if *pass != "" { config.Password = *pass }
			if *from != "" { fromAddr = *from }
		}
	} else {
		hostVal := *host
		if hostVal == "" { hostVal = "smtp.maleon.run" }
		portVal := *port
		if portVal == 0 { portVal = 587 }
		config = &smtp.Config{Host: hostVal, Port: portVal, Username: *user, Password: *pass}
		fromAddr = *from
		if fromAddr == "" { fromAddr = *user + "@maleon.run" }
	}

	if *to == "" {
		fmt.Println("Error: --to required")
		os.Exit(1)
	}

	e := &email.Email{From: fromAddr, To: *to, Subject: *subject, Body: *body, HTML: *html}

	if config.Username != "" && config.Password != "" {
		config.SendWithAuth(e.From, e.To, e.BuildMessage())
	} else {
		config.Send(e.To, e.BuildMessage())
	}
	fmt.Println("Email sent!")
}

func handleTest(args []string) {
	testCmd := flag.NewFlagSet("test", flag.ExitOnError)
	host := testCmd.String("host", "", "SMTP server")
	port := testCmd.Int("port", 0, "SMTP port")
	user := testCmd.String("user", "", "SMTP user")
	pass := testCmd.String("pass", "", "SMTP pass")
	profile := testCmd.String("profile", "", "Profile")
	testCmd.Parse(args)

	var config *smtp.Config

	if *profile != "" {
		profiles := loadProfiles()
		if p, ok := profiles[*profile]; ok {
			config = &smtp.Config{Host: p.Host, Port: p.Port, Username: p.User, Password: p.Pass}
			if *host != "" { config.Host = *host }
			if *port != 0 { config.Port = *port }
			if *user != "" { config.Username = *user }
			if *pass != "" { config.Password = *pass }
		}
	} else {
		hostVal := *host
		if hostVal == "" { hostVal = "smtp.maleon.run" }
		portVal := *port
		if portVal == 0 { portVal = 587 }
		config = &smtp.Config{Host: hostVal, Port: portVal, Username: *user, Password: *pass}
	}

	if err := config.Validate(); err != nil {
		fmt.Printf("Connection failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Connection successful!")
}

func getHelp() string {
	return `SMTP CLI - Email Client for Terminal

Usage:
  smtp-cli send [flags]              Send email
  smtp-cli test [flags]              Test SMTP connection
  smtp-cli login [flags]             Login to mail server
  smtp-cli inbox                     List inbox emails
  smtp-cli sent                      List sent emails
  smtp-cli read --id <n>             Read email by ID
  smtp-cli delete --id <n>           Delete email
  smtp-cli profile add [flags]       Add profile
  smtp-cli profile remove --name <n> Remove profile
  smtp-cli profiles                  List profiles
  smtp-cli help                      Show this help

Examples:
  # Add profile
  smtp-cli profile add --name trabajo --user micorreo --pass pass

  # Login with profile
  smtp-cli login --profile trabajo --email micorreo@maleon.run --pass pass

  # Check inbox
  smtp-cli inbox --profile trabajo

  # Check sent
  smtp-cli sent --profile trabajo

  # Read email
  smtp-cli read --id 1 --profile trabajo

  # Send email
  smtp-cli send -profile trabajo -to dest@email.com -subject "Hi" -body "Hello"

  # Test connection
  smtp-cli test -profile trabajo

Profiles:
  smtp-cli profile add --name <name> --user <user> --pass <pass>
  smtp-cli profile remove --name <name>
  smtp-cli profiles
`
}
