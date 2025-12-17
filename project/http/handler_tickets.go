package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/labstack/echo/v4"

	"tickets/entities"
)

type ticketsConfirmationRequest struct {
	Tickets []string `json:"tickets"`
}

type TicketStatusRequest struct {
	TicketID      string         `json:"ticket_id"`
	Status        string         `json:"status"`
	Price         entities.Money `json:"price"`
	CustomerEmail string         `json:"customer_email"`
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

	correlationID := c.Request().Header.Get("Correlation-ID")

	log.Println("Recieved Req======================", len(request.Tickets))
	for _, ticket := range request.Tickets {
		switch ticket.Status {
		case "confirmed":
			ticketEvent := entities.TicketBookingConfirmed{
				Header:        entities.NewMessageHeader(),
				TicketID:      ticket.TicketID,
				CustomerEmail: ticket.CustomerEmail,
				Price:         ticket.Price,
			}
			ticketEventJson, err := json.Marshal(ticketEvent)
			if err != nil {
				return err
			}
			msg := message.NewMessage(watermill.NewUUID(), ticketEventJson)
			msg.Metadata.Set("type", "TicketBookingConfirmed")
			middleware.SetCorrelationID(correlationID, msg)
			err = h.pub.Publish("TicketBookingConfirmed", msg)
			if err != nil {
				return err
			}

		case "canceled":
			ticketEvent := entities.TicketBookingCanceled{
				Header:        entities.NewMessageHeader(),
				TicketID:      ticket.TicketID,
				CustomerEmail: ticket.CustomerEmail,
				Price:         ticket.Price,
			}
			ticketEventJson, err := json.Marshal(ticketEvent)
			if err != nil {
				return err
			}
			msg := message.NewMessage(watermill.NewUUID(), ticketEventJson)
			msg.Metadata.Set("type", "TicketBookingCanceled")
			middleware.SetCorrelationID(correlationID, msg)
			err = h.pub.Publish("TicketBookingCanceled", msg)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("unknown ticket status: %s", ticket.Status)
		}
	}

	return c.NoContent(http.StatusOK)
}
