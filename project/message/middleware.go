package message

import (
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	messagesProcessingDuration = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  "messages",
			Name:       "processing_duration_seconds",
			Help:       "The total time spent processing messages",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"topic", "handler"},
	)
	messagesProcessedCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "messages",
		Name:      "processed_total",
		Help:      "Total number of message processed",
	}, []string{"topic", "handler"})
	messagesFailedCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "messages",
		Name:      "failed_total",
		Help:      "Total number of message failed",
	}, []string{"topic", "handler"})
)

func useMiddlewares(router *message.Router, logger watermill.LoggerAdapter) {

	router.AddMiddleware(middleware.CorrelationID)

	router.AddMiddleware(middleware.Retry{
		MaxRetries:      10,
		InitialInterval: time.Millisecond * 100,
		MaxInterval:     time.Second,
		Multiplier:      2,
		Logger:          logger,
	}.Middleware)

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

	router.AddMiddleware(func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			start := time.Now()
			topic := message.SubscribeTopicFromCtx(msg.Context())
			handler := message.HandlerNameFromCtx(msg.Context())
			labels := prometheus.Labels{"topic": topic, "handler": handler}
			messagesProcessedCounter.With(labels).Inc()
			m, err := h(msg)
			if err != nil {
				 messagesFailedCounter.With(labels).Inc()
			}

			messagesProcessingDuration.With(labels).Observe(time.Since(start).Seconds())
			return m, err
		}
	})

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
