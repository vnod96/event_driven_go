package event

import (
	"context"
	"tickets/entities"
)

func (h Handler) SaveTicket(ctx context.Context, tkt *entities.TicketBookingConfirmed) error {
	return h.ticketRepository.SaveTicket(ctx, tkt)
}
