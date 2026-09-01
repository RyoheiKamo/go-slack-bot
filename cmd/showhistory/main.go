package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/RyoheiKamo/go-slack-bot/internal/service"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf(
			"usage: go run ./cmd/showhistory <channel> <thread_ts>",
		)
	}

	channel := os.Args[1]
	threadTs := os.Args[2]

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		log.Fatal("REDIS_ADDR is not set")
	}

	chatHistoryTTLString := os.Getenv("CHAT_HISTORY_TTL")
	if chatHistoryTTLString == "" {
		log.Fatal("CHAT_HISTORY_TTL is not set")
	}

	chatHistoryTTL, err := time.ParseDuration(chatHistoryTTLString)
	if err != nil {
		log.Fatalf("invalid CHAT_HISTORY_TTL: %v", err)
	}

	ctx := context.Background()

	chatHistoryService := service.NewChatHistoryService(redisAddr, chatHistoryTTL)

	history, err := chatHistoryService.GetHistory(
		ctx,
		channel,
		threadTs,
	)
	if err != nil {
		log.Fatalf("failed to get chat history: %v", err)
	}

	if len(history) == 0 {
		fmt.Println("No chat history found.")
		return
	}

	fmt.Printf(
		"Slack conversation\nchannel: %s\nthread_ts: %s\n\n",
		channel,
		threadTs,
	)

	for _, message := range history {
		switch message.Role {
		case "user":
			fmt.Printf("User:\n%s\n\n", message.Content)

		case "assistant":
			fmt.Printf("Assistant:\n%s\n\n", message.Content)

		default:
			fmt.Printf("%s:\n%s\n\n", message.Role, message.Content)
		}
	}
}
