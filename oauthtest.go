package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type User struct {
	Id      string `json:id`
	Picture string
}

func main() {
	// Your credentials should be obtained from the Google
	// Developer Console (https://console.developers.google.com).
	conf := &oauth2.Config{
		ClientID:     os.GetEnv(""),
		ClientSecret: os.GetEnv(""),
		RedirectURL:  "http://localhost:8080/auth/google/callback",
		Scopes: []string{
			"openid",
		},
		Endpoint: google.Endpoint,
	}
	// Redirect user to Google's consent page to ask for permission
	// for the scopes specified above.
	url := conf.AuthCodeURL("state")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "<html><body><a href=\"%s\">Login With Google</a></body></html>", url)
	})

	var store = sessions.NewCookieStore([]byte("SUPER-SECURE-KEY"))

	http.HandleFunc("/auth/google/callback", func(w http.ResponseWriter, r *http.Request) {

		tok, err := conf.Exchange(oauth2.NoContext, r.URL.Query().Get("code"))
		if err != nil {
			log.Fatal(err)
		}
		client := conf.Client(oauth2.NoContext, tok)
		rsp, err := client.Get("https://www.googleapis.com/oauth2/v1/userinfo?alt=json")
		if err != nil {
			log.Fatal(err)
		}
		defer rsp.Body.Close()

		user := User{}
		err = json.NewDecoder(rsp.Body).Decode(&user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		session, _ := store.Get(r, "cookie-session")
		session.Values["ID"] = user.Id
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/profile", 302)
		//fmt.Fprintf(w, r.URL.Query().Get("code"))
	})

	http.HandleFunc("/profile", func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "cookie-session")
		id := session.Values["ID"]
		fmt.Fprintf(w, "saved: %s", id)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))

	/*
		// Handle the exchange code to initiate a transport.
		tok, err := conf.Exchange(oauth2.NoContext, authcode)
		if err != nil {
			log.Fatal(err)
		}
		client := conf.Client(oauth2.NoContext, tok)
		rsp, err := client.Get("https://www.googleapis.com/oauth2/v1/userinfo?alt=json")
		if err != nil {
			log.Fatal(err)
		}

		bodyBytes, err := ioutil.ReadAll(rsp.Body)
		// Error checking of the ioutil.ReadAll() request
		if err != nil {
			log.Fatal(err)
		}

		bodyString := string(bodyBytes)

		fmt.Printf("Response %s", bodyString)
	*/
}
