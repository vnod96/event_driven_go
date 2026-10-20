package event

import (
	"context"
	"fmt"
	"tickets/entities"
)

func (h Handler) PrintFile(ctx context.Context, tkt *entities.TicketBookingConfirmed_v1) error {
	fileId := tkt.TicketID + "-ticket.html"
	fileContent := fmt.Sprintf(`
		TicketID : %s
		Price    : %s %s
	`, tkt.TicketID, tkt.Price.Currency, tkt.Price.Amount)

	err := h.fileService.PutFile(ctx, fileId, fileContent)
	if err != nil {
		return err
	}

	return h.eventBus.Publish(ctx, entities.TicketPrinted_v1{
		Header:   entities.NewMessageHeader(),
		TicketID: tkt.TicketID,
		FileName: fileId,
	})

}
