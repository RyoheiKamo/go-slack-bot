package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/RyoheiKamo/go-slack-bot/internal/service"
)

type SlackEventRequest struct {
	Type      string     `json:"type"`
	Challenge string     `json:"challenge"`
	Event     SlackEvent `json:"event"`
}

type SlackEvent struct {
	Type    string `json:"type"`
	User    string `json:"user"`
	Text    string `json:"text"`
	Channel string `json:"channel"`
	Ts      string `json:"ts"`
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

		if req.Type == "url_verification" {
			w.Header().Set("Content-Type", "application/json")

			if err := json.NewEncoder(w).Encode(map[string]string{
				"challenge": req.Challenge,
			}); err != nil {
				log.Printf("failed to encode response: %v", err)
			}

			return
		}

		if req.Type == "event_callback" {
			if req.Event.Type == "app_mention" {
				log.Printf(
					"app_mention received: user=%s channel=%s text=%s",
					req.Event.User,
					req.Event.Channel,
					req.Event.Text,
				)
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
