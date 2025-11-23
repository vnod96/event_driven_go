package message

import (
	"os"
	"tickets/worker"

	"github.com/redis/go-redis/v9"
)

type PubSubWorker struct {
	spreadsheetsAPI worker.SpreadsheetsAPI
	receiptService  worker.ReceiptsService
	rc              *redis.Client
}

func NewPubSubWorker(spreadsheetsAPI worker.SpreadsheetsAPI, receiptService worker.ReceiptsService) *PubSubWorker {
	rc := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	wrkr := &PubSubWorker{
		spreadsheetsAPI: spreadsheetsAPI,
		receiptService:  receiptService,
		rc:              rc,
	}

	return wrkr
}
