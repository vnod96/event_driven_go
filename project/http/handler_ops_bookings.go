package http

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func (h Handler) GetOpsBookings(c echo.Context) error {
	issueDate := c.QueryParam("receipt_issue_date")
	var parsedTime *time.Time
	if issueDate != "" {
		date, err := time.Parse(time.DateOnly, issueDate)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "invalid receipt_issue_date format. Expected YYYY-MM-DD",
			})
		}
		parsedTime = &date
	}
	
	bookings, err := h.opsReadModel.AllBookings(parsedTime)
	if err != nil {
		return err
	}
	c.JSON(http.StatusOK, bookings)
	return nil
}

func (h Handler) GetOpsBookingByID(c echo.Context) error {
	booking, err := h.opsReadModel.ReservationReadModel(c.Request().Context(), c.Param("id"))
	if err != nil {
		return err
	}
	c.JSON(http.StatusOK, booking)
	return nil
}
