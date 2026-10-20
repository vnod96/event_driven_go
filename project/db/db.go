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

type ShowRepository interface {
	Save(ctx context.Context, show entities.Show) error
	ShowByID(ctx context.Context, showID string) (entities.Show, error)
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
	resp, err := t.db.ExecContext(context, "UPDATE tickets SET deleted_at = CURRENT_TIMESTAMP WHERE ticket_id=$1 AND deleted_at IS NULL;", ticketId)
	if err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}

	rows, err := resp.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("ticket not available")
	}
	
	
	return nil;
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
			tickets
		WHERE
			deleted_at is null;
		`,
	)

	if err != nil {
		return nil, err
	}
	return tickets, nil
}

