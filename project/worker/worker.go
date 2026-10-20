package worker

import (
	"context"
	"tickets/entities"

	"github.com/google/uuid"
)

type Task int

const (
	TaskIssueReceipt Task = iota
	TaskAppendToTracker
)

type Message struct {
	Task     Task
	TicketID string
}

type Worker struct {
	queue chan Message

	spreadsheetsAPI SpreadsheetsAPI
	receiptsService ReceiptsService
}

type SpreadsheetsAPI interface {
	AppendRow(ctx context.Context, sheetName string, row []string) error
}

type ReceiptsService interface {
	IssueReceipt(ctx context.Context, payload *entities.TicketBookingConfirmed_v1) (entities.IssueReceiptResponse, error)
	VoidReceipt(ctx context.Context, payload *entities.RefundTicket_v1) error
}

type FileService interface {
	PutFile(ctx context.Context, fileId string, fileContent string) error
}

type DeadNationService interface {
	PostTicket(ctx context.Context, e *entities.BookingMade_v1, deadNationId uuid.UUID) error
}

type PaymentService interface {
	RefundPayment(ctx context.Context, refundPayment entities.PaymentRefund) error
}

func NewWorker(
	spreadsheetsAPI SpreadsheetsAPI,
	receiptsService ReceiptsService,
) *Worker {
	return &Worker{
		queue: make(chan Message, 100),

		spreadsheetsAPI: spreadsheetsAPI,
		receiptsService: receiptsService,
	}
}

func (w *Worker) Send(msgs ...Message) {
	for _, msg := range msgs {
		w.queue <- msg
	}
}

func (w *Worker) Run(ctx context.Context) {
	// for msg := range w.queue {
	// switch msg.Task {
	// case TaskIssueReceipt:
	// 	err := w.receiptsService.IssueReceipt(ctx, msg.TicketID)
	// 	if err != nil {
	// 		slog.With("error", err).Error("failed to issue the receipt")
	// 		w.Send(msg)
	// 	}
	// case TaskAppendToTracker:
	// 	err := w.spreadsheetsAPI.AppendRow(ctx, "tickets-to-print", []string{msg.TicketID})
	// 	if err != nil {
	// 		slog.With("error", err).Error("failed to append to tracker")
	// 		w.Send(msg)
	// 	}
	// }
	// }
}
