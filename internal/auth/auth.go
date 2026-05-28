// Package auth implements auth logic.
package auth

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/TimKuno/stream-bot/internal/config"
	"github.com/TimKuno/stream-bot/internal/token"
)

var server *http.Server

// Runs the auth logic.
func HandleAuth() {
	getOauth2Token()
}

func getOauth2Token() {
	mux := http.NewServeMux()
	server = &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	mux.HandleFunc("/", handleLogin)
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		go handleCallback(w, r)
	})

	log.Println("Auth: Please connect on http://localhost:8080 to generate a new Oauth2 token.")
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("Auth: Server failed to start: %v", err)
	}
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	authURL := "https://id.twitch.tv/oauth2/authorize"

	params := url.Values{}
	params.Add("client_id", config.Cfg.ClientID)
	params.Add("redirect_uri", config.Cfg.RedirectURI)
	params.Add("response_type", "code")
	params.Add("scope", "user:read:chat user:write:chat")
	http.Redirect(w, r, authURL+"?"+params.Encode(), http.StatusFound)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "No Code received", http.StatusBadRequest)
		return
	}

	tokenURL := "https://id.twitch.tv/oauth2/token"

	data := url.Values{}
	data.Set("client_id", config.Cfg.ClientID)
	data.Set("client_secret", config.Cfg.ClientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", config.Cfg.RedirectURI)

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		http.Error(w, "Token Request failed.", 500)
		return
	}
	defer resp.Body.Close()

	var oauth2token token.Token
	json.NewDecoder(resp.Body).Decode(&oauth2token)

	log.Println("Auth: New Token generated.")

	token.SaveToken(oauth2token)

	// Graceful shutdown server after new oAuth token is received
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Auth: Graceful shutdown failed: %v", err)
		server.Close()
	}
}
