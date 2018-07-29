package main

import (
	"crypto/tls"
	"fmt"
	"html/template"
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

func LoadConfig() *Config {
	var config Config
	_, err := toml.DecodeFile("config.toml", &config)
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

func SendMail() {
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
	for k, v := range currentReminders {
		to := mail.Address{"", k}

		err = client.Mail(from.Address)
		if err != nil {
			panic(err.Error())
		}

		err = client.Rcpt(to.Address)
		if err != nil {
			panic(err.Error())
		}

		writer, err := client.Data()

		body := "Hey that show " + v.ShowName + " is about to start!"
		message := MakeMessage(config.Email, k, subject, body)
		_, err = writer.Write([]byte(message))
		if err != nil {
			fmt.Printf("Error sending mail: %s", err)
		}
		client.Quit()
	}

}

func submission(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Method:", r.Method+"\n")

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

func handler(w http.ResponseWriter, r *http.Request) {

	now := time.Now()

	homePage := Homepage{
		AppTime: now.Format("3:04 PM"),
	}

	t, err := template.ParseFiles("index.html")
	if err != nil {
		fmt.Printf("Error: %s", err)
	}

	err = t.Execute(w, homePage)
	if err != nil {
		fmt.Printf("Error: %s", err)
	}

}

func main() {
	SendMail()
	http.HandleFunc("/", handler)
	http.HandleFunc("/submission", submission)
	http.ListenAndServe(":8000", nil)
}
