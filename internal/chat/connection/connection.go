// Package connection implements the chat connection logic.
package connection

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/TimKuno/stream-bot/internal/chat/command"
	"github.com/TimKuno/stream-bot/internal/config"
	"github.com/TimKuno/stream-bot/internal/token"
	"github.com/gorilla/websocket"
)

var streamChannelID string
var botUserID string

// Handles the chat connection.
func HandleChatConnection() {
	streamChannelID = command.GetUserID(config.Cfg.StreamUserName)
	botUserID = command.GetUserID(config.Cfg.BotUserName)
	subscribeWebsocket()
}

// https://dev.twitch.tv/docs/eventsub/websocket-reference
type response struct {
	Metadata metadata `json:"metadata"`
	Payload  payload  `json:"payload"`
}

type metadata struct {
	MessageID        string `json:"message_id"`
	MessageType      string `json:"message_type"`
	MessageTimestamp string `json:"message_timestamp"`
	SubscriptionType string `json:"subscription_type"`
}

type payload struct {
	Session session `json:"session"`
	Event   event   `json:"event"`
}

type session struct {
	ID                      string `json:"id"`
	Status                  string `json:"status"`
	ConnectedAt             string `json:"connected_at"`
	KeepaliveTimeoutSeconds int    `json:"keepalive_timeout_seconds"`
	ReconnectUrl            string `json:"reconnect_url"`
	RecoveryUrl             string `json:"recovery_url"`
}

// https://dev.twitch.tv/docs/eventsub/eventsub-subscription-types
type event struct {
	BroadcasterUserID string `json:"broadcaster_user_id"`
	ChatterUserID     string `json:"chatter_user_id"`
	ChatterUserName   string `json:"chatter_user_name"`
	MessageID         string `json:"message_id"`
	Message           message
}

type message struct {
	Text string `json:"text"`
}

var messageCache response

func subscribeWebsocket() {
	conn, _, err := websocket.DefaultDialer.Dial(
		"wss://eventsub.wss.twitch.tv/ws", nil,
	)
	if err != nil {
		log.Fatalf("Connection: Websocket connection failed: %v", err)
	}
	defer conn.Close()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Connection: Could not read message: %v", err)
		}

		var result response
		err = json.Unmarshal(msg, &result)
		if err != nil {
			log.Printf("Connection: Could not deserialize message: %v", err)
		}

		switch result.Metadata.MessageType {
		case "session_welcome":
			registerEventSubListeners(result.Payload.Session.ID)
		case "session_keepalive":
			// Is handled automatically
		case "notification":
			switch result.Metadata.SubscriptionType {
			case "channel.chat.message":
				messageCache = result
				go command.CheckMessage(&command.Message{
					BroadcasterID:   streamChannelID,
					SenderID:        botUserID,
					Message:         result.Payload.Event.Message.Text,
					ParentMessageID: result.Payload.Event.MessageID,
				})
			}
		default:
			log.Printf("Connection: Unexpected message type: %v", result.Metadata.MessageType)
		}
	}
}

type condition struct {
	BroadcasterUserID string `json:"broadcaster_user_id"`
	UserID            string `json:"user_id"`
}

type transport struct {
	Method    string `json:"method"`
	SessionID string `json:"session_id"`
}

type eventSubListenerPaylod struct {
	MessageType string    `json:"type"`
	Version     string    `json:"version"`
	Condition   condition `json:"condition"`
	Transport   transport `json:"transport"`
}

func registerEventSubListeners(sessionID string) {
	data := eventSubListenerPaylod{
		MessageType: "channel.chat.message",
		Version:     "1",
		Condition: condition{
			BroadcasterUserID: streamChannelID,
			UserID:            botUserID,
		},
		Transport: transport{
			Method:    "websocket",
			SessionID: sessionID,
		},
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Connection: Could not serialize data: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.twitch.tv/helix/eventsub/subscriptions", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Connection: Could not send request: %v", err)
	}
	req.Header.Set("Authorization", token.GetAccessToken())
	req.Header.Set("Client-Id", config.Cfg.ClientID)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "go-eventsub-client/1.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Connection: Could not subscribe eventsub: %v", err)
	}
	defer resp.Body.Close()
}
