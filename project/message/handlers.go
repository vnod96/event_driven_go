package message

import (
	"context"
	"tickets/worker"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/redis/go-redis/v9"
)

func NewHandlers(
	spreadsheetsAPI worker.SpreadsheetsAPI,
	receiptService  worker.ReceiptsService,
	rc              *redis.Client,
	logger          watermill.LoggerAdapter,
) {

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

	go func() {
		messages, err := issueConsumer.Subscribe(context.Background(), "issue-receipt")

		if err != nil {
			panic(err)
		}

		for msg := range messages {
			tktId := string(msg.Payload)
			err := receiptService.IssueReceipt(msg.Context(), tktId)
			if err != nil {
				logger.Error("failed to issue receipt", err, nil)
				msg.Nack()
			} else {
				msg.Ack()
			}
		}
	}()

	go func() {
		messages, err := spreadsheetConsumer.Subscribe(context.Background(), "append-to-tracker")

		if err != nil {
			panic(err)
		}

		for msg := range messages {
			tktId := string(msg.Payload)
			err := spreadsheetsAPI.AppendRow(msg.Context(), "tickets-to-print", []string{tktId})
			if err != nil {
				logger.Error("failed to append to tracker", err, nil)
				msg.Nack()
			} else {
				msg.Ack()
			}
		}
	}()

}