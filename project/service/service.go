package service

import (
	"context"
	"errors"
	stdHTTP "net/http"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	watermillMsg "github.com/ThreeDotsLabs/watermill/message"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"

	"tickets/db"
	ticketsHttp "tickets/http"
	"tickets/message"
	"tickets/message/event"
	"tickets/outbox"
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
	fileService worker.FileService,
	deadNation worker.DeadNationService,
	redisClient *redis.Client,
) Service {
	logger := watermill.NewSlogLogger(nil)
	pub := message.NewRedisPublisher(redisClient, logger)
	sub := outbox.NewPostgresSubscriber(dbConn.DB, logger)
	eb := event.NewEventBus(pub, logger)
	repo := db.NewTicketReposity(dbConn)
	showRepo := db.NewShowRepository(dbConn)
	bookingRepo := db.NewBookingRepository(dbConn, func(pub watermillMsg.Publisher) *cqrs.EventBus { return event.NewEventBus(pub, logger) })
	eventsHandler := event.NewHandler(receiptsService, spreadsheetsAPI, fileService, repo, showRepo, eb, deadNation)
	watermillRouter := message.NewWatermillRouter(sub, pub, event.NewEventProcessorConfig(redisClient, logger), eventsHandler, logger)
	echoRouter := ticketsHttp.NewHttpRouter(eb, repo, showRepo, bookingRepo)

	return Service{
		db:              dbConn,
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
