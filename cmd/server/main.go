package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/RyoheiKamo/go-slack-bot/internal/service"
)

type SlackEventRequest struct {
	Type      string     `json:"type"`
	Challenge string     `json:"challenge"`
	EventID   string     `json:"event_id"`
	Event     SlackEvent `json:"event"`
}

type SlackEvent struct {
	Type    string `json:"type"`
	User    string `json:"user"`
	Text    string `json:"text"`
	Channel string `json:"channel"`
	Ts      string `json:"ts"`
}

var (
	processedEvents = make(map[string]time.Time)
	processedMu     sync.Mutex
)

func isDuplicateEvent(eventID string) bool {
	if eventID == "" {
		return false
	}

	processedMu.Lock()
	defer processedMu.Unlock()

	now := time.Now()

	for id, processedAt := range processedEvents {
		if now.Sub(processedAt) > 10*time.Minute {
			delete(processedEvents, id)
		}
	}

	if _, exists := processedEvents[eventID]; exists {
		return true
	}

	processedEvents[eventID] = now

	return false
}

func handleAppMention(event SlackEvent) {
	log.Printf(
		"app_mention received: user=%s channel=%s text=%s",
		event.User,
		event.Channel,
		event.Text,
	)

	// Botのメンション部分を除去
	message := event.Text

	if index := strings.Index(message, ">"); index != -1 {
		message = strings.TrimSpace(message[index+1:])
	}

	log.Printf("user message: %s", message)

	// OpenAI API
	openAIAPIKey := os.Getenv("OPENAI_API_KEY")
	if openAIAPIKey == "" {
		log.Println("OPENAI_API_KEY is not set")
		return
	}

	openAIService := service.NewOpenAIService(openAIAPIKey)

	aiResponse, err := openAIService.GenerateResponse(message)
	if err != nil {
		log.Printf(
			"failed to generate OpenAI response: %v",
			err,
		)
		return
	}

	log.Printf("OpenAI response: %s", aiResponse)

	// Slackへ返信
	botToken := os.Getenv("SLACK_BOT_TOKEN")
	if botToken == "" {
		log.Println("SLACK_BOT_TOKEN is not set")
		return
	}

	messageService := service.NewSlackMessageService(botToken)

	if err := messageService.SendMessage(
		event.Channel,
		aiResponse,
		event.Ts,
	); err != nil {
		log.Printf(
			"failed to send Slack message: %v",
			err,
		)
		return
	}
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "go-slack-bot")
	})

	http.HandleFunc("/slack/events", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		timestamp := r.Header.Get("X-Slack-Request-Timestamp")
		signature := r.Header.Get("X-Slack-Signature")

		signingSecret := os.Getenv("SLACK_SIGNING_SECRET")
		if signingSecret == "" {
			log.Println("SLACK_SIGNING_SECRET is not set")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		signatureService := service.NewSlackSignatureService(signingSecret)

		if !signatureService.Verify(timestamp, signature, body) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req SlackEventRequest

		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		log.Printf(
			"Slack request received: type=%s event_type=%s body=%s",
			req.Type,
			req.Event.Type,
			string(body),
		)

		// Slack URL verification
		if req.Type == "url_verification" {
			w.Header().Set("Content-Type", "application/json")

			if err := json.NewEncoder(w).Encode(map[string]string{
				"challenge": req.Challenge,
			}); err != nil {
				log.Printf("failed to encode response: %v", err)
			}

			return
		}

		// Slack event
		if req.Type == "event_callback" {
			if isDuplicateEvent(req.EventID) {
				log.Printf("duplicate event ignored: event_id=%s", req.EventID)
				w.WriteHeader(http.StatusOK)
				return
			}

			if req.Event.Type == "app_mention" {
				go handleAppMention(req.Event)
			}

			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	log.Println("server started on :8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
