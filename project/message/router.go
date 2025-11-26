package message

import (
	"encoding/json"
	"tickets/entities"
	"tickets/message/event"
	"tickets/worker"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)

func NewWatermillRouter(
	spreadsheetsAPI worker.SpreadsheetsAPI,
	receiptService worker.ReceiptsService,
	rc *redis.Client,
	logger watermill.LoggerAdapter,
) *message.Router {
	router := message.NewDefaultRouter(logger)

	handler := event.NewHandler(receiptService, spreadsheetsAPI)

	issueConsumer, err := redisstream.NewSubscriber(
		redisstream.SubscriberConfig{
			Client:        rc,
			ConsumerGroup: "issue-receipt-consumer-group",
		}, logger,
	)

	if err != nil {
		panic(err)
	}

	spreadsheetConsumer, err := redisstream.NewSubscriber(
		redisstream.SubscriberConfig{
			Client:        rc,
			ConsumerGroup: "append-to-tracker-consumer-group",
		}, logger,
	)

	if err != nil {
		panic(err)
	}

	ticketCancelledConsumer, err := redisstream.NewSubscriber(
		redisstream.SubscriberConfig{
			Client:        rc,
			ConsumerGroup: "ticket-canceled-consumer",
		}, logger,
	)

	if err != nil {
		panic(err)
	}

	router.AddConsumerHandler(
		"issue-receipt-handler",
		"TicketBookingConfirmed",
		issueConsumer,
		func(msg *message.Message) error {
			var event entities.TicketBookingConfirmed
			err := json.Unmarshal(msg.Payload, &event)
			if err != nil {
				return err
			}

			return handler.IssueReceipt(msg.Context(), event)
		},
	)

	router.AddConsumerHandler(
		"append-to-tracker-handler",
		"TicketBookingConfirmed",
		spreadsheetConsumer,
		func(msg *message.Message) error {
			var event entities.TicketBookingConfirmed
			err := json.Unmarshal(msg.Payload, &event)
			if err != nil {
				return err
			}
			return handler.AppendRow(msg.Context(), event)
		},
	)

	router.AddConsumerHandler(
		"ticket-canceled-handler",
		"TicketBookingCanceled",
		ticketCancelledConsumer,
		func(msg *message.Message) error {
			var event entities.TicketBookingCanceled
			err := json.Unmarshal(msg.Payload, &event)
			if err != nil {
				return err
			}
			return handler.RefundTicket(msg.Context(), event)
		},
	)


	return router
}