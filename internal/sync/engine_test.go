package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/paulmeller/xero-cli/internal/api"
)

type fakeDestination struct {
	written map[string][]json.RawMessage
}

func newFakeDestination() *fakeDestination {
	return &fakeDestination{written: make(map[string][]json.RawMessage)}
}

func (d *fakeDestination) Init(ctx context.Context) error { return nil }
func (d *fakeDestination) Write(ctx context.Context, stream, primaryKey string, records []json.RawMessage) (int, error) {
	d.written[stream] = append(d.written[stream], records...)
	return len(records), nil
}
func (d *fakeDestination) Close() error { return nil }

// Regression for the infinite-loop risk newly introduced by adding "bank_transfers" to
// StreamRegistry: Xero's GET /BankTransfers has no page/pageSize support, so it returns the
// full result set on every call. Without the stall check, syncStream's own pagination loop
// would re-fetch the same "page" forever, exactly like the bug fixed in api.PaginateAll.
func TestSyncStream_NonPaginatingEndpointStalls(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		// Always return a full page (>= the hardcoded pageSize of 100) regardless of the
		// page/pageSize params - simulates an endpoint that silently ignores pagination.
		var items []string
		for i := 0; i < 100; i++ {
			items = append(items, fmt.Sprintf(`{"BankTransferID":"id-%d"}`, i))
		}
		fmt.Fprintf(w, `{"BankTransfers":[%s]}`, strings.Join(items, ","))
	}))
	defer srv.Close()

	client := api.NewClient(&http.Client{}, "t", false, false, io.Discard)
	client.SetBaseURL(srv.URL)

	dest := newFakeDestination()
	cfg := &SyncConfig{}
	state := &SyncState{Streams: map[string]StreamState{}}
	e := NewEngine(client, cfg, state, dest, io.Discard, false)

	sc := StreamConfig{Name: "bank_transfers", Enabled: true, SyncMode: "full_refresh", PrimaryKey: "BankTransferID"}
	if err := e.syncStream(context.Background(), sc); err != nil {
		t.Fatalf("syncStream returned error: %v", err)
	}

	if calls != 2 {
		t.Errorf("expected exactly 2 requests (initial + one stall-detecting repeat), got %d", calls)
	}
	if len(dest.written["bank_transfers"]) != 100 {
		t.Errorf("expected 100 records written (one page's worth, deduped), got %d", len(dest.written["bank_transfers"]))
	}
}
