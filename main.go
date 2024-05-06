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

// Initialize database connection
func initDB() {
	// Define the PostgreSQL connection string
    connectionString := "postgresql://postgres:password@localhost:5433/filter_messages"

    // Establish a database connection using the PostgreSQL connection string
    conn, err := pgx.Connect(context.Background(), connectionString)
    if err != nil {
        log.Fatalf("Unable to connect to database: %v\n", err)
    }

    dbConn = conn
}

func main() {
	
    initDB() // Initialize database connection

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

				if filterWord == "" {

					response := fmt.Sprintf("You have to set a filter word first.")
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
					_, err := bot.Send(msg)
					if err != nil {
						log.Printf("Error sending message: %v", err)
					}

				} else {

					if strings.Contains(strings.ToLower(update.Message.Text), strings.ToLower(filterWord)) {
						saveMessageFilteredTable(bot, update.Message)
					} else {
						saveMessageNotFilteredTable(bot, update.Message)
					}

				}
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
		_, err := bot.Send(msg)
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
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

	msg := tgbotapi.NewMessage(message.Chat.ID, response)
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
	
}

func saveMessageFilteredTable(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	senderID := message.Chat.ID
	sendingDate := message.Time()
	messageID := message.MessageID
	messageText := message.Text

	// Save message into filtered_messages table
    _, err := dbConn.Exec(context.Background(),
        "INSERT INTO filtered_messages (sender_id, sending_date, message_id, message_contents, filter_word) VALUES ($1, $2, $3, $4, $5)",
        senderID, sendingDate, messageID, messageText, filterWord)
    if err != nil {
        log.Printf("Error saving message to filtered_messages table: %v\n", err)
        return
    }
	
	response := fmt.Sprintf("Message is saved to filtered table with keyword: %s", filterWord)
	msg := tgbotapi.NewMessage(message.Chat.ID, response)
	_, sendErr := bot.Send(msg)

	if sendErr != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func saveMessageNotFilteredTable(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	senderID := message.Chat.ID
	sendingDate := message.Time()
	messageID := message.MessageID
	messageText := message.Text

	// Save message into not_filtered_messages table
    _, err := dbConn.Exec(context.Background(),
        "INSERT INTO not_filtered_messages (sender_id, sending_date, message_id, message_contents) VALUES ($1, $2, $3, $4)",
        senderID, sendingDate, messageID, messageText)
    if err != nil {
        log.Printf("Error saving message to not_filtered_messages table: %v\n", err)
        return
    }

    response := "Message is saved to not filtered table"
    msg := tgbotapi.NewMessage(message.Chat.ID, response)
    _, sendErr := bot.Send(msg)
    if sendErr != nil {
        log.Printf("Error sending message: %v\n", err)
    }
}
