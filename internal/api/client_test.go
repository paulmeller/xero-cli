package api

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestNewClient_WrapsTransport(t *testing.T) {
	httpClient := &http.Client{}
	c := NewClient(httpClient, "tenant-123", false, false, nil)

	if c == nil {
		t.Fatal("NewClient returned nil")
	}

	transport, ok := httpClient.Transport.(*xeroTransport)
	if !ok {
		t.Fatal("transport was not wrapped with xeroTransport")
	}
	if transport.tenantID != "tenant-123" {
		t.Errorf("transport.tenantID = %q, want %q", transport.tenantID, "tenant-123")
	}
}

func TestNewClient_DefaultErrOut(t *testing.T) {
	httpClient := &http.Client{}
	c := NewClient(httpClient, "t", false, false, nil)

	// errOut should default to io.Discard when nil is passed
	if c.errOut != io.Discard {
		t.Error("errOut should default to io.Discard when nil is passed")
	}
}

func TestSetVerbose(t *testing.T) {
	httpClient := &http.Client{}
	c := NewClient(httpClient, "t", false, false, nil)

	errOut := io.Discard
	c.SetVerbose(true, errOut)

	if !c.verbose {
		t.Error("verbose should be true after SetVerbose(true)")
	}

	transport, ok := httpClient.Transport.(*xeroTransport)
	if !ok {
		t.Fatal("transport not xeroTransport")
	}
	if !transport.verbose {
		t.Error("transport.verbose should be true")
	}
}

func TestSetDryRun(t *testing.T) {
	httpClient := &http.Client{}
	c := NewClient(httpClient, "t", false, false, nil)

	c.SetDryRun(true, io.Discard)

	if !c.dryRun {
		t.Error("dryRun should be true after SetDryRun(true, ...)")
	}

	transport, ok := httpClient.Transport.(*xeroTransport)
	if !ok {
		t.Fatal("transport not xeroTransport")
	}
	if !transport.dryRun {
		t.Error("transport.dryRun should be true")
	}
}

func TestSetDryRun_NilErrOutPreservesExisting(t *testing.T) {
	httpClient := &http.Client{}
	c := NewClient(httpClient, "t", false, false, nil)
	var buf bytes.Buffer
	c.SetVerbose(true, &buf)

	// Passing nil must not clobber a writer set by an earlier SetVerbose/SetDryRun call.
	c.SetDryRun(true, nil)

	if c.errOut != &buf {
		t.Error("SetDryRun(dryRun, nil) must preserve the previously configured errOut")
	}
}

// Regression for the "dry-run without --verbose is silently swallowed" bug: NewClientFromConfig
// always constructs with io.Discard, and cmdutil.ApplyClientFlags used to call SetDryRun(true)
// with no way to route output anywhere else — so a `--dry-run` invocation without `--verbose`
// printed nothing and returned a fake 200 {}, indistinguishable from a real empty success.
func TestSetDryRun_OutputVisibleWithoutVerbose(t *testing.T) {
	httpClient := &http.Client{}
	c := NewClient(httpClient, "t", false, false, nil) // nil errOut -> io.Discard, verbose never set
	var buf bytes.Buffer

	c.SetDryRun(true, &buf) // simulates ApplyClientFlags's dry-run-only path

	_, err := c.Put(context.Background(), "Accounts", map[string]string{"Code": "200"}, "")
	if err != nil {
		t.Fatalf("Put returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "[dry-run] Would send PUT") {
		t.Errorf("dry-run output not written to the configured errOut; got %q", buf.String())
	}
}

func TestSetTenantID(t *testing.T) {
	httpClient := &http.Client{}
	c := NewClient(httpClient, "old-tenant", false, false, nil)

	c.SetTenantID("new-tenant")

	if c.tenantID != "new-tenant" {
		t.Errorf("tenantID = %q, want %q", c.tenantID, "new-tenant")
	}

	transport, ok := httpClient.Transport.(*xeroTransport)
	if !ok {
		t.Fatal("transport not xeroTransport")
	}
	if transport.tenantID != "new-tenant" {
		t.Errorf("transport.tenantID = %q, want %q", transport.tenantID, "new-tenant")
	}
}

func TestSetTimeout(t *testing.T) {
	httpClient := &http.Client{}
	c := NewClient(httpClient, "t", false, false, nil)

	c.SetTimeout(30 * time.Second)

	if c.timeout != 30*time.Second {
		t.Errorf("timeout = %v, want 30s", c.timeout)
	}
	if httpClient.Timeout != 30*time.Second {
		t.Errorf("httpClient.Timeout = %v, want 30s", httpClient.Timeout)
	}
}

func TestBuildURL_NoParams(t *testing.T) {
	httpClient := &http.Client{}
	c := NewClient(httpClient, "t", false, false, nil)

	got := c.buildURL("Invoices", nil)
	want := BaseURL + "/Invoices"
	if got != want {
		t.Errorf("buildURL = %q, want %q", got, want)
	}
}

func TestBuildURL_WithParams(t *testing.T) {
	httpClient := &http.Client{}
	c := NewClient(httpClient, "t", false, false, nil)

	params := url.Values{}
	params.Set("page", "1")
	params.Set("pageSize", "50")

	got := c.buildURL("Invoices", params)
	// Parse to avoid query param ordering issues
	u, err := url.Parse(got)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if u.Query().Get("page") != "1" {
		t.Errorf("page param = %q, want %q", u.Query().Get("page"), "1")
	}
	if u.Query().Get("pageSize") != "50" {
		t.Errorf("pageSize param = %q, want %q", u.Query().Get("pageSize"), "50")
	}
}

func TestBuildURL_LeadingSlashTrimmed(t *testing.T) {
	httpClient := &http.Client{}
	c := NewClient(httpClient, "t", false, false, nil)

	got := c.buildURL("/Invoices", nil)
	want := BaseURL + "/Invoices"
	if got != want {
		t.Errorf("buildURL = %q, want %q (leading slash should be trimmed)", got, want)
	}
}

// Regression for the "PUT drops Idempotency-Key" bug: the transport retries on network
// error/429/5xx and replays the exact same body via req.GetBody, so a PUT-create with no
// idempotency header has zero duplicate protection on retry — real money moved twice by a
// bank-transfers create is exactly the failure mode this header exists to prevent.
func TestPut_SetsIdempotencyKeyHeader(t *testing.T) {
	var gotKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("Idempotency-Key")
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := NewClient(&http.Client{}, "t", false, false, nil)
	c.baseURL = srv.URL

	if _, err := c.Put(context.Background(), "BankTransfers", map[string]string{"Amount": "100"}, ""); err != nil {
		t.Fatalf("Put returned error: %v", err)
	}
	if gotKey == "" {
		t.Error("Put sent no Idempotency-Key header — a retried PUT can now create a duplicate resource")
	}
}

func TestPut_HonoursCallerSuppliedIdempotencyKey(t *testing.T) {
	var gotKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("Idempotency-Key")
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := NewClient(&http.Client{}, "t", false, false, nil)
	c.baseURL = srv.URL

	if _, err := c.Put(context.Background(), "BankTransfers", map[string]string{"Amount": "100"}, "my-fixed-key"); err != nil {
		t.Fatalf("Put returned error: %v", err)
	}
	if gotKey != "my-fixed-key" {
		t.Errorf("Idempotency-Key = %q, want the caller-supplied key %q", gotKey, "my-fixed-key")
	}
}

func TestPutRaw_SetsIdempotencyKeyHeader(t *testing.T) {
	var gotKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("Idempotency-Key")
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := NewClient(&http.Client{}, "t", false, false, nil)
	c.baseURL = srv.URL

	if _, err := c.PutRaw(context.Background(), "BankTransfers", []byte(`{"Amount":100}`), ""); err != nil {
		t.Fatalf("PutRaw returned error: %v", err)
	}
	if gotKey == "" {
		t.Error("PutRaw sent no Idempotency-Key header")
	}
}

// Regression for the "--modified-since sent as query param" bug: Xero only honours this filter
// as a real If-Modified-Since HTTP header. BuildListParams stashes the value under that same
// key in url.Values as a carrier; Get must pull it back out onto the header, not leave it as a
// literal "?If-Modified-Since=..." query string param (which Xero silently ignores).
func TestGet_ModifiedSinceSentAsHeaderNotQueryParam(t *testing.T) {
	var gotHeader, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("If-Modified-Since")
		gotQuery = r.URL.Query().Get("If-Modified-Since")
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := NewClient(&http.Client{}, "t", false, false, nil)
	c.baseURL = srv.URL

	params := url.Values{}
	params.Set("If-Modified-Since", "2026-01-01T00:00:00Z")
	params.Set("where", `Status=="ACTIVE"`)

	if _, err := c.Get(context.Background(), "Invoices", params); err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if gotHeader != "2026-01-01T00:00:00Z" {
		t.Errorf("If-Modified-Since header = %q, want the configured date", gotHeader)
	}
	if gotQuery != "" {
		t.Errorf("If-Modified-Since leaked into the query string as %q; Xero ignores it there", gotQuery)
	}
	// The caller's own params map must not be mutated - Get clones before stripping the key.
	if params.Get("If-Modified-Since") == "" {
		t.Error("Get must not mutate the caller-supplied params map")
	}
}

func TestNewClient_NilTransport(t *testing.T) {
	httpClient := &http.Client{Transport: nil}
	c := NewClient(httpClient, "t", false, false, nil)

	transport, ok := httpClient.Transport.(*xeroTransport)
	if !ok {
		t.Fatal("transport not xeroTransport")
	}
	// When original transport is nil, base should be http.DefaultTransport
	if transport.base != http.DefaultTransport {
		t.Error("base transport should be http.DefaultTransport when original is nil")
	}
	_ = c // use c
}
