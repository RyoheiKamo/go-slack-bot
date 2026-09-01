package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatHistoryService struct {
	client *redis.Client
	ttl    time.Duration
}

func NewChatHistoryService(
	addr string,
	ttl time.Duration,
) *ChatHistoryService {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	return &ChatHistoryService{
		client: client,
		ttl:    ttl,
	}
}

func (s *ChatHistoryService) GetHistory(
	ctx context.Context,
	channel string,
	threadTs string,
) ([]ChatMessage, error) {
	key := buildChatHistoryKey(channel, threadTs)

	value, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return []ChatMessage{}, nil
	}
	if err != nil {
		return nil, err
	}

	var history []ChatMessage

	if err := json.Unmarshal([]byte(value), &history); err != nil {
		return nil, err
	}

	return history, nil
}

func (s *ChatHistoryService) SaveHistory(
	ctx context.Context,
	channel string,
	threadTs string,
	history []ChatMessage,
) error {
	key := buildChatHistoryKey(channel, threadTs)

	value, err := json.Marshal(history)
	if err != nil {
		return err
	}

	return s.client.Set(
		ctx,
		key,
		value,
		s.ttl,
	).Err()
}

func buildChatHistoryKey(channel string, threadTs string) string {
	return fmt.Sprintf(
		"slack:thread:%s:%s",
		channel,
		threadTs,
	)
}
