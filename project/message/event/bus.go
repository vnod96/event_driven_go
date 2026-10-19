package event

import (
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
)

func NewEventBus(pub message.Publisher, logger watermill.LoggerAdapter) *cqrs.EventBus {
	eb, err := cqrs.NewEventBusWithConfig(pub, cqrs.EventBusConfig{
		GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
			return "events." + params.EventName, nil
		},
		Marshaler: marshaler,
		Logger:    logger,
	})
	if err != nil {
		panic(err)
	}
	return eb
}

func NewCommandBus(pub message.Publisher, logger watermill.LoggerAdapter) *cqrs.CommandBus {
	cb, err := cqrs.NewCommandBusWithConfig(pub, cqrs.CommandBusConfig{
		GeneratePublishTopic: func(params cqrs.CommandBusGeneratePublishTopicParams) (string, error) {
			return "commands." + params.CommandName, nil
		},
		Marshaler: marshaler,
		Logger:    logger,
	})
	if err != nil {
		panic(err)
	}
	return cb
}

type CorrelationPublisherDecorator struct {
	message.Publisher
}

func (c CorrelationPublisherDecorator) Publish(topic string, messages ...*message.Message) error {
	for i := range messages {
		cId := log.CorrelationIDFromContext(messages[i].Context())
		messages[i].Metadata.Set("correlation_id", cId)
	}

	return c.Publisher.Publish(topic, messages...)
}
