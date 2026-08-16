package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// Regression for the "bank-transfers list --all infinite loop" bug: some Xero collections
// (BankTransfers) have no page/pageSize support at all and return the full result set on every
// call, so len(arr) < pageSize never fires. PaginateAll must detect the stall (identical
// consecutive "pages") and stop instead of re-fetching forever.
func TestPaginateAll_NonPaginatingEndpointStalls(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		// Always return the same 150 items regardless of page/pageSize - simulates an endpoint
		// that silently ignores pagination params.
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, buildItemsJSON("Things", 150))
	}))
	defer srv.Close()

	c := NewClient(&http.Client{}, "t", false, false, nil)
	c.baseURL = srv.URL

	result, err := PaginateAll(context.Background(), c, "Things", nil, "Things", 100)
	if err != nil {
		t.Fatalf("PaginateAll returned error: %v", err)
	}
	if len(result.Array()) != 150 {
		t.Errorf("got %d items, want 150 (deduped to one page's worth)", len(result.Array()))
	}
	if calls != 2 {
		t.Errorf("expected exactly 2 requests (initial + one stall-detecting repeat), got %d", calls)
	}
}

// Regression for the "--page-size 0" bug: len(arr) < pageSize becomes len(arr) < 0 (never true)
// when pageSize is 0, breaking termination for every resource, not just non-paginating ones.
func TestPaginateAll_ZeroPageSizeStillTerminates(t *testing.T) {
	page := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		w.Header().Set("Content-Type", "application/json")
		if page == 1 {
			fmt.Fprint(w, buildItemsJSON("Things", 100)) // full page - default effective size
		} else {
			fmt.Fprint(w, buildItemsJSON("Things", 10)) // short page - should stop here
		}
	}))
	defer srv.Close()

	c := NewClient(&http.Client{}, "t", false, false, nil)
	c.baseURL = srv.URL

	result, err := PaginateAll(context.Background(), c, "Things", nil, "Things", 0)
	if err != nil {
		t.Fatalf("PaginateAll returned error: %v", err)
	}
	if len(result.Array()) != 110 {
		t.Errorf("got %d items, want 110 (100 + 10, terminating on the short page)", len(result.Array()))
	}
	if page != 2 {
		t.Errorf("expected exactly 2 requests, got %d", page)
	}
}

func TestPaginateAll_NormalMultiPageFetch(t *testing.T) {
	page := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		w.Header().Set("Content-Type", "application/json")
		// Offset IDs by page so consecutive pages are never byte-identical - real Xero pages
		// return different records, unlike the deliberately-static fixture in the stall test.
		if page < 3 {
			fmt.Fprint(w, buildItemsJSONFrom("Things", page*10, 2))
		} else {
			fmt.Fprint(w, buildItemsJSONFrom("Things", page*10, 1))
		}
	}))
	defer srv.Close()

	c := NewClient(&http.Client{}, "t", false, false, nil)
	c.baseURL = srv.URL

	result, err := PaginateAll(context.Background(), c, "Things", url.Values{}, "Things", 2)
	if err != nil {
		t.Fatalf("PaginateAll returned error: %v", err)
	}
	if len(result.Array()) != 5 {
		t.Errorf("got %d items, want 5 (2+2+1)", len(result.Array()))
	}
}

// buildItemsJSON generates n items under the given key, always starting at ID 0 - used where
// the test deliberately wants every "page" to be byte-identical.
func buildItemsJSON(key string, n int) string {
	return buildItemsJSONFrom(key, 0, n)
}

// buildItemsJSONFrom generates n items under the given key with IDs starting at `from`, so
// distinct pages of real pagination are never byte-identical to each other.
func buildItemsJSONFrom(key string, from, n int) string {
	s := fmt.Sprintf(`{"%s":[`, key)
	for i := 0; i < n; i++ {
		if i > 0 {
			s += ","
		}
		s += fmt.Sprintf(`{"ID":%d}`, from+i)
	}
	s += "]}"
	return s
}
