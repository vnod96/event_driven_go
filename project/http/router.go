package http

import (
	"net/http"

	libHttp "github.com/ThreeDotsLabs/go-event-driven/v2/common/http"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/labstack/echo/v4"
	"tickets/db"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewHttpRouter(
	eventBus *cqrs.EventBus,
	cmdBus *cqrs.CommandBus,
	repo db.TicketRepository,
	showRepo db.ShowRepository,
	bookingRepo db.BookingRepository,
	opsReadModel db.OpsBookingReadModel,
) *echo.Echo {
	e := libHttp.NewEcho()

	handler := Handler{
		eventBus: eventBus,
		cmdBus: cmdBus,
		repo : repo,
		showRepo: showRepo,
		bookingRepo: bookingRepo,
		opsReadModel: opsReadModel,
	}

	e.POST("/tickets-confirmation", handler.PostTicketsConfirmation)
	e.POST("/tickets-status", handler.PostTicketsConfirmation)

	e.POST("/shows", handler.CreateShow)
	e.POST("/book-tickets", handler.BookTickets)

	e.PUT("/ticket-refund/:ticket_id", handler.RefundTicket)

	e.GET("tickets", handler.GetTickets)

	e.GET("/ops/bookings", handler.GetOpsBookings)
	e.GET("/ops/bookings/:id", handler.GetOpsBookingByID)
	

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	return e
}
