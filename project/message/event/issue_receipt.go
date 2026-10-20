package event

import (
	"context"
	"log/slog"
	"tickets/entities"
)

func (h Handler) IssueReceipt(ctx context.Context, event *entities.TicketBookingConfirmed_v1) error {
	slog.Info("issueing receipt for ticket")
	resp, err := h.receiptsService.IssueReceipt(ctx, event)
	if err != nil {
		return err
	}
	return h.eventBus.Publish(ctx, entities.TicketReceiptIssued_v1 {
		Header: event.Header,
		TicketID: event.TicketID,
		ReceiptNumber: resp.ReceiptNumber,
		IssuedAt: resp.IssuedAt,
	})
}
