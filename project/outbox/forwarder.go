package outbox

import (
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
)

func AddForwarder(
	pgSub message.Subscriber,
	rdPub message.Publisher,
	router *message.Router,
	logger watermill.LoggerAdapter,
) {
	_, err := forwarder.NewForwarder(
		pgSub,
		rdPub,
		logger,
		forwarder.Config{
			ForwarderTopic: outboxTopic,
			Router: router,
			Middlewares: []message.HandlerMiddleware{
				func(h message.HandlerFunc) message.HandlerFunc {
					return func(msg *message.Message) ([]*message.Message, error) {
						log.FromContext(msg.Context()).With(
							"message_id", msg.UUID,
							"payload", string(msg.Payload),
							"metadata", msg.Metadata,
						).Info("forwarding message")

						return h(msg)
					}
				},
			},
		},
	)

	if err != nil {
		panic(err)
	}
}