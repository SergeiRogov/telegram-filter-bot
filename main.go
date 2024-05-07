package main

import (
	"context"
	"log"
	"fmt"
	"strings"
	"github.com/jackc/pgx/v4"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var dbConn *pgx.Conn
var filterWord string

func main() {
    initDB()

	bot, err := tgbotapi.NewBotAPI("token")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			if update.Message.IsCommand() {
				handleCommand(bot, update.Message)
			} else {
				handleRegularMessage(bot, update.Message)
			}
		}
	}
}

// Initialize database connection
func initDB() {
	// Define the PostgreSQL connection string
    connectionString := "postgresql://postgres:pass@localhost:5433/filter_messages"

    // Establish a database connection using the PostgreSQL connection string
    conn, err := pgx.Connect(context.Background(), connectionString)
    if err != nil {
        log.Fatalf("Unable to connect to database: %v\n", err)
    }

    dbConn = conn
}

func handleCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	switch message.Command() {
	case "filter":
		handleFilterCommand(bot, message)
	case "start":
		response := "Welcome! This bot is designed to filter messages based on a specified keyword.\n\n" +
		"To use this bot:\n" +
		"- Type `/filter your_word` to set a filter keyword. After setting a filter, messages containing the specified keyword will be saved to the `filtered_messages` table.\n" +
		"- Messages that do not contain the filter keyword will be saved to the `not_filtered_messages` table.\n\n" +
		"Start by setting your filter with `/filter your_word` to begin filtering messages."
		sendMessage(bot, message.Chat.ID, response)
	default:
		response := fmt.Sprintf("/%s command is not supported here.", message.Command())
		sendMessage(bot, message.Chat.ID, response)
	}
}

func handleFilterCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	args := message.CommandArguments()
	argumentWord := strings.TrimSpace(args)
	
	var response string

	if argumentWord == "" {
		response = fmt.Sprintf("No filter word entered. Please provide a word after /filter.")
	} else {
		filterWord = argumentWord
		response = fmt.Sprintf("Filter keyword set to: %s", filterWord)
	}
	sendMessage(bot, message.Chat.ID, response)
}

func handleRegularMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message){
	if filterWord == "" {
		response := fmt.Sprintf("You have to set a filter word first.")
		sendMessage(bot, message.Chat.ID, response)
	} else {
		if strings.Contains(strings.ToLower(message.Text), strings.ToLower(filterWord)) {
			saveMessageToTable(bot, message, "filtered_messages")
		} else {
			saveMessageToTable(bot, message, "not_filtered_messages")
		}
	}
}

func sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func saveMessageToTable(bot *tgbotapi.BotAPI, message *tgbotapi.Message, tableName string) {
	senderID := message.Chat.ID
	sendingDate := message.Time()
	messageID := message.MessageID
	messageText := message.Text

	var query string
	var response string

	switch tableName {
	case "filtered_messages":
		query = "INSERT INTO filtered_messages (sender_id, sending_date, message_id, message_contents, filter_word) VALUES ($1, $2, $3, $4, $5)"
		response = fmt.Sprintf("Message is saved to filtered_messages table with keyword: %s", filterWord)
		// Save message into the filtered_messages table
		_, err := dbConn.Exec(context.Background(), query, senderID, sendingDate, messageID, messageText, filterWord)
		if err != nil {
			log.Printf("Error saving message to filtered_messages table: %v", err)
			response = "Failed to save message."
		}
	case "not_filtered_messages":
		query = "INSERT INTO not_filtered_messages (sender_id, sending_date, message_id, message_contents) VALUES ($1, $2, $3, $4)"
		response = "Message is saved to not_filtered_messages table"
		// Save message into the not_filtered_messages table
		_, err := dbConn.Exec(context.Background(), query, senderID, sendingDate, messageID, messageText)
		if err != nil {
			log.Printf("Error saving message to not_filtered_messages table: %v", err)
			response = "Failed to save message."
		}
	default:
		log.Printf("Invalid table name specified: %s", tableName)
		response = "Invalid table name specified."
	}

	sendMessage(bot, message.Chat.ID, response)
}
