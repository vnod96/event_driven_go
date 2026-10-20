package db

import (
	"context"
	"tickets/entities"

	"github.com/jmoiron/sqlx"
)

type EventRepository interface {
	Save(ctx context.Context, event *entities.Event, eventName string, payload []byte) error
}

type PostgresEventRepository struct {
	db *sqlx.DB
}

func NewPostgresEventRepository(db *sqlx.DB) *PostgresEventRepository {
	return &PostgresEventRepository{db}
}

func (p *PostgresEventRepository) Save(ctx context.Context, event *entities.Event, eventName string, payload []byte) error {
	_, err := p.db.ExecContext(ctx, `INSERT INTO events (event_id, published_at, event_name, event_payload)
		VALUES ($1, $2, $3, $4);`,event.Header.ID, event.Header.PublishedAt,  eventName, payload)
	if err != nil {
		return err
	}
	return nil
}
