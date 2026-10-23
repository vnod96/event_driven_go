package db

import (
	"context"
	"fmt"
	"tickets/entities"
	"time"

	"github.com/jmoiron/sqlx"
)

type EventRepository interface {
	Save(ctx context.Context, event *entities.Event, eventName string, payload []byte) error
	GetEvents(ctx context.Context) ([]DataLakeEvent, error)
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

type DataLakeEvent struct {
	EventID      string    `db:"event_id"`
	PublishedAt  time.Time `db:"published_at"`
	EventName    string    `db:"event_name"`
	EventPayload []byte    `db:"event_payload"`
}

// ...

func (d *PostgresEventRepository) GetEvents(ctx context.Context) ([]DataLakeEvent, error) {
	var events []DataLakeEvent
	err := d.db.SelectContext(ctx, &events, "SELECT * FROM events ORDER BY published_at ASC")
	if err != nil {
		return nil, fmt.Errorf("could not get events from data lake: %w", err)
	}

	return events, nil
}
