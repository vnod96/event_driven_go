package service

import (
	"context"
	"errors"
	stdHTTP "net/http"

	"github.com/ThreeDotsLabs/watermill"
	watermillMsg "github.com/ThreeDotsLabs/watermill/message"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"

	"tickets/db"
	ticketsHttp "tickets/http"
	"tickets/message"
	"tickets/message/event"
	"tickets/worker"
)

type Service struct {
	db              *sqlx.DB
	echoRouter      *echo.Echo
	watermillRouter *watermillMsg.Router
}

func New(
	dbConn *sqlx.DB,
	spreadsheetsAPI worker.SpreadsheetsAPI,
	receiptsService worker.ReceiptsService,
	redisClient *redis.Client,
) Service {
	logger := watermill.NewSlogLogger(nil)
	pub := message.NewRedisPublisher(redisClient, logger)
	eb := event.NewEventBus(pub, logger)
	repo := db.NewTicketReposity(dbConn)
	eventsHandler := event.NewHandler(receiptsService, spreadsheetsAPI, repo)
	watermillRouter := message.NewWatermillRouter(event.NewEventProcessorConfig(redisClient, logger), eventsHandler, logger)
	echoRouter := ticketsHttp.NewHttpRouter(eb, repo)

	return Service{
		db: dbConn,
		echoRouter:      echoRouter,
		watermillRouter: watermillRouter,
	}
}

func (s Service) Run(ctx context.Context) error {
	err := db.InitializeSchema(s.db)

	if err != nil {
		return err
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return s.watermillRouter.Run(ctx)
	})

	g.Go(func() error {
		<-s.watermillRouter.Running()

		err := s.echoRouter.Start(":8080")
		if err != nil && !errors.Is(err, stdHTTP.ErrServerClosed) {
			return err
		}

		return nil
	})

	g.Go(func() error {
		// Shut down the HTTP server
		<-ctx.Done()
		return s.echoRouter.Shutdown(context.Background())
	})

	return g.Wait()
}
