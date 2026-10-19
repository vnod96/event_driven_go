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
	cmdConfig cqrs.CommandProcessorConfig,
	handler event.Handler,
	logger watermill.LoggerAdapter,
) *message.Router {
	router := message.NewDefaultRouter(logger)
	useMiddlewares(router, logger)
	outbox.AddForwarder(postgresSubscriber, publisher, router, logger)
	err := AddEventProcessor(router, handler, config)
	if err != nil {
		panic(err)
	}
	err = AddCommandProcessor(router, handler, cmdConfig)
	if err != nil {
		panic(err)
	}
	return router
}

func AddCommandProcessor(router *message.Router, handler event.Handler, cmdConfig cqrs.CommandProcessorConfig) error {
	c, err := cqrs.NewCommandProcessorWithConfig(router, cmdConfig)
	if err != nil {
		return err
	}

	err = c.AddHandlers(
		cqrs.NewCommandHandler(
			"refund-ticket-handler",
			handler.VoidReceipt,
		),
	)

	return nil
}

func AddEventProcessor(router *message.Router, handler event.Handler, config cqrs.EventProcessorConfig) error {
	p, err := cqrs.NewEventProcessorWithConfig(router, config)
	if err != nil {
		return err
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
	return err
}
