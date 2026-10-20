package adapters

import (
	"context"
	"fmt"
	"net/http"
	"tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients/receipts"
)

type ReceiptsServiceClient struct {
	// we are not mocking this client: it's pointless to use interface here
	clients *clients.Clients
}

func NewReceiptsServiceClient(clients *clients.Clients) *ReceiptsServiceClient {
	if clients == nil {
		panic("NewReceiptsServiceClient: clients is nil")
	}

	return &ReceiptsServiceClient{clients: clients}
}

func (c ReceiptsServiceClient) IssueReceipt(ctx context.Context, payload *entities.TicketBookingConfirmed) (entities.IssueReceiptResponse, error) {
	key := payload.Header.IdempotencyKey + payload.TicketID
	resp, err := c.clients.Receipts.PutReceiptsWithResponse(ctx, receipts.CreateReceipt{
		IdempotencyKey: &key,
		TicketId: payload.TicketID,
		Price: receipts.Money{
			MoneyAmount: payload.Price.Amount,
			MoneyCurrency: payload.Price.Currency,
		},
	})
	if err != nil {
		return entities.IssueReceiptResponse{}, fmt.Errorf("failed to post receipt: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		// receipt already exists
		return entities.IssueReceiptResponse{
			IssuedAt: resp.JSON200.IssuedAt,
			ReceiptNumber: resp.JSON200.Number,
		},nil
	case http.StatusCreated:
		// receipt was created
		return entities.IssueReceiptResponse{
			IssuedAt: resp.JSON201.IssuedAt,
			ReceiptNumber: resp.JSON201.Number,
		},nil
	default:
		return entities.IssueReceiptResponse{}, fmt.Errorf("unexpected status code for POST receipts-api/receipts: %d", resp.StatusCode())
	}
}

func (c ReceiptsServiceClient) VoidReceipt(ctx context.Context, payload *entities.RefundTicket) error {
	resp, err := c.clients.Receipts.PutVoidReceiptWithResponse(ctx, receipts.VoidReceiptRequest{
		IdempotentId: &payload.Header.IdempotencyKey,
		Reason: "customer requested refund",
		TicketId: payload.TicketID,
	})
	if err != nil {
		return fmt.Errorf("failed to void receipt: %w", err)
	}

	if resp.StatusCode() == http.StatusOK {
		return nil
	}
	
	return fmt.Errorf("unexpected status code for PUT void receipts: %d", resp.StatusCode())
}