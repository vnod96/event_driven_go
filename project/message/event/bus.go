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
			return params.EventName, nil
		},
		Marshaler: marshaler,
		Logger: logger,
	})
	if err != nil {
		panic(err)
	}
	return eb
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
