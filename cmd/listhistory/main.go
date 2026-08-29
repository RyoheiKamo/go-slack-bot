package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/redis/go-redis/v9"
)

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		log.Fatal("REDIS_ADDR is not set")
	}

	ctx := context.Background()

	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	keys, err := client.Keys(ctx, "slack:thread:*").Result()
	if err != nil {
		log.Fatalf("failed to get chat history keys: %v", err)
	}

	if len(keys) == 0 {
		fmt.Println("No chat histories found.")
		return
	}

	fmt.Printf("Found %d chat histories\n\n", len(keys))

	for _, key := range keys {
		channel, threadTs := parseKey(key)

		ttl, err := client.TTL(ctx, key).Result()
		if err != nil {
			log.Printf("failed to get TTL: key=%s err=%v", key, err)
			continue
		}

		fmt.Printf("Channel   : %s\n", channel)
		fmt.Printf("Thread TS : %s\n", threadTs)
		fmt.Printf("TTL       : %s\n", ttl)
		fmt.Printf("Key       : %s\n", key)
		fmt.Println()
	}
}

func parseKey(key string) (string, string) {
	const prefix = "slack:thread:"

	value := strings.TrimPrefix(key, prefix)

	parts := strings.SplitN(value, ":", 2)

	if len(parts) != 2 {
		return "", ""
	}

	return parts[0], parts[1]
}
