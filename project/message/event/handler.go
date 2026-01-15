package event

import (
	"tickets/db"
	"tickets/worker"
)



type Handler struct {
	spreadsheetsAPI worker.SpreadsheetsAPI
	receiptsService worker.ReceiptsService
	ticketRepository db.TicketRepository
}

func NewHandler(receiptService worker.ReceiptsService, spreadsheetAPI worker.SpreadsheetsAPI, repo db.TicketRepository) Handler {
	if receiptService == nil {
		panic("receipt service missing")
	}
	if spreadsheetAPI == nil {
		panic("spreadsheet api missing")
	}
	if repo == nil {
		panic("repo is missing")
	}

	return Handler{
		spreadsheetsAPI: spreadsheetAPI,
		receiptsService: receiptService,
		ticketRepository: repo,
	}
}