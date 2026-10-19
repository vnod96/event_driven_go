package event

import (
	"tickets/db"
	"tickets/worker"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type Handler struct {
	spreadsheetsAPI  worker.SpreadsheetsAPI
	receiptsService  worker.ReceiptsService
	fileService      worker.FileService
	ticketRepository db.TicketRepository
	showRepository   db.ShowRepository
	deadNationClient worker.DeadNationService
	eventBus         *cqrs.EventBus
	paymentService   worker.PaymentService
}

func NewHandler(receiptService worker.ReceiptsService, spreadsheetAPI worker.SpreadsheetsAPI, fileService worker.FileService, repo db.TicketRepository,
	showRepo db.ShowRepository, eventBus *cqrs.EventBus, deadNation worker.DeadNationService, paymentService worker.PaymentService) Handler {
	if receiptService == nil {
		panic("receipt service missing")
	}
	if spreadsheetAPI == nil {
		panic("spreadsheet api missing")
	}
	if fileService == nil {
		panic("fileService is missing")
	}

	if paymentService == nil {
		panic("paymentService is missing")
	}

	if repo == nil {
		panic("repo is missing")
	}
	if showRepo == nil {
		panic("show repo is missing")
	}
	if eventBus == nil {
		panic("eventBus is missing")
	}

	return Handler{
		spreadsheetsAPI:  spreadsheetAPI,
		receiptsService:  receiptService,
		fileService:      fileService,
		ticketRepository: repo,
		eventBus:         eventBus,
		showRepository:   showRepo,
		deadNationClient: deadNation,
		paymentService: paymentService,
	}
}
