package db

import (
	"context"
	"fmt"
	"tickets/entities"

	"github.com/jmoiron/sqlx"
)

type TicketRepository interface {
	SaveTicket(context context.Context, ticket *entities.TicketBookingConfirmed) error
}

type PostgresTicketRepostory struct {
	db *sqlx.DB
}

func NewTicketReposity(dbConn *sqlx.DB) *PostgresTicketRepostory {
	return &PostgresTicketRepostory{db: dbConn}
}

func (t *PostgresTicketRepostory) SaveTicket(context context.Context, ticket *entities.TicketBookingConfirmed) error {
	fmt.Println("saving ticket in db.")
	_, err := t.db.NamedExec(`INSERT INTO tickets(ticket_id, price_amount, price_currency, customer_email)
	VALUES (:ticket_id, 0, 'USD', :customer_email);`, ticket)
	return err;
}

