package event

import (
	"context"
	"tickets/entities"
)

func (h Handler) AppendRow(ctx context.Context, event entities.TicketBookingConfirmed) error {
	return h.spreadsheetsAPI.AppendRow(ctx, "tickets-to-print", []string{event.TicketID, event.CustomerEmail, event.Price.Amount, event.Price.Currency})
}
