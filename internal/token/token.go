// Package token implements utility for the token handling i.e. storing, loading and refreshing.
package token

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/TimKuno/stream-bot/internal/config"
)

// Token represents the token structure.
type Token struct {
	AccessToken  string   `json:"access_token"`
	Expiration   int      `json:"expires_in"`
	RefreshToken string   `json:"refresh_token"`
	Scope        []string `json:"scope"`
	TokenType    string   `json:"token_type"`
}

var token Token
var tokenFilePath string = "./configs/token.json"

// Runs the token logic.
func ManageToken() {
	periodicTokenRefresh()
}

func refreshToken() {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", token.RefreshToken)
	data.Set("client_id", config.Cfg.ClientID)
	data.Set("client_secret", config.Cfg.ClientSecret)

	resp, err := http.PostForm("https://id.twitch.tv/oauth2/token", data)
	if err != nil {
		log.Fatal("Token: No new access token received.")
	}
	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&token)
	SaveToken(token)
	log.Println("Token: New Token generated.")
}

// Saves the given oauth2 token.
func SaveToken(oauth2token Token) {
	file, err := os.Create(tokenFilePath)
	if err != nil {
		log.Println("Token: Could not load token from file.")
	}
	defer file.Close()
	json.NewEncoder(file).Encode(oauth2token)
	token = oauth2token
	periodicTokenRefresh()
}

func periodicTokenRefresh() {
	time.AfterFunc(time.Duration(token.Expiration)*time.Second, refreshToken)
}

// Loads the token from file. Returns error in case of failure.
func LoadToken() error {
	file, err := os.Open(tokenFilePath)
	if err != nil {
		log.Println("Token: Could not open file.")
		return errors.New("File not found.")
	}
	defer file.Close()
	decodeErr := json.NewDecoder(file).Decode(&token)
	if decodeErr != nil {
		return errors.New("Token: Could not decode token from file.")
	}
	return nil
}

// Returns the access token i.e. 'Bearer xyz123'
func GetAccessToken() string {
	return token.TokenType + " " + token.AccessToken
}
