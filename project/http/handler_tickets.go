package http

import (
	"net/http"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/labstack/echo/v4"
)

type ticketsConfirmationRequest struct {
	Tickets []string `json:"tickets"`
}

func (h Handler) PostTicketsConfirmation(c echo.Context) error {
	var request ticketsConfirmationRequest
	err := c.Bind(&request)
	if err != nil {
		return err
	}

	for _, ticket := range request.Tickets {
		err := h.pub.Publish("issue-receipt", message.NewMessage(watermill.NewUUID(), []byte(ticket)))

		if err != nil {
			return err
		}
		err = h.pub.Publish("append-to-tracker", message.NewMessage(watermill.NewUUID(), []byte(ticket)))
		if err != nil {
			return err
		}

		// h.w.Send(worker.Message{
		// 	Task:     worker.TaskIssueReceipt,
		// 	TicketID: ticket,
		// },
		// 	worker.Message{
		// 		Task:     worker.TaskAppendToTracker,
		// 		TicketID: ticket,
		// 	})
		// err = h.receiptsService.IssueReceipt(context.Background(), ticket)
		// if err != nil {
		// 	return err
		// }

		// err = h.spreadsheetsAPI.AppendRow(context.Background(), "tickets-to-print", []string{ticket})
		// if err != nil {
		// 	return err
		// }
	}

	return c.NoContent(http.StatusOK)
}
