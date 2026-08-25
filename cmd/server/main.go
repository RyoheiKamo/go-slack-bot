package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type SlackEventRequest struct {
	Type      string `json:"type"`
	Challenge string `json:"challenge"`
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

		var req SlackEventRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		if req.Type == "url_verification" {
			w.Header().Set("Content-Type", "application/json")

			if err := json.NewEncoder(w).Encode(map[string]string{
				"challenge": req.Challenge,
			}); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}

			return
		}

		w.WriteHeader(http.StatusOK)
	})

	log.Println("server started on :8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
