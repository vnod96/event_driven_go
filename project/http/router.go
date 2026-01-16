package http

import (
	"net/http"

	libHttp "github.com/ThreeDotsLabs/go-event-driven/v2/common/http"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/labstack/echo/v4"
	"tickets/db"
)

func NewHttpRouter(
	eventBus *cqrs.EventBus,
	repo db.TicketRepository,
) *echo.Echo {
	e := libHttp.NewEcho()

	handler := Handler{
		eventBus: eventBus,
		repo : repo,
	}

	e.POST("/tickets-confirmation", handler.PostTicketsConfirmation)
	e.POST("/tickets-status", handler.PostTicketsConfirmation)

	e.GET("tickets", handler.GetTickets)

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	return e
}
