package http

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"tickets/db"
)

type Handler struct {
	eventBus *cqrs.EventBus
	repo db.TicketRepository
}
