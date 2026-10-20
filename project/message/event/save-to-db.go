package event

import (
	"context"
	"tickets/entities"
)

func (h Handler) SaveTicket(ctx context.Context, tkt *entities.TicketBookingConfirmed_v1) error {
	return h.ticketRepository.Save(ctx, entities.ToTicket(tkt))
}
