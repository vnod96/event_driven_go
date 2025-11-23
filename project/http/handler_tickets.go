package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/labstack/echo/v4"

	"tickets/entities"
)

type ticketsConfirmationRequest struct {
	Tickets []string `json:"tickets"`
}

type TicketStatusRequest struct {
	TicketID string `json:"ticket_id"`
	Status string `json:"status"`
	Price entities.Money  `json:"price"`
	CustomerEmail string `json:"customer_email"`
}

type TicketsStatusRequest struct {
	Tickets []TicketStatusRequest `json:"tickets"`
}

func (h Handler) PostTicketsConfirmation(c echo.Context) error {
	var request TicketsStatusRequest
	err := c.Bind(&request)
	if err != nil {
		return err
	}

	for _, ticket := range request.Tickets {
		if ticket.Status == "confirmed" {
			
		err := h.pub.Publish("issue-receipt", message.NewMessage(watermill.NewUUID(), []byte(ticket.TicketID)))

		if err != nil {
			return err
		}
		event := entities.AppendToTrackerPayload{
			TicketID: ticket.TicketID,
			CustomerEmail: ticket.CustomerEmail,
			Price: ticket.Price,
		}
		eventMsg, err := json.Marshal(event)
		if err != nil {
			return err
		}
		err = h.pub.Publish("append-to-tracker", message.NewMessage(watermill.NewUUID(), eventMsg))
		if err != nil {
			return err
		}
		}else {
			return fmt.Errorf("unknown ticket status: %s", ticket.Status)
		}
	}

	return c.NoContent(http.StatusOK)
}
