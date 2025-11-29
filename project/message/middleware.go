package message

import (
	"log/slog"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
)

func useMiddlewares(router *message.Router) {

	router.AddMiddleware(middleware.CorrelationID)

	router.AddMiddleware(func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			correlationId := middleware.MessageCorrelationID(msg)
			ctx := log.ToContext(msg.Context(), slog.With("correlation_id", correlationId))
			ctx = log.ContextWithCorrelationID(ctx, correlationId)
			msg.SetContext(ctx)
			return h(msg)
		}
	})
	router.AddMiddleware(LoggingMiddleware)
}


func LoggingMiddleware(next message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		logger := log.FromContext(msg.Context())
		logger = logger.With(
			"message_id", msg.UUID,
			"payload", string(msg.Payload),
			"metadata", msg.Metadata,
			"handler", message.HandlerNameFromCtx(msg.Context()),
		)
		logger.Info("Handling a message")
		msgs, err := next(msg)
		defer func() {
			if err != nil {
				logger.With(
					"message_id", msg.UUID,
					"error", err,
				).Error("Error while handling a message")
			}
		}()

		return msgs, err
	}
}