package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

func getClient(config *oauth2.Config) *http.Client {

	postBody, _ := json.Marshal(map[string]string{
		"client_id":     config.ClientID,
		"client_secret": config.ClientSecret,
		"refresh_token": os.Getenv("REFRESH_TOKEN"),
		"grant_type":    "refresh_token",
	})
	responseBody := bytes.NewBuffer(postBody)
	resp, err := http.Post(config.Endpoint.TokenURL, "application/json", responseBody)

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	tok := &oauth2.Token{}
	err = json.NewDecoder(resp.Body).Decode(tok)

	if err != nil {
		log.Fatal(err)
	}

	return config.Client(context.Background(), tok)

}

func sendMail(ctx context.Context, srv *gmail.Service, email ...string) error {
	header := make(map[string]string)

	header["To"] = strings.Join(email, ",")
	header["From"] = "me"
	header["Subject"] = "Test Subject"
	header["Content-Type"] = "text/html; charset=\"UTF-8\""

	var msg string
	for k, v := range header {
		msg += fmt.Sprintf("%s: %s\r\n", k, v)
	}

	msg += "\r\n" + "<a href=\"https://www.google.com\">Test email body</a>"

	gmsg := gmail.Message{
		Raw: base64.URLEncoding.EncodeToString([]byte(msg)),
	}

	_, err := srv.Users.Messages.Send("me", &gmsg).Do()

	return err

}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	ctx := context.Background()

	config := &oauth2.Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		RedirectURL:  os.Getenv("REDIRECT_URI"),
		Scopes:       []string{gmail.GmailSendScope},
		Endpoint: oauth2.Endpoint{
			AuthURL:  os.Getenv("AUTH_URI"),
			TokenURL: os.Getenv("TOKEN_URI"),
		},
	}

	client := getClient(config)

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

	s := []string{"emailshouldcome@here"}

	err = sendMail(ctx, srv, s...)

	if err != nil {
		log.Fatalf("Unable to send mail: %v", err)
	}

}
