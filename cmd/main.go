package main

import (
	"fmt"
	"log"
	"os"

	_ "github.com/fi9ish/ethTgBot/pkg/config"
	"github.com/fi9ish/ethTgBot/pkg/controllers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	fmt.Println("how are we doing?")
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAMBOT_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// Use a goroutine to handle updates
	go func() {
		for update := range updates {
			if update.Message == nil {
				continue
			}

			if update.Message.IsCommand() {
				// Extract the command and arguments from the message
				command := update.Message.Command()
				args := update.Message.CommandArguments()

				// Do something with the command and arguments
				controllers.HandleCommand(bot, update, command, args)
			} else {
				// Get the message text
				message := update.Message.Text

				// Do something with the message
				controllers.HandleMessage(bot, update, message)
			}
		}
	}()

	// Keep the program running until interrupted
	select {}
}
