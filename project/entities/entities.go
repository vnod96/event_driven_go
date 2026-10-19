package entities

import (
	"time"

	"github.com/google/uuid"
)

type Money struct {
	Amount   string `json:"amount" db:"amount"`
	Currency string `json:"currency" db:"currency"`
}

type Ticket struct {
	TicketID      string `json:"ticket_id" db:"ticket_id"`
	CustomerEmail string `json:"customer_email" db:"customer_email"`
	Price         Money  `json:"price" db:"price"`
}

func ToTicket(confirmedEvent *TicketBookingConfirmed) *Ticket {
	return &Ticket{
		TicketID:      confirmedEvent.TicketID,
		CustomerEmail: confirmedEvent.CustomerEmail,
		Price:         confirmedEvent.Price,
	}
}

type AppendToTrackerPayload struct {
	TicketID      string `json:"ticket_id"`
	CustomerEmail string `json:"customer_email"`
	Price         Money  `json:"price"`
}

type IssueReceiptPayload struct {
	TicketID string `json:"ticket_id"`
	Price    Money  `json:"price"`
}

type TicketBookingConfirmed struct {
	Header MessageHeader `json:"header"`

	TicketID      string `json:"ticket_id"`
	CustomerEmail string `json:"customer_email"`
	Price         Money  `json:"price"`
}

type TicketBookingCanceled struct {
	Header MessageHeader `json:"header"`

	TicketID      string `json:"ticket_id"`
	CustomerEmail string `json:"customer_email"`
	Price         Money  `json:"price"`
}

type TicketPrinted struct {
	Header MessageHeader `json:"header"`

	TicketID string `json:"ticket_id"`
	FileName string `json:"file_name"`
}

type MessageHeader struct {
	ID             string    `json:"id"`
	PublishedAt    time.Time `json:"published_at"`
	IdempotencyKey string    `json:"idempotency_key"`
}

func NewMessageHeader() MessageHeader {
	return MessageHeader{
		ID:             uuid.NewString(),
		PublishedAt:    time.Now().UTC(),
		IdempotencyKey: uuid.NewString(),
	}
}

func NewMessageHeaderWithIdempotencyKey(idempotencyKey string) MessageHeader {
	return MessageHeader{
		ID:             uuid.NewString(),
		PublishedAt:    time.Now().UTC(),
		IdempotencyKey: idempotencyKey,
	}
}

type Show struct {
	ShowID          string    `json:"show_id" db:"show_id"`
	DeadNationId    string    `json:"dead_nation_id" db:"dead_nation_id"`
	NumberOfTickets int       `json:"number_of_tickets" db:"number_of_tickets"`
	StartTime       time.Time `json:"start_time" db:"start_time"`
	Title           string    `json:"title" db:"title"`
	Venue           string    `json:"venue" db:"venue"`
}

type Booking struct {
	BookingID string `json:"booking_id" db:"booking_id"`
	ShowID    string `json:"show_id" db:"show_id"`
	NumberOfTickets int `json:"number_of_tickets" db:"number_of_tickets"`
	CustomerEmail string `json:"customer_email" db:"customer_email"`
}

type BookingMade struct {
    Header MessageHeader `json:"header"`

    NumberOfTickets int    `json:"number_of_tickets"`
    BookingID       string `json:"booking_id"`
    CustomerEmail   string `json:"customer_email"`
    ShowID          string `json:"show_id"`
}

type RefundTicket struct {
	Header MessageHeader `json:"header"`

	TicketID string `json:"ticket_id"`
}

type PaymentRefund struct {
	TicketID       string
	RefundReason   string
	IdempotencyKey string
}