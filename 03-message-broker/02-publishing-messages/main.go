package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)

func main() {
	logger := watermill.NewSlogLogger(nil)

	rc := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	pub, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client: rc,
	}, logger)

	if err != nil {
		fmt.Printf("failed to start publisher: %v", err)
	}

	sub, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client: rc,
	}, logger)

	if err != nil {
		fmt.Printf("failed to start subscriber: %v", err)
	}

	messages, err := sub.Subscribe(context.Background(), "progress")

	if err != nil {
		fmt.Printf("failed to subscribe: %v", err)
	}

	logger.Info("subscriber started.", nil)
	go func() {

		for msg := range messages {
			val := string(msg.Payload)
			fmt.Printf("Message ID: %s - %s\n", msg.UUID, val)
			msg.Ack()
		}

	err = pub.Publish("progress", message.NewMessage(watermill.NewUUID(), []byte("50")))
	if err != nil {
		fmt.Printf("failed to publish msg: %v", err)
	}
	err = pub.Publish("progress", message.NewMessage(watermill.NewUUID(), []byte("100")))
	if err != nil {
		fmt.Printf("failed to publish msg: %v", err)
	}
}
