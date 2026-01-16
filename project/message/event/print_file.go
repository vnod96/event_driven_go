package event

import (
	"context"
	"fmt"
	"tickets/entities"
)

func (h Handler) PrintFile(ctx context.Context, tkt *entities.TicketBookingConfirmed) error {
	fileId := tkt.TicketID+"-ticket.html"
	fileContent := fmt.Sprintf(`
		TicketID : %s
		Price    : %s %s
	`, tkt.TicketID, tkt.Price.Currency, tkt.Price.Amount)

	return h.fileService.PutFile(ctx, fileId, fileContent)
}