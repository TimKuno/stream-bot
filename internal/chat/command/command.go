// Package command implements the chat logic e.g. check messages for command usage.
package command

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/TimKuno/stream-bot/internal/config"
	"github.com/TimKuno/stream-bot/internal/token"
)

var commandFilePath string = "./configs/commands.json"
var commandsList []string
var textCommands map[string]string

// Represents a message with all required information to send a reply.
type Message struct {
	BroadcasterID   string `json:"broadcaster_id"`
	SenderID        string `json:"sender_id"`
	Message         string `json:"message"`
	ParentMessageID string `json:"reply_parent_message_id"`
}

// Checks a message for command usage.
func CheckMessage(message *Message) {
	for _, command := range commandsList {
		if strings.EqualFold(message.Message, command) {
			log.Printf("Command: Found command: %v", command)
			message.Message = textCommands[command]
			sendReplyMessage(message)
		}
	}
}

func sendReplyMessage(data *Message) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Connection: Could not serialize data: %v", err)
	}
	req, _ := http.NewRequest(
		"POST",
		"https://api.twitch.tv/helix/chat/messages",
		bytes.NewBuffer(jsonData),
	)

	req.Header.Set("Client-Id", config.Cfg.ClientID)
	req.Header.Set("Authorization", token.GetAccessToken())
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Connection: Unable to get stream user ID")
	}
	defer resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	log.Println(result)
}

type userData struct {
	Users []user `json:"data"`
}

type user struct {
	ID              string `json:"id"`
	Login           string `json:"login"`
	DisplayName     string `json:"display_name"`
	Type            string `json:"type"`
	BroadcasterType string `json:"broadcaster_type"`
}

// Returns the ID of the given username.
func GetUserID(username string) string {
	req, _ := http.NewRequest(
		"GET",
		"https://api.twitch.tv/helix/users?login="+username,
		nil,
	)

	req.Header.Set("Client-Id", config.Cfg.ClientID)
	req.Header.Set("Authorization", token.GetAccessToken())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Connection: Unable to get stream user ID")
	}
	defer resp.Body.Close()

	var result userData
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Users[0].ID
}

// Loads the text commands from the file system.
// In case the file can't be parsed or deserialized we start without the text command.
// Then only the moderation commands can be used.
func LoadCommands() {
	file, err := os.Open(commandFilePath)
	if err != nil {
		log.Println("Commands: Could not load the commands from file.")
		return
	}
	defer file.Close()
	err = json.NewDecoder(file).Decode(&textCommands)
	if err != nil {
		log.Println("Commands: Could not deserialize commands from file.")
		return
	}

	for key := range textCommands {
		commandsList = append(commandsList, key)
	}
}
