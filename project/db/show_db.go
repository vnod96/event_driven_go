package db

import (
	"context"
	"fmt"
	"tickets/entities"

	"github.com/jmoiron/sqlx"
)

type PostgersShowRepository struct {
	db *sqlx.DB
}

func NewShowRepository(dbConn *sqlx.DB) *PostgersShowRepository {
	return &PostgersShowRepository{db: dbConn}
}

func (r *PostgersShowRepository) Save(context context.Context, show entities.Show) error {
	fmt.Println("saving show")
	_, err := r.db.NamedExecContext(context, `INSERT INTO shows(show_id, dead_nation_id, number_of_tickets, start_time, title, venue)
	 VALUES(:show_id, :dead_nation_id, :number_of_tickets, :start_time, :title, :venue) ON CONFLICT DO NOTHING;`, show)
	return err
}

func (r *PostgersShowRepository) ShowByID(ctx context.Context, showID string) (entities.Show, error) {
	var show entities.Show
	err := r.db.GetContext(ctx, &show, `SELECT * FROM shows WHERE show_id = $1`, showID)
	if(err != nil) {
		return entities.Show{}, err
	}
	return show, nil
}