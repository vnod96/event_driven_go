package message

import (
	"context"
	"os"
	"tickets/worker"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/redis/go-redis/v9"
)

type PubSubWorker struct {
	spreadsheetsAPI worker.SpreadsheetsAPI
	receiptService  worker.ReceiptsService
	rc              *redis.Client
	logger          watermill.LoggerAdapter
	Pub             *redisstream.Publisher
}

func NewPubSubWorker(spreadsheetsAPI worker.SpreadsheetsAPI, receiptService worker.ReceiptsService) *PubSubWorker {
	logger := watermill.NewSlogLogger(nil)
	rc := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	pub, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client: rc,
	},
		logger,
	)

	if err != nil {
		panic(err)
	}
	wrkr := &PubSubWorker{
		spreadsheetsAPI: spreadsheetsAPI,
		receiptService:  receiptService,
		rc:              rc,
		Pub:             pub,
		logger:          logger,
	}

	return wrkr
}

func (w *PubSubWorker) Run() error {
	issueConsumer, err := redisstream.NewSubscriber(
		redisstream.SubscriberConfig{
			Client:        w.rc,
			ConsumerGroup: "issue-receipt-consumer-group",
		}, w.logger,
	)

	if err != nil {
		return err
	}

	spreadsheetConsumer, err := redisstream.NewSubscriber(
		redisstream.SubscriberConfig{
			Client:        w.rc,
			ConsumerGroup: "append-to-tracker-consumer-group",
		}, w.logger,
	)

	if err != nil {
		return err
	}

	go func() {
		messages, err := issueConsumer.Subscribe(context.Background(), "issue-receipt")

		if err != nil {
			panic(err)
		}

		for msg := range messages {
			tktId := string(msg.Payload)
			err := w.receiptService.IssueReceipt(msg.Context(), tktId)
			if err != nil {
				w.logger.Error("failed to issue receipt", err, nil)
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
			err := w.spreadsheetsAPI.AppendRow(msg.Context(), "tickets-to-print", []string{tktId})
			if err != nil {
				w.logger.Error("failed to append to tracker", err, nil)
				msg.Nack()
			} else {
				msg.Ack()
			}
		}
	}()

	return nil

}
