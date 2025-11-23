package http

import "tickets/worker"

type Handler struct {
	spreadsheetsAPI worker.SpreadsheetsAPI
	receiptsService worker.ReceiptsService
	w *worker.Worker
}
