package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"tickets/entities"
	"time"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type OpsBookingReadModel struct {
	db *sqlx.DB
	eb *cqrs.EventBus
}

func NewOpsBookingReadModel(db *sqlx.DB, eb *cqrs.EventBus) *OpsBookingReadModel {
	return &OpsBookingReadModel{
		db: db,
		eb: eb,
	}
}

func (r OpsBookingReadModel) AllBookings(issueDate *time.Time) ([]entities.OpsBooking, error) {
	query := `SELECT payload FROM read_model_ops_bookings`
	var queryArgs []any
	
	if issueDate != nil {
		query += `
			WHERE booking_id IN (
    SELECT booking_id FROM (
        SELECT booking_id,
            DATE(jsonb_path_query(payload, '$.tickets.*.receipt_issued_at')::text) as receipt_issued_at
        FROM
            read_model_ops_bookings
    ) bookings_within_date
    WHERE receipt_issued_at = $1
			)
			`
		queryArgs = append(queryArgs, issueDate)
	}

	rows, err := r.db.Query(query, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []entities.OpsBooking
	for rows.Next() {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}

		var reservation entities.OpsBooking
		if err := json.Unmarshal(payload, &reservation); err != nil {
			return nil, err
		}

		result = append(result, reservation)
	}

	return result, nil
}

func (r OpsBookingReadModel) ReservationReadModel(ctx context.Context, bookingID string) (entities.OpsBooking, error) {
	return r.findReadModelByBookingID(ctx, bookingID, r.db)
}

func (r OpsBookingReadModel) OnBookingMade(ctx context.Context, payload *entities.BookingMade_v1) error {
	readModel := entities.OpsBooking{
		BookingID:  uuid.MustParse(payload.BookingID),
		BookedAt:   payload.Header.PublishedAt,
		LastUpdate: time.Now(),
	}
	err := r.createReadModel(ctx, readModel)
	if err != nil {
		return fmt.Errorf("could not create read model: %w", err)
	}
	return r.eb.Publish(ctx, &entities.InternalOpsReadModelUpdated{
		Header: entities.NewMessageHeader(),
		BookingID: uuid.MustParse(payload.BookingID),
	})
}

func (r OpsBookingReadModel) OnTicketReceiptIssued(ctx context.Context, payload *entities.TicketReceiptIssued_v1) error {
	return r.updateReadModelByTicketID(ctx, payload.TicketID, func(ticket entities.OpsTicket) (entities.OpsTicket, error) {
		ticket.ReceiptIssuedAt = payload.IssuedAt
		ticket.ReceiptNumber = payload.ReceiptNumber
		return ticket, nil
	})
}
func (r OpsBookingReadModel) OnTicketBookingCOnfirmed(ctx context.Context, payload *entities.TicketBookingConfirmed_v1) error {
	return r.updateReadModelByBookingID(ctx, payload.BookingID, func(b entities.OpsBooking) (entities.OpsBooking, error) {
		tkt, ok := b.Tickets[payload.TicketID]
		if !ok {
			log.FromContext(ctx).With("ticket_id", payload.TicketID).Debug("creating a new read model for ticket " + payload.TicketID)
			tkt = entities.OpsTicket{}
		}
		tkt.CustomerEmail = payload.CustomerEmail
		tkt.PriceAmount = payload.Price.Amount
		tkt.PriceCurrency = payload.Price.Currency
		tkt.ConfirmedAt = payload.Header.PublishedAt
		b.Tickets[payload.TicketID] = tkt
		return b, nil
	})
}
func (r OpsBookingReadModel) OnTicketPrinted(ctx context.Context, payload *entities.TicketPrinted_v1) error {
	return r.updateReadModelByTicketID(ctx, payload.TicketID, func(ticket entities.OpsTicket) (entities.OpsTicket, error) {
		ticket.PrintedAt = payload.Header.PublishedAt
		ticket.PrintedFileName = payload.FileName
		return ticket, nil
	})
}
func (r OpsBookingReadModel) OnTicketRefunded(ctx context.Context, payload *entities.TicketRefunded_v1) error {
	return r.updateReadModelByTicketID(ctx, payload.TicketID, func(ticket entities.OpsTicket) (entities.OpsTicket, error) {
		ticket.RefundedAt = payload.Header.PublishedAt
		return ticket, nil
	})
}

func (r OpsBookingReadModel) createReadModel(
	ctx context.Context,
	booking entities.OpsBooking,
) (err error) {
	payload, err := json.Marshal(booking)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO
		    read_model_ops_bookings (payload, booking_id)
		VALUES
			($1, $2)
		ON CONFLICT (booking_id) DO NOTHING; -- read model may be already updated by another event - we don't want to override
`, payload, booking.BookingID)

	if err != nil {
		return fmt.Errorf("could not create read model: %w", err)
	}

	return nil
}

func (r OpsBookingReadModel) updateReadModelByBookingID(
	ctx context.Context,
	bookingID string,
	updateFunc func(ticket entities.OpsBooking) (entities.OpsBooking, error),
) (err error) {
	err = updateInTx(
		ctx,
		r.db,
		sql.LevelRepeatableRead,
		func(ctx context.Context, tx *sqlx.Tx) error {
			rm, err := r.findReadModelByBookingID(ctx, bookingID, tx)
			if errors.Is(err, sql.ErrNoRows) {
				// events arrived out of order - it should spin until the read model is created
				return fmt.Errorf("read model for booking %s not exist yet", bookingID)
			} else if err != nil {
				return fmt.Errorf("could not find read model: %w", err)
			}

			updatedRm, err := updateFunc(rm)
			if err != nil {
				return err
			}

			return r.updateReadModel(ctx, tx, updatedRm)
		},
	)
	if err != nil {
		return err
	}
	return r.eb.Publish(ctx, &entities.InternalOpsReadModelUpdated{
		Header: entities.NewMessageHeader(),
		BookingID: uuid.MustParse(bookingID),
	})
}

func (r OpsBookingReadModel) updateReadModelByTicketID(
	ctx context.Context,
	ticketID string,
	updateFunc func(ticket entities.OpsTicket) (entities.OpsTicket, error),
) (err error) {
	return updateInTx(
		ctx,
		r.db,
		sql.LevelRepeatableRead,
		func(ctx context.Context, tx *sqlx.Tx) error {
			rm, err := r.findReadModelByTicketID(ctx, ticketID, tx)
			if errors.Is(err, sql.ErrNoRows) {
				// events arrived out of order - it should spin until the read model is created
				return fmt.Errorf("read model for ticket %s not exist yet", ticketID)
			} else if err != nil {
				return fmt.Errorf("could not find read model: %w", err)
			}

			ticket, _ := rm.Tickets[ticketID]

			updatedRm, err := updateFunc(ticket)
			if err != nil {
				return err
			}

			rm.Tickets[ticketID] = updatedRm

			err = r.updateReadModel(ctx, tx, rm)
			if err != nil {
				return err
			}
			return r.eb.Publish(ctx, &entities.InternalOpsReadModelUpdated{
				Header: entities.NewMessageHeader(),
				BookingID: rm.BookingID,
			})
			
		},
	)
}

func (r OpsBookingReadModel) updateReadModel(
	ctx context.Context,
	tx *sqlx.Tx,
	rm entities.OpsBooking,
) error {
	rm.LastUpdate = time.Now()

	payload, err := json.Marshal(rm)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO
			read_model_ops_bookings (payload, booking_id)
		VALUES
			($1, $2)
		ON CONFLICT (booking_id) DO UPDATE SET payload = excluded.payload;
		`, payload, rm.BookingID)
	if err != nil {
		return fmt.Errorf("could not update read model: %w", err)
	}

	return nil
}

func (r OpsBookingReadModel) findReadModelByTicketID(
	ctx context.Context,
	ticketID string,
	db dbExecutor,
) (entities.OpsBooking, error) {
	var payload []byte

	err := db.QueryRowContext(
		ctx,
		"SELECT payload FROM read_model_ops_bookings WHERE payload::jsonb -> 'tickets' ? $1",
		ticketID,
	).Scan(&payload)
	if err != nil {
		return entities.OpsBooking{}, err
	}

	return r.unmarshalReadModelFromDB(payload)
}

func (r OpsBookingReadModel) findReadModelByBookingID(
	ctx context.Context,
	bookingID string,
	db dbExecutor,
) (entities.OpsBooking, error) {
	var payload []byte

	err := db.QueryRowContext(
		ctx,
		"SELECT payload FROM read_model_ops_bookings WHERE booking_id = $1",
		bookingID,
	).Scan(&payload)
	if err != nil {
		return entities.OpsBooking{}, err
	}

	return r.unmarshalReadModelFromDB(payload)
}

func (r OpsBookingReadModel) unmarshalReadModelFromDB(payload []byte) (entities.OpsBooking, error) {
	var dbReadModel entities.OpsBooking
	if err := json.Unmarshal(payload, &dbReadModel); err != nil {
		return entities.OpsBooking{}, err
	}

	if dbReadModel.Tickets == nil {
		dbReadModel.Tickets = map[string]entities.OpsTicket{}
	}

	return dbReadModel, nil
}

type dbExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}
