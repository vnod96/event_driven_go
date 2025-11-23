package service

import (
	"context"
	"errors"
	"log/slog"
	stdHTTP "net/http"

	"github.com/ThreeDotsLabs/watermill"
	watermillMsg "github.com/ThreeDotsLabs/watermill/message"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"

	ticketsHttp "tickets/http"
	"tickets/message"
	"tickets/worker"
)

type Service struct {
	echoRouter *echo.Echo
	watermillRouter *watermillMsg.Router
}

func New(
	spreadsheetsAPI worker.SpreadsheetsAPI,
	receiptsService worker.ReceiptsService,
	redisClient *redis.Client,
) Service {
	logger := watermill.NewSlogLogger(nil)
	pub := message.NewRedisPublisher(redisClient, logger)
	watermillRouter := message.NewWatermillRouter(spreadsheetsAPI, receiptsService, redisClient, logger)
	echoRouter := ticketsHttp.NewHttpRouter(pub)

	return Service{
		echoRouter: echoRouter,
		watermillRouter: watermillRouter,
	}
}

func (s Service) Run(ctx context.Context) error {
	go func() {
		err:=s.watermillRouter.Run(ctx)
		if err != nil {
			slog.With("error", err).Error("failed to start router")
		}
	}()
	err := s.echoRouter.Start(":8080")
	if err != nil && !errors.Is(err, stdHTTP.ErrServerClosed) {
		return err
	}

	return nil
}
