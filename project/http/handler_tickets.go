package http

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"

	commonLog "github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
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
	ctx := context.Background()
	ctx = commonLog.ContextWithCorrelationID(ctx, correlationID)

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
			err = h.eventBus.Publish(ctx, ticketEvent)
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
			err = h.eventBus.Publish(ctx, ticketEvent)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("unknown ticket status: %s", ticket.Status)
		}
	}

	return c.NoContent(http.StatusOK)
}
