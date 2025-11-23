package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/redis/go-redis/v9"
)

func main() {
	logger := watermill.NewSlogLogger(nil)

	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	sub, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client: rdb,
	}, logger)
	if err != nil {
		panic(err)
	}

	router := message.NewDefaultRouter(logger)

	router.AddConsumerHandler("feherheit-router", "temperature-fahrenheit", sub, func(msg *message.Message) error {
		val := string(msg.Payload)
		fmt.Printf("Temperature read: %s\n", val)
		return nil
	})

	err = router.Run(context.Background())
	if err != nil {
		panic(err)
	}

}
