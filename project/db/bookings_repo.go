package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"tickets/entities"
	"tickets/outbox"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

type BookingRepository interface {
	Save(context context.Context, booking *entities.Booking) error
}

type PostgresBookingRepository struct {
	db          *sqlx.DB
	newEventBus func(pub message.Publisher) *cqrs.EventBus
}

func NewBookingRepository(dbConn *sqlx.DB, newEventBus func(pub message.Publisher) *cqrs.EventBus) *PostgresBookingRepository {
	return &PostgresBookingRepository{db: dbConn, newEventBus: newEventBus}
}

func (r *PostgresBookingRepository) Save(ctx context.Context, booking *entities.Booking) error {

	return updateInTx(ctx, r.db,
		sql.LevelSerializable,
		func(ctx context.Context, tx *sqlx.Tx) error {

			availableSeats := 0
			bookedSeats := 0

			err := tx.GetContext(ctx, &availableSeats, `
							SELECT
								number_of_tickets AS available_seats
							FROM
								shows
							WHERE
								show_id = $1
						`, booking.ShowID)

			if err != nil {
				return fmt.Errorf("could not get available seats : %w", err)
			}

			err = tx.GetContext(ctx, &bookedSeats, `SELECT COALESCE(SUM(number_of_tickets), 0) AS already_booked_seats FROM bookings WHERE show_id = $1`, booking.ShowID)

			if err != nil {
				return fmt.Errorf("could not get booked seats : %w", err)
			}

			if availableSeats - bookedSeats < booking.NumberOfTickets {
				return echo.NewHTTPError(http.StatusBadRequest, "not enough seats available")
			}

			_, err = tx.NamedExecContext(ctx, `INSERT INTO bookings (booking_id, show_id, number_of_tickets, customer_email)
			VALUES (:booking_id, :show_id, :number_of_tickets, :customer_email) ON CONFLICT DO NOTHING;`, booking)
			pub, err := outbox.NewPublisherForDb(tx)
			if err != nil {
				return fmt.Errorf("could not create event bus: %w", err)
			}
			bus := r.newEventBus(pub)
			bus.Publish(ctx, entities.BookingMade{
				Header: entities.NewMessageHeader(),
				BookingID:       booking.BookingID,
				CustomerEmail:   booking.CustomerEmail,
				ShowID:          booking.ShowID,
				NumberOfTickets: booking.NumberOfTickets,
			})
			return err
		})
}

func updateInTx(ctx context.Context, db *sqlx.DB, isolation sql.IsolationLevel, fn func(ctx context.Context, tx *sqlx.Tx) error) (err error) {
	tx, err := db.BeginTxx(ctx, &sql.TxOptions{Isolation: isolation})
	if err != nil {
		return fmt.Errorf("could not begin transaction : %w", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				err = errors.Join(err, rollbackErr)
			}
			return
		}
		err = tx.Commit()
	}()
	return fn(ctx, tx)
}
