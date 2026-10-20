package event

import (
	"context"
	"fmt"
	"log/slog"
	"tickets/entities"
)

func (h Handler) VoidReceipt(ctx context.Context, event *entities.RefundTicket_v1) error {
	slog.Info("voiding receipt for ticket")
	err := h.receiptsService.VoidReceipt(ctx, event)
	if err != nil {
		return fmt.Errorf("failed to void receipt: %w", err)
	}

	err =  h.paymentService.RefundPayment(ctx, entities.PaymentRefund{
		TicketID: event.TicketID,
		RefundReason: "voided",
		IdempotencyKey: event.Header.IdempotencyKey,
	})
	if err != nil {
		return fmt.Errorf("failed to refund payment : %w", err)
	}
	return h.eventBus.Publish(ctx, entities.TicketRefunded_v1{
		Header: event.Header,
		TicketID: event.TicketID,
	})
}
