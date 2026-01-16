package event

import (
	"context"
	"fmt"
	"tickets/entities"
)

func (h Handler) RemoveTicket(ctx context.Context, tkt *entities.TicketBookingCanceled) error {
	fmt.Println("removing ticket.")
	return h.ticketRepository.Remove(ctx, tkt.TicketID)
}