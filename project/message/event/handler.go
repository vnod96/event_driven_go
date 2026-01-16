package event

import (
	"tickets/db"
	"tickets/worker"
)



type Handler struct {
	spreadsheetsAPI worker.SpreadsheetsAPI
	receiptsService worker.ReceiptsService
	fileService worker.FileService
	ticketRepository db.TicketRepository
}

func NewHandler(receiptService worker.ReceiptsService, spreadsheetAPI worker.SpreadsheetsAPI, fileService worker.FileService, repo db.TicketRepository) Handler {
	if receiptService == nil {
		panic("receipt service missing")
	}
	if spreadsheetAPI == nil {
		panic("spreadsheet api missing")
	}
	if fileService == nil {
		panic("fileService is missing")
	}

	if repo == nil {
		panic("repo is missing")
	}

	return Handler{
		spreadsheetsAPI: spreadsheetAPI,
		receiptsService: receiptService,
		fileService: fileService,
		ticketRepository: repo,
	}
}