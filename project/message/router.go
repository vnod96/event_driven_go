package message

import (
	"tickets/message/event"
	"tickets/outbox"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
)

func NewWatermillRouter(
	postgresSubscriber message.Subscriber,
	publisher message.Publisher,
	config cqrs.EventProcessorConfig,
	handler event.Handler,
	logger watermill.LoggerAdapter,
) *message.Router {
	router := message.NewDefaultRouter(logger)
	useMiddlewares(router, logger)
	outbox.AddForwarder(postgresSubscriber, publisher, router, logger)
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
		cqrs.NewEventHandler(
			"print-ticket",
			handler.PrintFile,
		),
		cqrs.NewEventHandler(
			"make-booking",
			handler.MakeBooking,
		),
	)
	if err != nil {
		panic(err)
	}
	return router
}
