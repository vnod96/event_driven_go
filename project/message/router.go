package message

import (
	"encoding/json"
	"tickets/entities"
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

	router.AddConsumerHandler(
		"issue-receipt-handler",
		"TicketBookingConfirmed",
		issueConsumer,
		func(msg *message.Message) error {
			var event entities.IssueReceiptPayload
			err := json.Unmarshal(msg.Payload, &event)
			if err != nil {
				return err
			}

			return receiptService.IssueReceipt(msg.Context(), event)
		},
	)

	router.AddConsumerHandler(
		"append-to-tracker-handler",
		"TicketBookingConfirmed",
		spreadsheetConsumer,
		func(msg *message.Message) error {
			var event entities.AppendToTrackerPayload
			err := json.Unmarshal(msg.Payload, &event)
			if err != nil {
				return err
			}
			 return spreadsheetsAPI.AppendRow(msg.Context(), "tickets-to-print", []string{event.TicketID, event.CustomerEmail, event.Price.Amount, event.Price.Currency})
		},
	)


	return router
}