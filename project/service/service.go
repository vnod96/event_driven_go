package service

import (
	"context"
	"errors"
	stdHTTP "net/http"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"

	ticketsHttp "tickets/http"
	"tickets/message"
	"tickets/worker"
)

type Service struct {
	echoRouter *echo.Echo
}

func New(
	spreadsheetsAPI worker.SpreadsheetsAPI,
	receiptsService worker.ReceiptsService,
	redisClient *redis.Client,
) Service {
	w := message.NewPubSubWorker(spreadsheetsAPI, receiptsService)
	echoRouter := ticketsHttp.NewHttpRouter(spreadsheetsAPI, receiptsService, w.Pub)

	go w.Run()

	return Service{
		echoRouter: echoRouter,
	}
}

func (s Service) Run(ctx context.Context) error {
	err := s.echoRouter.Start(":8080")
	if err != nil && !errors.Is(err, stdHTTP.ErrServerClosed) {
		return err
	}

	return nil
}
