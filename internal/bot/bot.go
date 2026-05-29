// Package bot implements the logic for the stream-bot.
package bot

import (
	"log"
	"time"

	"github.com/TimKuno/stream-bot/internal/auth"
	"github.com/TimKuno/stream-bot/internal/chat/command"
	"github.com/TimKuno/stream-bot/internal/chat/connection"
	"github.com/TimKuno/stream-bot/internal/config"
	"github.com/TimKuno/stream-bot/internal/token"
)

// Runs the bot logic.
func RunBot() {
	config.LoadConfig()
	err := token.LoadToken()
	if err != nil {
		auth.HandleAuth()
	}
	go token.ManageToken()
	time.Sleep(1 * time.Second) // Artificial delay to prevent main thread being faster than initial goroutine execution

	command.LoadCommands()
	log.Println("Bot: Start Successful.")
	connection.HandleChatConnection()
	log.Println("Bot: Shutdown.")
}
