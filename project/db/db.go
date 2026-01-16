package db

import (
	"context"
	"fmt"
	"tickets/entities"

	"github.com/jmoiron/sqlx"
)

type TicketRepository interface {
	Save(context context.Context, ticket *entities.Ticket) error
	Remove(context context.Context, ticketId string) error
	FindAll(context context.Context) ([]entities.Ticket, error)
}

type PostgresTicketRepostory struct {
	db *sqlx.DB
}

func NewTicketReposity(dbConn *sqlx.DB) *PostgresTicketRepostory {
	return &PostgresTicketRepostory{db: dbConn}
}

func (t *PostgresTicketRepostory) Save(context context.Context, ticket *entities.Ticket) error {
	fmt.Println("saving ticket in db.")
	_, err := t.db.NamedExecContext(context, `INSERT INTO tickets(ticket_id, price_amount, price_currency, customer_email)
	VALUES (:ticket_id, :price.amount, :price.currency, :customer_email) ON CONFLICT DO NOTHING;`, ticket)
	return err;
}

func (t *PostgresTicketRepostory) Remove(context context.Context, ticketId string) error {
	_, err := t.db.ExecContext(context, "DELETE FROM tickets WHERE ticket_id=$1", ticketId)
	return err;
}

func (t *PostgresTicketRepostory) FindAll(ctx context.Context) ([]entities.Ticket, error) {
	var tickets []entities.Ticket
	err := t.db.SelectContext(
		ctx,
		&tickets,
		`SELECT
			ticket_id,
			price_amount as "price.amount",
			price_currency as "price.currency",
			customer_email
		FROM
			tickets;
		`,
	)

	if err != nil {
		return nil, err
	}
	return tickets, nil
}

