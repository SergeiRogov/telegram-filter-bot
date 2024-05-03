package main

import (
	"log"
	"fmt"
	"strings"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var filterWord string

func main() {
	bot, err := tgbotapi.NewBotAPI("6972013736:AAHc-yMReJWPesgEVSeSU890WKD8OH19vQw")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // If we got a message
			if update.Message.IsCommand() {
				// Handle commands sent by users
				handleCommand(bot, update.Message)
			} else {
				// Check if the message contains the filter keyword
				if filterWord != "" && strings.Contains(strings.ToLower(update.Message.Text), strings.ToLower(filterWord)) {
					// Save the filtered message to the database
					saveMessageToDatabase(bot, update.Message)
				}

				// Echo back the received message
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
				msg.ReplyToMessageID = update.Message.MessageID
				bot.Send(msg)
			}
		}
	}
}

func handleCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	switch message.Command() {
	case "filter":
		handleFilterCommand(bot, message)
	default:
		response := fmt.Sprintf("%s command is not supported here.", message.Command())
		msg := tgbotapi.NewMessage(message.Chat.ID, response)
		bot.Send(msg)
	}
}

func handleFilterCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	// Extract the filter keyword from the command
	args := message.CommandArguments()
	argumentWord := strings.TrimSpace(args)
	
	var response string

	if argumentWord == "" {
		response = fmt.Sprintf("No filter word entered. Please provide a word after /filter.")
	} else {
		filterWord = argumentWord
		// Respond to the user with a confirmation message
		response = fmt.Sprintf("Filter keyword set to: %s", filterWord)
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, response)
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
	
}

func saveMessageToDatabase(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	
	// todo
	response := fmt.Sprintf("Message is saved to database with keyword: %s", filterWord)
	msg := tgbotapi.NewMessage(message.Chat.ID, response)
	bot.Send(msg)
}
