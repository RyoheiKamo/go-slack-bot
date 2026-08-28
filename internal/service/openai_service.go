package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/RyoheiKamo/go-slack-bot/internal/prompt"
)

type OpenAIService struct {
	apiKey string
	client *http.Client
}

type openAIRequest struct {
	Model        string            `json:"model"`
	Instructions string            `json:"instructions"`
	Input        []openAIInputItem `json:"input"`
}

type openAIInputItem struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Output []openAIOutput `json:"output"`
}

type openAIOutput struct {
	Type    string          `json:"type"`
	Content []openAIContent `json:"content"`
}

type openAIContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func NewOpenAIService(apiKey string) *OpenAIService {
	return &OpenAIService{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

func (s *OpenAIService) GenerateResponse(message string) (string, error) {
	payload := openAIRequest{
		Model:        "gpt-5-mini",
		Instructions: prompt.SystemPrompt,
		Input: []openAIInputItem{
			{
				Role:    "user",
				Content: message,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		"https://api.openai.com/v1/responses",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("OpenAI API returned status %d", resp.StatusCode)
	}

	var result openAIResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	for _, output := range result.Output {
		for _, content := range output.Content {
			if content.Type == "output_text" && content.Text != "" {
				return content.Text, nil
			}
		}
	}

	return "", fmt.Errorf("OpenAI response did not contain output text")
}
