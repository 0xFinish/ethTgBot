package controllers

import (
	"fmt"
	"log"

	"github.com/fi9ish/ethTgBot/pkg/gethfuncs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update, command string, args string) {
	// Do something with the command and arguments, such as processing them or storing them in a database
	response := switcher(command, args)
	fmt.Println(response)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
	_, err := bot.Send(msg)
	if err != nil {
		log.Fatal(err)
	}
}

func HandleMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, message string) {
	// Do something with the message, such as processing it or storing it in a database
}

func switcher(command string, args string) string {
	switch command {
	case "create":
		return gethfuncs.CreateNewWallet()
	case "getCurrentBlockNum":
		return gethfuncs.GetCurrentBlockNum()
	case "getGasSpent":
		return gethfuncs.GetGasSpent(args)
	case "getTransactionFee":
		return gethfuncs.GetTransactionFee(args)
	case "getBiggestGasSpender":
		return gethfuncs.GetBiggestGasSpender(args)
	case "getBiggestBlockWallet":
		return gethfuncs.GetBiggestBlockWallet(args)
	case "getAddressInfo":
		return gethfuncs.GetAddressInfo(args)
	default:
		return "Please choose the provided command"
	}

}