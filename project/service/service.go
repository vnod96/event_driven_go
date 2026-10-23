package service

import (
	"context"
	"errors"
	"fmt"
	stdHTTP "net/http"
	"time"

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
	"tickets/migrate"
	"tickets/outbox"
	"tickets/worker"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Service struct {
	db              *sqlx.DB
	echoRouter      *echo.Echo
	watermillRouter *watermillMsg.Router
	migrator        *migrate.Migrator
}

func New(
	dbConn *sqlx.DB,
	spreadsheetsAPI worker.SpreadsheetsAPI,
	receiptsService worker.ReceiptsService,
	fileService worker.FileService,
	deadNation worker.DeadNationService,
	redisClient *redis.Client,
	paymentService worker.PaymentService,
) Service {
	logger := watermill.NewSlogLogger(nil)
	pub := message.NewRedisPublisher(redisClient, logger)
	sub := outbox.NewPostgresSubscriber(dbConn.DB, logger)
	redisSub := message.NewRedisSubscriber(redisClient, "svc-tickets.events_splitter", logger)
	redisDatalakeSub := message.NewRedisSubscriber(redisClient, "svc-tickets.store_to_data_lake", logger)
	eb := event.NewEventBus(pub, logger)
	cb := event.NewCommandBus(pub, logger)
	repo := db.NewTicketReposity(dbConn)
	showRepo := db.NewShowRepository(dbConn)
	opsReadModel := db.NewOpsBookingReadModel(dbConn, eb)
	bookingRepo := db.NewBookingRepository(dbConn, func(pub watermillMsg.Publisher) *cqrs.EventBus { return event.NewEventBus(pub, logger) })
	eventRepo := db.NewPostgresEventRepository(dbConn)
	eventsHandler := event.NewHandler(receiptsService, spreadsheetsAPI, fileService, repo, showRepo, eb, deadNation, paymentService, opsReadModel, eventRepo)
	watermillRouter := message.NewWatermillRouter(sub, redisSub, redisDatalakeSub, pub, event.NewEventProcessorConfig(redisClient, logger), event.NewCommandProcessorConfig(redisClient, logger), eventsHandler, logger)
	echoRouter := ticketsHttp.NewHttpRouter(eb, cb, repo, showRepo, bookingRepo, *opsReadModel)

	mig := migrate.NewMigrator(eventRepo, *opsReadModel)

	return Service{
		db:              dbConn,
		echoRouter:      echoRouter,
		watermillRouter: watermillRouter,
		migrator:        mig,
	}
}

func (s Service) Run(ctx context.Context) error {
	err := db.InitializeSchema(s.db)

	var (
		veryImportantCounter = promauto.NewCounter(prometheus.CounterOpts{
			// metric will be named tickets_very_important_counter_total
			Namespace: "tickets",
			Name:      "very_important_counter_total",
			Help:      "Total number of very important things processed",
		})
	)
	go func() {
		for {
			veryImportantCounter.Inc()
			time.Sleep(time.Millisecond * 100)
		}
	}()

	if err != nil {
		return err
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return s.watermillRouter.Run(ctx)
	})

	g.Go(func() error {
		<-s.watermillRouter.Running()
		if err = s.migrator.Run(ctx); err != nil {
			fmt.Printf("failed to migrate %w \n", err)
		}

		return nil
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
