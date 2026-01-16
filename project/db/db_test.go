package db

import (
	"context"
	"os"
	"sync"
	"testing"
	"tickets/entities"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

var dbConn *sqlx.DB
var getDbOnce sync.Once

func GetDb() *sqlx.DB {
	getDbOnce.Do(func() {
		var err error
		dbConn, err = sqlx.Open("postgres", os.Getenv("POSTGRES_URL"))
		if err != nil {
			panic(err)
		}
	})
	return dbConn
}

func TestRepo(t *testing.T) {
	 ctx := context.Background()
	db := GetDb()
	err := InitializeSchema(db)
	require.NoError(t, err)
	repo := NewTicketReposity(db)
	tkt := entities.Ticket {
		TicketID: uuid.NewString(),
		CustomerEmail: "v.x@mail.com",
		Price: entities.Money{
			Amount: "50",
			Currency: "USD",
		},
	}

	for i := 0; i < 2; i++ {
		err = repo.Save(ctx, &tkt)
		require.NoError(t, err)

		tkts, err := repo.FindAll(ctx)
		require.NoError(t, err)

		var len int

		for _, _t := range tkts {
			if _t.TicketID == tkt.TicketID {
				len++
			}
		}

		require.Equal(t, len, 1)
	}

}