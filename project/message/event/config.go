package event

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)

var marshaler = cqrs.JSONMarshaler{
	GenerateName: cqrs.StructName,
}

func NewEventProcessorConfig(redisClient *redis.Client, logger watermill.LoggerAdapter) cqrs.EventProcessorConfig {
	return cqrs.EventProcessorConfig{
		GenerateSubscribeTopic: func(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
			return "events." + params.EventName, nil
		},
		Logger:    logger,
		Marshaler: marshaler,
		SubscriberConstructor: func(params cqrs.EventProcessorSubscriberConstructorParams) (message.Subscriber, error) {
			return redisstream.NewSubscriber(redisstream.SubscriberConfig{
				Client:        redisClient,
				ConsumerGroup: "events." + params.HandlerName,
			}, logger)
		},
	}
}

func NewCommandProcessorConfig(rc *redis.Client, logger watermill.LoggerAdapter) cqrs.CommandProcessorConfig {
	return cqrs.CommandProcessorConfig{
		GenerateSubscribeTopic: func(param cqrs.CommandProcessorGenerateSubscribeTopicParams) (string, error) {
			return "commands." + param.CommandName, nil
		},
		Logger: logger,
		Marshaler: marshaler,
		SubscriberConstructor: func(params cqrs.CommandProcessorSubscriberConstructorParams) (message.Subscriber, error) {
			return redisstream.NewSubscriber(redisstream.SubscriberConfig{
				Client: rc,
				ConsumerGroup: "commands." + params.HandlerName,
			}, logger)
		},
	}
}
