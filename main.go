package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

var clientID string
var clientSecret string

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	clientID = os.Getenv("CLIENT_ID")
	clientSecret = os.Getenv("CLIENT_SECRET")

	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", fs)

	// Set a http client to make external http requests.
	httpClient := http.Client{}

	http.HandleFunc("/oauth/redirect", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()

		if err != nil {
			fmt.Fprintf(os.Stdout, "Could not parse query: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		code := r.FormValue("code")

		// Call the Github oauth endpoint to get an access token
		reqURL := fmt.Sprintf("https://github.com/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s", clientID, clientSecret, code)
		req, err := http.NewRequest(http.MethodPost, reqURL, nil)
		if err != nil {
			fmt.Fprintf(os.Stdout, "Cound not create http request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Set the header as we want a json response
		req.Header.Set("accept", "application/json")

		// Send the request
		res, err := httpClient.Do(req)
		if err != nil {
			fmt.Fprintf(os.Stdout, "could not send HTTP request: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		defer res.Body.Close()

		// Parse the request body into the 'OAuthAccessResponse' struct
		var t OAuthAccessResponse
		if err := json.NewDecoder(res.Body).Decode(&t); err != nil {
			fmt.Fprintf(os.Stdout, "could not parse JSON response: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Finally send a response to redirect the user to the "welcome" page with the access token.
		w.Header().Set("Location", "/welcome.html?access_token="+t.AccessToken)
		w.WriteHeader(http.StatusFound)
	})

	http.ListenAndServe(":8080", nil)
}

type OAuthAccessResponse struct {
	AccessToken string `json:"access_token"`
}
