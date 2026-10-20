package adapters

import (
	"context"
	"tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients/dead_nation"
	"github.com/google/uuid"
)

type DeadNationClient struct {
	clients *clients.Clients
}

func NewDeadNationClient(clients *clients.Clients) *DeadNationClient {
	return &DeadNationClient{
		clients: clients,
	}
}

func (d *DeadNationClient) PostTicket(ctx context.Context, e *entities.BookingMade_v1, deadNationId uuid.UUID) error {
	_, err := d.clients.DeadNation.PostTicketBookingWithResponse(ctx, dead_nation.PostTicketBookingRequest{
		BookingId:       uuid.MustParse(e.BookingID),
		CustomerAddress: e.CustomerEmail,
		EventId:         deadNationId,
		NumberOfTickets: e.NumberOfTickets,
	})

	if err != nil {
		return err
	}

	return nil
}
