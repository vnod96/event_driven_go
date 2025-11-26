package event

import (
	"context"
	"log/slog"
	"tickets/entities"
)

func (h Handler) RefundTicket(ctx context.Context, event entities.TicketBookingCanceled) error {
	slog.Info("refunding ticket")
	return h.spreadsheetsAPI.AppendRow(ctx, "tickets-to-refund", []string{event.TicketID, event.CustomerEmail, event.Price.Amount, event.Price.Currency})
}