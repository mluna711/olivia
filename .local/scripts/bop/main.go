package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

var helpMessage = `
Usage:
  bop -command search -query "super shy"

Available commands:
  - search
`

func randomString() string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, 10)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

const PORT = 8888

func main() {
	app, err := initializeApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: \n%v", err)
		os.Exit(1)
	}

	redirectURL := fmt.Sprintf("http://localhost:%d/callback", PORT)
	redirectComps := strings.Split(redirectURL, "/")
	redirectPath := fmt.Sprintf("/%s", redirectComps[len(redirectComps)-1])

	state := randomString()
	auth := spotifyauth.New(spotifyauth.WithRedirectURL(redirectURL), spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate, spotifyauth.ScopeUserModifyPlaybackState, spotifyauth.ScopeUserReadCurrentlyPlaying), spotifyauth.WithClientID(app.clientId))
	url := auth.AuthURL(state)
	fmt.Printf("Authenticate using the following link: \n%s\n\n", url)

	router := http.NewServeMux()
	router.HandleFunc(redirectPath, func(w http.ResponseWriter, r *http.Request) {
		if app.client != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte("already authenticated"))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		token, err := auth.Token(context.Background(), state, r)
		if err != nil {
			http.Error(w, "Couldn't get token", http.StatusNotFound)
			return
		}
		// create a client using the specified token
		client := spotify.New(auth.Client(context.Background(), token))
		app.client = client

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte("user authenticated"))
		log.Println("user authenticated")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	router.HandleFunc("/search", logging(app.search))
	router.HandleFunc("/play", logging(app.playSong))
	router.HandleFunc("/pause", logging(app.pause))
	router.HandleFunc("/next", logging(app.next))
	router.HandleFunc("/prev", logging(app.prev))
	router.HandleFunc("/status", logging(app.status))
	router.HandleFunc("/restart", logging(app.restart))

	fmt.Println("Waiting for requests")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", PORT), router))
}
