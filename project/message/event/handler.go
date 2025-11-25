package event

import "tickets/worker"



type Handler struct {
	spreadsheetsAPI worker.SpreadsheetsAPI
	receiptsService worker.ReceiptsService
}

func NewHandler(receiptService worker.ReceiptsService, spreadsheetAPI worker.SpreadsheetsAPI) Handler {
	if receiptService == nil {
		panic("receipt service missing")
	}
	if spreadsheetAPI == nil {
		panic("spreadsheet api missing")
	}

	return Handler{
		spreadsheetsAPI: spreadsheetAPI,
		receiptsService: receiptService,
	}
}