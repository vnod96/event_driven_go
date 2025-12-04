package tests_test

import (
	"context"
	"net/http"
	"os"
	"sync"
	"testing"
	"tickets/entities"
	"tickets/message"
	"tickets/service"
	"time"

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

	go func ()  {
		svc := service.New(
			spreadsheetsAPI,
			receiptService,
			rc,
		)

		err := svc.Run(ctx)
		assert.NoError(t, err)
	}()



	waitForHttpServer(t)
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
	lock sync.Mutex
	IssuedReceipts []entities.TicketBookingConfirmed
}

func (r *ReceiptsServiceStub) IssueReceipt(ctx context.Context, request entities.TicketBookingConfirmed) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.IssuedReceipts = append(r.IssuedReceipts, request)
	return nil
}

type SpreadsheetsAPIStub struct {
	lock sync.Mutex
	sheets map[string][][]string
}

func (s *SpreadsheetsAPIStub) AppendRow(ctx context.Context, sheetName string, row []string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.sheets[sheetName] = append(s.sheets[sheetName], row)

	return nil
}