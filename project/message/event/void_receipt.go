package event

import (
	"context"
	"fmt"
	"log/slog"
	"tickets/entities"
)

func (h Handler) VoidReceipt(ctx context.Context, event *entities.RefundTicket) error {
	slog.Info("voiding receipt for ticket")
	err := h.receiptsService.VoidReceipt(ctx, event)
	if err != nil {
		return fmt.Errorf("failed to void receipt: %w", err)
	}

	return h.paymentService.RefundPayment(ctx, entities.PaymentRefund{
		TicketID: event.TicketID,
		RefundReason: "voided",
		IdempotencyKey: event.Header.IdempotencyKey,
	})
}
