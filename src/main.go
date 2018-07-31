package main

import (
	"crypto/tls"
	"fmt"
	"html/template"
	"math"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"time"

	"github.com/BurntSushi/toml"
)

type Homepage struct {
	AppTime string
}

type Config struct {
	Server   string
	Email    string
	Password string
}

// This decodes the variables from config.toml for use in the program.
func LoadConfig() *Config {
	var config Config
	_, err := toml.DecodeFile("../config/config.toml", &config)
	if err != nil {
		panic(err.Error())
	}

	return &Config{
		Server:   config.Server,
		Email:    config.Email,
		Password: config.Password,
	}
}

func TLSDial(address string) (*tls.Conn, error) {
	// Necessary for SSL encryption
	return tls.Dial("tcp", address, nil)
}

// Constructs the email send to users, usually called by SendMail.
func MakeMessage(sender string, recipient string, subject string, body string) (message string) {
	content := make(map[string]string)
	content["From"] = sender
	content["To"] = recipient
	content["Subject"] = subject

	for k, v := range content {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	return message
}

// Using returned values of DB and Config, form and send email.
func SendMail() {
	// Get values from DB and config.
	currentReminders := QueryDB()
	config := LoadConfig()

	serverHost, _, _ := net.SplitHostPort(config.Server)
	conn, err := TLSDial(config.Server)
	if err != nil {
		fmt.Printf("Error sending mail: %s", err)
	}

	client, err := smtp.NewClient(conn, serverHost)
	if err != nil {
		fmt.Printf("Client error: %s", err)
	}

	// Set up auth necessary to send email. Uses values from config.
	auth := smtp.PlainAuth("", config.Email, config.Password, serverHost)
	authErr := client.Auth(auth)
	if authErr != nil {
		fmt.Printf("Auth error: %s\n", authErr)
	}

	if err != nil {
		fmt.Printf("Writer error: %s\n", err)
	}

	subject := "A new show is about to start!"

	from := mail.Address{"", config.Email}
	// In this for range, because k is just the auto-increment PK ID, it is not needed here.
	for _, v := range currentReminders {

		date := time.Now()
		format := "2006-01-02"

		systemDate, _ := time.Parse(format, v.ShowDate)

		diff := date.Sub(systemDate)

		if math.Abs(diff.Hours()/24) <= float64(v.ReminderInterval) {
			fmt.Println("Less then 7 days.")

			to := mail.Address{"", v.UserEmail}

			err = client.Mail(from.Address)
			if err != nil {
				fmt.Printf("From address error: %s", err)
			}

			err = client.Rcpt(to.Address)
			if err != nil {
				fmt.Printf("Rcpt error: %s\n", err)
			}

			writer, err := client.Data()
			if err != nil {
				fmt.Printf("Writer error: %s\n", err)
			}

			body := "Hey that show " + v.ShowName + " is about to start!"
			// Build the email to send to user.
			message := MakeMessage(config.Email, v.UserEmail, subject, body)
			_, err = writer.Write([]byte(message))
			if err != nil {
				fmt.Printf("Error sending mail: %s", err)

				fmt.Println("Successful email sent to: " + v.UserEmail)

			}
			writer.Close()
		}
		// Close writer in current loop so the next loop doesn't error out.

	}
	// Once loop is done, close smtp client.
	client.Quit()

}

// This http handler function displays when user hits submit button.
func submission(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Method:", r.Method+"\n")
	fmt.Println("REACHED SUBMISSION")
	if r.Method == "GET" {
		t, _ := template.ParseFiles("index.html")
		t.Execute(w, nil)
	} else {
		r.ParseForm()

		fmt.Fprintf(w, "New reminder for %s", r.Form["sname"])
		fmt.Fprintf(w, " has been created.")
	}
	AddToDatabase(r.Form)
}

// Default index handler, has all input forms.
func index(w http.ResponseWriter, r *http.Request) {

	now := time.Now()

	homePage := Homepage{
		AppTime: now.Format("3:04 PM"),
	}

	t, err := template.ParseFiles("./index.html")
	if err != nil {
		fmt.Printf("Error: %s", err)
	}

	err = t.Execute(w, homePage)
	if err != nil {
		fmt.Printf("Error: %s", err)
	}

}

func main() {
	// Run in new thread so SendMail doesn't halt web app submissions
	go SendMail()
	// Standard http stuff for handlers and port.
	http.HandleFunc("/", index)
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
	http.HandleFunc("/submission", submission)
	// Golang has to serve css folder manually for html to read it.

	http.ListenAndServe(":8000", nil)
}
