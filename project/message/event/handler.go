package event

import (
	"tickets/db"
	"tickets/worker"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)



type Handler struct {
	spreadsheetsAPI worker.SpreadsheetsAPI
	receiptsService worker.ReceiptsService
	fileService worker.FileService
	ticketRepository db.TicketRepository
	eventBus *cqrs.EventBus
}

func NewHandler(receiptService worker.ReceiptsService, spreadsheetAPI worker.SpreadsheetsAPI, fileService worker.FileService, repo db.TicketRepository,
	eventBus *cqrs.EventBus) Handler {
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
	if eventBus == nil {
		panic("eventBus is missing")
	}

	return Handler{
		spreadsheetsAPI: spreadsheetAPI,
		receiptsService: receiptService,
		fileService: fileService,
		ticketRepository: repo,
		eventBus: eventBus,
	}
}