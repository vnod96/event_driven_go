package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/redis/go-redis/v9"
)

func main() {
	logger := watermill.NewSlogLogger(nil)
	rc := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	sub, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:        rc,
		ConsumerGroup: "consumers-1",
	}, logger)

	if err != nil {
		panic(err)
	}

	messages, err := sub.Subscribe(context.Background(), "progress")

	for msg := range messages {
		val := string(msg.Payload)
		fmt.Printf("Message ID: %s - %s", msg.UUID, val)
		msg.Ack()
	}

}
