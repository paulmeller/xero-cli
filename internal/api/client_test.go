package api

import (
	"io"
	"net/http"
	"net/url"
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

	c.SetDryRun(true)

	if !c.dryRun {
		t.Error("dryRun should be true after SetDryRun(true)")
	}

	transport, ok := httpClient.Transport.(*xeroTransport)
	if !ok {
		t.Fatal("transport not xeroTransport")
	}
	if !transport.dryRun {
		t.Error("transport.dryRun should be true")
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
