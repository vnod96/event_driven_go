package message

import (
	"tickets/message/event"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
)

func NewWatermillRouter(
	config cqrs.EventProcessorConfig,
	handler event.Handler,
	logger watermill.LoggerAdapter,
) *message.Router {
	router := message.NewDefaultRouter(logger)
	useMiddlewares(router, logger)
	p, err := cqrs.NewEventProcessorWithConfig(router, config)
	if err != nil {
		panic(err)
	}
	err = p.AddHandlers(
		cqrs.NewEventHandler(
			"issue-receipt-handler",
			handler.IssueReceipt,
		),
		cqrs.NewEventHandler(
			"append-to-tracker-handler",
			handler.AppendRow,
		),
		cqrs.NewEventHandler(
			"ticket-canceled-handler",
			handler.RefundTicket,
		),
		cqrs.NewEventHandler(
			"ticket-saver",
			handler.SaveTicket,
		),
		cqrs.NewEventHandler(
			"ticket-remover",
			handler.RemoveTicket,
		),
	)
	if err != nil {
		panic(err)
	}
	return router
}
