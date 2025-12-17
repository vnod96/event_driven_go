package message

import (
	"tickets/message/event"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)

func NewRedisClient(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: addr,
	})
}

func NewRedisPublisher(rc *redis.Client, logger watermill.LoggerAdapter) message.Publisher {
	var pub message.Publisher
	pub, err :=  redisstream.NewPublisher(
		redisstream.PublisherConfig{
			Client: rc,
		}, logger,
	)
	if err != nil {
		panic(err)
	}
	return event.CorrelationPublisherDecorator{pub}
}
