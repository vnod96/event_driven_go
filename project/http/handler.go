package http

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"tickets/db"
)

type Handler struct {
	eventBus *cqrs.EventBus
	cmdBus *cqrs.CommandBus
	repo db.TicketRepository
	showRepo db.ShowRepository
	bookingRepo db.BookingRepository
	opsReadModel db.OpsBookingReadModel
}
