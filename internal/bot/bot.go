// Package bot implements the logic for the stream-bot.
package bot

import (
	"log"

	"github.com/TimKuno/stream-bot/internal/auth"
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
	log.Println("Bot: Start Successful.")
}
