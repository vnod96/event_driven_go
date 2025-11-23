package http

import (
	"tickets/worker"

	libHttp "github.com/ThreeDotsLabs/go-event-driven/v2/common/http"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/labstack/echo/v4"
)

func NewHttpRouter(
	spreadsheetsAPI worker.SpreadsheetsAPI,
	receiptsService worker.ReceiptsService,
	pub *redisstream.Publisher,
) *echo.Echo {
	e := libHttp.NewEcho()

	handler := Handler{
		spreadsheetsAPI: spreadsheetsAPI,
		receiptsService: receiptsService,
		pub: pub,
	}

	e.POST("/tickets-confirmation", handler.PostTicketsConfirmation)

	return e
}
