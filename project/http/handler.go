package http

import (
	"tickets/worker"

	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
)

type Handler struct {
	spreadsheetsAPI worker.SpreadsheetsAPI
	receiptsService worker.ReceiptsService
	pub *redisstream.Publisher
}
