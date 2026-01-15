package db

import (
	"context"
	"fmt"
	"tickets/entities"

	"github.com/jmoiron/sqlx"
)

type TicketRepository interface {
	SaveTicket(context context.Context, ticket *entities.TicketBookingConfirmed) error
	RemoveTicket(context context.Context, ticket *entities.TicketBookingCanceled) error
}

type PostgresTicketRepostory struct {
	db *sqlx.DB
}

func NewTicketReposity(dbConn *sqlx.DB) *PostgresTicketRepostory {
	return &PostgresTicketRepostory{db: dbConn}
}

func (t *PostgresTicketRepostory) SaveTicket(context context.Context, ticket *entities.TicketBookingConfirmed) error {
	fmt.Println("saving ticket in db.")
	_, err := t.db.NamedExecContext(context, `INSERT INTO tickets(ticket_id, price_amount, price_currency, customer_email)
	VALUES (:ticket_id, :price.amount, :price.currency, :customer_email);`, ticket)
	return err;
}

func (t *PostgresTicketRepostory) RemoveTicket(context context.Context, ticket *entities.TicketBookingCanceled) error {
	_, err := t.db.ExecContext(context, "DELETE FROM tickets WHERE ticket_id=$1", ticket.TicketID)
	return err;
}

