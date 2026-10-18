package event

import (
	"context"
	"tickets/entities"

	"github.com/google/uuid"
)


func (h *Handler) MakeBooking(ctx context.Context, e *entities.BookingMade) error {
	show, err := h.showRepository.ShowByID(ctx, e.ShowID)
	if err != nil {
		return err
	}
	showId := uuid.MustParse(show.DeadNationId)

	return h.deadNationClient.PostTicket(ctx, e, showId)
}
