package service

import (
	"context"
	"errors"
	stdHTTP "net/http"

	"github.com/ThreeDotsLabs/watermill"
	watermillMsg "github.com/ThreeDotsLabs/watermill/message"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"

	ticketsHttp "tickets/http"
	"tickets/message/event"
	"tickets/message"
	"tickets/worker"
)

type Service struct {
	echoRouter      *echo.Echo
	watermillRouter *watermillMsg.Router
}

func New(
	spreadsheetsAPI worker.SpreadsheetsAPI,
	receiptsService worker.ReceiptsService,
	redisClient *redis.Client,
) Service {
	logger := watermill.NewSlogLogger(nil)
	pub := message.NewRedisPublisher(redisClient, logger)
	eb := event.NewEventBus(pub, logger)
	watermillRouter := message.NewWatermillRouter(spreadsheetsAPI, receiptsService, redisClient, logger)
	echoRouter := ticketsHttp.NewHttpRouter(eb)

	return Service{
		echoRouter:      echoRouter,
		watermillRouter: watermillRouter,
	}
}

func (s Service) Run(ctx context.Context) error {

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return s.watermillRouter.Run(ctx)
	})

	g.Go(func() error {
		<- s.watermillRouter.Running()

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
