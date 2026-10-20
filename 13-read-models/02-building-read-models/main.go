package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/shopspring/decimal"
)

type InvoiceIssued struct {
	InvoiceID    string
	CustomerName string
	Amount       decimal.Decimal
	IssuedAt     time.Time
}

type InvoicePaymentReceived struct {
	PaymentID  string
	InvoiceID  string
	PaidAmount decimal.Decimal
	PaidAt     time.Time

	FullyPaid bool
}

type InvoiceVoided struct {
	InvoiceID string
	VoidedAt  time.Time
}

type InvoiceReadModel struct {
	InvoiceID    string
	CustomerName string
	Amount       decimal.Decimal
	IssuedAt     time.Time

	LastPaymentID string
	FullyPaid     bool
	PaidAmount    decimal.Decimal
	LastPaymentAt time.Time

	Voided   bool
	VoidedAt time.Time
}

type InvoiceReadModelStorage struct {
	invoices map[string]InvoiceReadModel
	payments map[string]string
}

func NewInvoiceReadModelStorage() *InvoiceReadModelStorage {
	return &InvoiceReadModelStorage{
		invoices: make(map[string]InvoiceReadModel),
		payments: make(map[string]string),
	}
}

func (s *InvoiceReadModelStorage) Invoices() []InvoiceReadModel {
	invoices := make([]InvoiceReadModel, 0, len(s.invoices))
	for _, invoice := range s.invoices {
		invoices = append(invoices, invoice)
	}
	return invoices
}

func (s *InvoiceReadModelStorage) InvoiceByID(id string) (InvoiceReadModel, bool) {
	invoice, ok := s.invoices[id]
	return invoice, ok
}

func (s *InvoiceReadModelStorage) OnInvoiceIssued(ctx context.Context, event *InvoiceIssued) error {
	invoice := InvoiceReadModel{
		InvoiceID:    event.InvoiceID,
		CustomerName: event.CustomerName,
		Amount:       event.Amount,
		IssuedAt:     event.IssuedAt,
	}

	s.invoices[event.InvoiceID] = invoice

	return nil
}

func (s *InvoiceReadModelStorage) OnInvoicePaymentReceived(ctx context.Context, event *InvoicePaymentReceived) error {
	invoice, ok := s.InvoiceByID(event.InvoiceID)
	if !ok {
		return fmt.Errorf("invoice %s not found", event.InvoiceID)
	}

	if s.payments[invoice.InvoiceID] == event.PaymentID {
		return nil
	}

	if invoice.FullyPaid {
		return nil
	}

	invoice.LastPaymentAt = event.PaidAt
	invoice.PaidAmount = invoice.PaidAmount.Add(event.PaidAmount)
	invoice.FullyPaid = event.FullyPaid

	s.payments[event.InvoiceID] = event.PaymentID

	s.invoices[event.InvoiceID] = invoice
	return nil
}

func (s *InvoiceReadModelStorage) OnInvoiceVoided(ctx context.Context, event *InvoiceVoided) error {
	invoice, ok := s.InvoiceByID(event.InvoiceID)
	if !ok {
		return fmt.Errorf("invoice %s not found", event.InvoiceID)
	}
	invoice.VoidedAt = event.VoidedAt
	invoice.Voided = true

	s.invoices[event.InvoiceID] = invoice
	return nil
}

func NewRouter(storage *InvoiceReadModelStorage, eventProcessorConfig cqrs.EventProcessorConfig, watermillLogger watermill.LoggerAdapter) (*message.Router, error) {
	router := message.NewDefaultRouter(watermillLogger)

	eventProcessor, err := cqrs.NewEventProcessorWithConfig(router, eventProcessorConfig)
	if err != nil {
		return nil, fmt.Errorf("could not create command processor: %w", err)
	}

	err = eventProcessor.AddHandlers(
		cqrs.NewEventHandler(
			"issue-handler",
			storage.OnInvoiceIssued,
		),
		cqrs.NewEventHandler(
			"payment-handler",
			storage.OnInvoicePaymentReceived,
		),
		cqrs.NewEventHandler(
			"void-handler",
			storage.OnInvoiceVoided,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("could not add event handlers: %w", err)
	}

	return router, nil
}
