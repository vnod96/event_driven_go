package message

import (
	"context"
	"tickets/worker"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)

func NewHandlers(
	spreadsheetsAPI worker.SpreadsheetsAPI,
	receiptService worker.ReceiptsService,
	rc *redis.Client,
	logger watermill.LoggerAdapter,
) {

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
		"issue-receipt",
		issueConsumer,
		func(msg *message.Message) error {
			tktId := string(msg.Payload)
			return receiptService.IssueReceipt(msg.Context(), tktId)
		},
	)

	router.AddConsumerHandler(
		"append-to-tracker-handler",
		"append-to-tracker",
		spreadsheetConsumer,
		func(msg *message.Message) error {
			tktId := string(msg.Payload)
			return spreadsheetsAPI.AppendRow(msg.Context(), "tickets-to-print", []string{tktId})
		},
	)

	go func() {
		err := router.Run(context.Background())
		if err != nil {
			panic(err)
		}
	}()

}
