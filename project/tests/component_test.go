package tests_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"testing"
	"tickets/entities"
	ticketsHttp "tickets/http"
	"tickets/message"
	"tickets/service"
	"time"

	"github.com/lithammer/shortuuid/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComponent(t *testing.T) {
	rc := message.NewRedisClient(os.Getenv("REDIS_ADDR"))
	defer rc.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	spreadsheetsAPI := &SpreadsheetsAPIStub{}
	receiptService := &ReceiptsServiceStub{}

	go func() {
		svc := service.New(
			spreadsheetsAPI,
			receiptService,
			rc,
		)

		err := svc.Run(ctx)
		assert.NoError(t, err)
	}()

	waitForHttpServer(t)
	ticket := ticketsHttp.TicketStatusRequest{
		TicketID: shortuuid.New(),
		Status:   "confirmed",
		Price: entities.Money{
			Amount:   "2",
			Currency: "USD",
		},
		CustomerEmail: "v.x@tdl.com",
	}
	sendTicketsStatus(t, ticketsHttp.TicketsStatusRequest{ Tickets: []ticketsHttp.TicketStatusRequest{ticket} })

	assertReceiptForTicketIssued(t, receiptService, ticket)
	assertTicketInSpreadsheet(t, spreadsheetsAPI, "tickets-to-print", ticket)

	ticket.Status = "canceled"
	sendTicketsStatus(t, ticketsHttp.TicketsStatusRequest{ Tickets: []ticketsHttp.TicketStatusRequest{ticket} })
	assertTicketInSpreadsheet(t, spreadsheetsAPI, "tickets-to-refund", ticket)
}

func waitForHttpServer(t *testing.T) {
	t.Helper()

	require.EventuallyWithT(
		t,
		func(t *assert.CollectT) {
			resp, err := http.Get("http://localhost:8080/health")
			if !assert.NoError(t, err) {
				return
			}
			defer resp.Body.Close()

			if assert.Less(t, resp.StatusCode, 300, "API not ready, http status: %d", resp.StatusCode) {
				return
			}
		},
		time.Second*10,
		time.Millisecond*50,
	)
}

type ReceiptsServiceStub struct {
	lock           sync.Mutex
	IssuedReceipts []entities.TicketBookingConfirmed
}

func (r *ReceiptsServiceStub) IssueReceipt(ctx context.Context, request entities.TicketBookingConfirmed) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.IssuedReceipts = append(r.IssuedReceipts, request)
	return nil
}

type SpreadsheetsAPIStub struct {
	lock   sync.Mutex
	sheets map[string][][]string
}

func (s *SpreadsheetsAPIStub) AppendRow(ctx context.Context, sheetName string, row []string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.sheets == nil {
		s.sheets = make(map[string][][]string)
	}

	s.sheets[sheetName] = append(s.sheets[sheetName], row)

	return nil
}

func sendTicketsStatus(t *testing.T, req ticketsHttp.TicketsStatusRequest) {
	t.Helper()

	payload, err := json.Marshal(req)
	require.NoError(t, err)

	correlationID := shortuuid.New()

	httpReq, err := http.NewRequest(
		http.MethodPost,
		"http://localhost:8080/tickets-status",
		bytes.NewBuffer(payload),
	)

	require.NoError(t, err)

	httpReq.Header.Set("Correlation-ID", correlationID)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func assertReceiptForTicketIssued(t *testing.T, receiptService *ReceiptsServiceStub, ticket ticketsHttp.TicketStatusRequest) {
	t.Helper()

	parentT := t

	assert.EventuallyWithT(t,
		func(t *assert.CollectT) {
			issueReceipts := len(receiptService.IssuedReceipts)
			parentT.Log("issued receipts", issueReceipts)

			assert.Greater(t, issueReceipts, 0, "no receipts issued")
		},
		10*time.Second, 100*time.Millisecond)

	var receipt entities.TicketBookingConfirmed
	var ok bool

	for _, r := range receiptService.IssuedReceipts {
		if r.TicketID != ticket.TicketID {
			continue
		}

		receipt = r
		ok = true
		break
	}

	require.Truef(t, ok, "receipt for %s not found", ticket.TicketID)
	assert.Equal(t, receipt.TicketID, ticket.TicketID)
	assert.Equal(t, receipt.Price.Amount, ticket.Price.Amount)
	assert.Equal(t, receipt.Price.Currency, ticket.Price.Currency)

}

func assertTicketInSpreadsheet(t *testing.T, spreadsheetAPI *SpreadsheetsAPIStub, sheetName string, ticket ticketsHttp.TicketStatusRequest){
	t.Helper()

	parentT := t

	require.EventuallyWithT(t, 
	func(t *assert.CollectT) {
		rows := len(spreadsheetAPI.sheets[sheetName])
		parentT.Log("rows added", rows)
		assert.Greater(t, rows, 0, "no rows added.")
	}, 10 * time.Second, 100 * time.Millisecond)

	var row []string
	var ok bool

	for _, r := range spreadsheetAPI.sheets[sheetName] {
		if r[0] != ticket.TicketID {
			continue
		}

		row = r
		ok = true
		break
	}

	require.Truef(t, ok , "row not found for ticket %s", ticket.TicketID)
	assert.Equal(t, row[0], ticket.TicketID)
	assert.Equal(t, row[1], ticket.CustomerEmail)
	assert.Equal(t, row[2], ticket.Price.Amount)
	assert.Equal(t, row[3], ticket.Price.Currency)

}