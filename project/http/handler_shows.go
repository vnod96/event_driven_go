package http

import (
	"net/http"
	"tickets/entities"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)


func (h *Handler) CreateShow(c echo.Context) error {
	var request entities.Show
	err := c.Bind(&request)
	if err != nil {
		return err
	}
	request.ShowID = uuid.NewString()
	err = h.showRepo.Save(c.Request().Context(), request)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, map[string]string{
		"show_id": request.ShowID,
	})
}

func (h *Handler) BookTickets(c echo.Context) error {
	var request entities.Booking
	err := c.Bind(&request)
	if err != nil {
		return err
	}
	request.BookingID = uuid.NewString()
	err = h.bookingRepo.Save(c.Request().Context(), &request);
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"booking_id": request.BookingID,
	})
	
}