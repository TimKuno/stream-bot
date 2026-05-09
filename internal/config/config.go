// Package config implements utility to load and export the config.
package config

import (
	"encoding/json"
	"log"
	"os"
)

// Config represents the mapping of the config file.
type Config struct {
	ClientID       string `json:"clientID"`
	ClientSecret   string `json:"clientSecret"`
	StreamUserName string `json:"streamUserName"`
	BotUserName    string `json:"botUserName"`
	RedirectURI    string `json:"redirectURI"`
}

// Cfg holds the loaded config
var Cfg Config
var cfgFilePath string = "./configs/config.json"

// Loads the config from the file system.
// Exits the program in case of error i.e. config file is missing.
func LoadConfig() {
	file, err := os.Open(cfgFilePath)
	if err != nil {
		log.Fatal("Config: Could not load the config. Exiting.")
	}
	decoder := json.NewDecoder(file)

	decoder.Decode(&Cfg)
}
