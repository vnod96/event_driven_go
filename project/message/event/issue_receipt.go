package event

import (
	"context"
	"log/slog"
	"tickets/entities"
)

func (h Handler) IssueReceipt(ctx context.Context, event entities.TicketBookingConfirmed) error {
	slog.Info("issueing receipt for ticket")
	return h.receiptsService.IssueReceipt(ctx, event)
}
