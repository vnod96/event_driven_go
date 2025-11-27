package message

import (
	"log/slog"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
)

func useMiddlewares(router *message.Router) {

	router.AddMiddleware(middleware.CorrelationID)
	router.AddMiddleware(LoggingMiddleware)
	router.AddMiddleware(func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			correlationId := middleware.MessageCorrelationID(msg)
			ctx := log.ContextWithCorrelationID(msg.Context(), correlationId)
			msg.SetContext(ctx)
			return h(msg)
		}
	})
}


func LoggingMiddleware(next message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		logger := slog.With(
			"message_id", msg.UUID,
			"payload", string(msg.Payload),
			"metadata", msg.Metadata,
			"handler", message.HandlerNameFromCtx(msg.Context()),
		)
		logger.Info("Handling a message")
		return next(msg)
	}
}