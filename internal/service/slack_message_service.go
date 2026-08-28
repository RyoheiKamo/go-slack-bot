package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type SlackMessageService struct {
	botToken string
	client   *http.Client
}

type slackPostMessageRequest struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

type slackPostMessageResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

func NewSlackMessageService(botToken string) *SlackMessageService {
	return &SlackMessageService{
		botToken: botToken,
		client:   &http.Client{},
	}
}

func (s *SlackMessageService) SendMessage(channel string, text string) error {
	payload := slackPostMessageRequest{
		Channel: channel,
		Text:    text,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		"https://slack.com/api/chat.postMessage",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+s.botToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result slackPostMessageResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if !result.OK {
		return fmt.Errorf("Slack API error: %s", result.Error)
	}

	return nil
}
