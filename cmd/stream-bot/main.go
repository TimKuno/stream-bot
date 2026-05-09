// Package main runs the app.
package main

import (
	"log"

	"github.com/TimKuno/stream-bot/internal/bot"
)

func main() {
	log.Println("Starting the bot ...")
	bot.RunBot()
}
