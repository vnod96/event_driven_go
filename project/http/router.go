package http

import (
	"tickets/worker"

	libHttp "github.com/ThreeDotsLabs/go-event-driven/v2/common/http"
	"github.com/labstack/echo/v4"
)

func NewHttpRouter(
	spreadsheetsAPI worker.SpreadsheetsAPI,
	receiptsService worker.ReceiptsService,
	w *worker.Worker,
) *echo.Echo {
	e := libHttp.NewEcho()

	handler := Handler{
		spreadsheetsAPI: spreadsheetsAPI,
		receiptsService: receiptsService,
		w: w,
	}

	e.POST("/tickets-confirmation", handler.PostTicketsConfirmation)

	return e
}
