package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"

	"github.com/paulmeller/xero-cli/internal/auth"
	"github.com/paulmeller/xero-cli/internal/config"
)

const (
	BaseURL = "https://api.xero.com/api.xro/2.0"
)

// Client is the Xero API client.
type Client struct {
	http     *http.Client
	baseURL  string
	tenantID string
	verbose  bool
	dryRun   bool
	errOut   io.Writer
	timeout  time.Duration
}

type xeroTransport struct {
	base     http.RoundTripper
	tenantID string
	verbose  bool
	dryRun   bool
	errOut   io.Writer
	retries  int
}

func (t *xeroTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Xero-Tenant-Id", t.tenantID)
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "application/json")
	}

	if t.verbose {
		fmt.Fprintf(t.errOut, "> %s %s\n", req.Method, req.URL.String())
		for k, v := range req.Header {
			if !strings.EqualFold(k, "Authorization") {
				fmt.Fprintf(t.errOut, "> %s: %s\n", k, strings.Join(v, ", "))
			}
		}
		fmt.Fprintln(t.errOut)
	}

	if t.dryRun && req.Method != http.MethodGet {
		fmt.Fprintf(t.errOut, "[dry-run] Would send %s %s\n", req.Method, req.URL.String())
		if req.Body != nil {
			body, _ := io.ReadAll(req.Body)
			req.Body.Close()
			fmt.Fprintf(t.errOut, "[dry-run] Body: %s\n", string(body))
		}
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(`{}`)),
			Header:     http.Header{},
		}, nil
	}

	var resp *http.Response
	var err error
	maxRetries := t.retries
	if maxRetries == 0 {
		maxRetries = 3
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			if req.GetBody != nil {
				req.Body, _ = req.GetBody()
			}
		}

		resp, err = t.base.RoundTrip(req)
		if err != nil {
			if attempt < maxRetries {
				time.Sleep(time.Duration(1<<uint(attempt)) * time.Second)
				continue
			}
			return nil, err
		}

		if t.verbose {
			fmt.Fprintf(t.errOut, "< %s\n", resp.Status)
			if v := resp.Header.Get("X-MinLimit-Remaining"); v != "" {
				fmt.Fprintf(t.errOut, "< Rate limit remaining (minute): %s\n", v)
			}
			if v := resp.Header.Get("X-DayLimit-Remaining"); v != "" {
				fmt.Fprintf(t.errOut, "< Rate limit remaining (day): %s\n", v)
			}
			fmt.Fprintln(t.errOut)
		}

		if resp.StatusCode == 429 {
			if attempt < maxRetries {
				retryAfter := resp.Header.Get("Retry-After")
				wait := 60 * time.Second
				if secs, parseErr := strconv.Atoi(retryAfter); parseErr == nil && secs > 0 {
					wait = time.Duration(secs) * time.Second
				}
				if t.verbose {
					fmt.Fprintf(t.errOut, "Rate limited, waiting %s before retry...\n", wait)
				}
				resp.Body.Close()
				time.Sleep(wait)
				continue
			}
		}

		if resp.StatusCode >= 500 {
			if attempt < maxRetries {
				resp.Body.Close()
				wait := time.Duration(1<<uint(attempt)) * time.Second
				if t.verbose {
					fmt.Fprintf(t.errOut, "Server error %d, retrying in %s...\n", resp.StatusCode, wait)
				}
				time.Sleep(wait)
				continue
			}
		}

		break
	}

	return resp, err
}

// NewClient creates a new Xero API client.
func NewClient(httpClient *http.Client, tenantID string, verbose, dryRun bool, errOut io.Writer) *Client {
	if errOut == nil {
		errOut = io.Discard
	}

	transport := httpClient.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	httpClient.Transport = &xeroTransport{
		base:     transport,
		tenantID: tenantID,
		verbose:  verbose,
		dryRun:   dryRun,
		errOut:   errOut,
		retries:  3,
	}

	return &Client{
		http:     httpClient,
		baseURL:  BaseURL,
		tenantID: tenantID,
		verbose:  verbose,
		dryRun:   dryRun,
		errOut:   errOut,
	}
}

// NewClientFromConfig creates a Client from the loaded config.
func NewClientFromConfig(cfg *config.Config) (*Client, error) {
	conn := cfg.ActiveConn()
	connName := cfg.ActiveConnectionName()

	if conn.ClientID == "" {
		return nil, fmt.Errorf("client ID not configured; set XERO_CLIENT_ID or configure in config.toml")
	}
	if conn.ActiveTenant == "" {
		return nil, fmt.Errorf("no active tenant; run 'xero tenants switch' or set XERO_TENANT_ID")
	}

	var tokenSource oauth2.TokenSource
	if conn.GrantType == "client_credentials" {
		tokenSource = auth.ClientCredentialsTokenSource(context.Background(), conn)
	} else {
		underlying, err := auth.LoadToken(connName)
		if err != nil {
			return nil, fmt.Errorf("not authenticated; run 'xero auth login': %w", err)
		}
		oauthCfg := auth.OAuthConfig(conn)
		// Use WithConfig so PersistentTokenSource can rebuild the token source
		// with the latest refresh token from storage (Xero tokens are single-use).
		pts := auth.NewPersistentTokenSourceWithConfig(oauthCfg, underlying, connName)
		httpClient := oauth2.NewClient(context.Background(), pts)
		return NewClient(httpClient, conn.ActiveTenant, false, false, io.Discard), nil
	}

	pts := auth.NewPersistentTokenSource(tokenSource, connName)
	httpClient := oauth2.NewClient(context.Background(), pts)

	return NewClient(httpClient, conn.ActiveTenant, false, false, io.Discard), nil
}

// NewClientFromToken creates a Client using an externally-provided access token.
// The token is used as-is with no refresh capability.
func NewClientFromToken(token string, tenantID string) *Client {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: token,
		TokenType:   "Bearer",
	})
	httpClient := oauth2.NewClient(context.Background(), tokenSource)
	return NewClient(httpClient, tenantID, false, false, io.Discard)
}

// SetVerbose enables verbose logging.
func (c *Client) SetVerbose(verbose bool, errOut io.Writer) {
	c.verbose = verbose
	c.errOut = errOut
	if t, ok := c.http.Transport.(*xeroTransport); ok {
		t.verbose = verbose
		t.errOut = errOut
	}
}

// SetDryRun enables dry-run mode. errOut is where "[dry-run] Would send ..." lines are
// written; it takes its own writer rather than reusing whatever SetVerbose set (or the
// io.Discard every constructor defaults to), because dry-run output must be visible whether
// or not --verbose is also set — previously a dry-run without --verbose wrote to io.Discard
// and returned a fake 200 {}, which was indistinguishable from a real empty response.
func (c *Client) SetDryRun(dryRun bool, errOut io.Writer) {
	c.dryRun = dryRun
	if errOut != nil {
		c.errOut = errOut
	}
	if t, ok := c.http.Transport.(*xeroTransport); ok {
		t.dryRun = dryRun
		if errOut != nil {
			t.errOut = errOut
		}
	}
}

// SetTimeout sets the request timeout.
func (c *Client) SetTimeout(d time.Duration) {
	c.timeout = d
	c.http.Timeout = d
}

// SetBaseURL overrides the API base URL (default: BaseURL). Used by tests in other packages to
// point the client at an httptest server, since baseURL is otherwise unexported.
func (c *Client) SetBaseURL(url string) {
	c.baseURL = url
}

// SetTenantID overrides the tenant ID.
func (c *Client) SetTenantID(id string) {
	c.tenantID = id
	if t, ok := c.http.Transport.(*xeroTransport); ok {
		t.tenantID = id
	}
}

// GetPDF performs a GET request with Accept: application/pdf and returns the raw bytes.
func (c *Client) GetPDF(ctx context.Context, path string) ([]byte, error) {
	u := c.buildURL(path, nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/pdf")
	return c.doBytes(req)
}

// Get performs a GET request and returns raw JSON.
func (c *Client) Get(ctx context.Context, path string, params url.Values) (json.RawMessage, error) {
	req, err := c.newGetRequest(ctx, path, params)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

// GetWithHeaders performs a GET request and returns raw JSON plus response headers.
func (c *Client) GetWithHeaders(ctx context.Context, path string, params url.Values) (json.RawMessage, http.Header, error) {
	req, err := c.newGetRequest(ctx, path, params)
	if err != nil {
		return nil, nil, err
	}
	return c.doWithHeaders(req)
}

// newGetRequest builds a GET request. Xero's modified-since filter is a real HTTP header
// (If-Modified-Since), not a query parameter - BuildListParams stashes the value under the
// "If-Modified-Since" key in params as a convenient carrier, and this pulls it back out onto
// the request header where Xero actually reads it. Previously it was sent as a literal
// "?If-Modified-Since=..." query string parameter, which Xero silently ignores, making
// --modified-since a silent no-op.
func (c *Client) newGetRequest(ctx context.Context, path string, params url.Values) (*http.Request, error) {
	var ifModifiedSince string
	if params != nil {
		if v := params.Get("If-Modified-Since"); v != "" {
			cloned := make(url.Values, len(params))
			for k, vals := range params {
				cloned[k] = append([]string(nil), vals...)
			}
			cloned.Del("If-Modified-Since")
			params = cloned
			ifModifiedSince = v
		}
	}

	u := c.buildURL(path, params)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	if ifModifiedSince != "" {
		req.Header.Set("If-Modified-Since", ifModifiedSince)
	}
	return req, nil
}

// GetRaw performs a GET to an arbitrary URL (e.g. /connections).
func (c *Client) GetRaw(ctx context.Context, rawURL string) (json.RawMessage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	return c.doRaw(req)
}

// GetConnections fetches /connections.
func (c *Client) GetConnections(ctx context.Context) (json.RawMessage, error) {
	return c.GetRaw(ctx, "https://api.xero.com/connections")
}

// Post performs a POST request with JSON body and idempotency key.
func (c *Client) Post(ctx context.Context, path string, body any, idempotencyKey string) (json.RawMessage, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal request body: %w", err)
	}

	u := c.buildURL(path, nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(data)), nil
	}

	if idempotencyKey == "" {
		idempotencyKey = uuid.New().String()
	}
	req.Header.Set("Idempotency-Key", idempotencyKey)

	return c.do(req)
}

// PostRaw posts raw JSON bytes with idempotency key.
func (c *Client) PostRaw(ctx context.Context, path string, data json.RawMessage, idempotencyKey string) (json.RawMessage, error) {
	u := c.buildURL(path, nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(data)), nil
	}

	if idempotencyKey == "" {
		idempotencyKey = uuid.New().String()
	}
	req.Header.Set("Idempotency-Key", idempotencyKey)

	return c.do(req)
}

// Put performs a PUT request with JSON body and idempotency key. Xero supports
// Idempotency-Key on PUT-create endpoints (e.g. createAccount, createBankTransfer) the same
// way it does on POST; the transport retries on network errors/429/5xx and replays this exact
// body via GetBody, so a request without a stable key risks duplicate creation on retry.
func (c *Client) Put(ctx context.Context, path string, body any, idempotencyKey string) (json.RawMessage, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal request body: %w", err)
	}

	u := c.buildURL(path, nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(data)), nil
	}

	if idempotencyKey == "" {
		idempotencyKey = uuid.New().String()
	}
	req.Header.Set("Idempotency-Key", idempotencyKey)

	return c.do(req)
}

// PutRaw puts raw JSON bytes with idempotency key. See Put for why this is required on retry.
func (c *Client) PutRaw(ctx context.Context, path string, data json.RawMessage, idempotencyKey string) (json.RawMessage, error) {
	u := c.buildURL(path, nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(data)), nil
	}

	if idempotencyKey == "" {
		idempotencyKey = uuid.New().String()
	}
	req.Header.Set("Idempotency-Key", idempotencyKey)

	return c.do(req)
}

// PutAttachment uploads a file attachment.
func (c *Client) PutAttachment(ctx context.Context, path string, data []byte, contentType string) (json.RawMessage, error) {
	u := c.buildURL(path, nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(data)), nil
	}
	return c.do(req)
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, path string) (json.RawMessage, error) {
	u := c.buildURL(path, nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

func (c *Client) buildURL(path string, params url.Values) string {
	u := c.baseURL + "/" + strings.TrimPrefix(path, "/")
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	return u
}

func (c *Client) do(req *http.Request) (json.RawMessage, error) {
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		xerr := ParseXeroError(resp.StatusCode, bytes.NewReader(body))
		return nil, xerr
	}

	if len(body) == 0 {
		return json.RawMessage("{}"), nil
	}

	return json.RawMessage(body), nil
}

func (c *Client) doBytes(req *http.Request) ([]byte, error) {
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		xerr := ParseXeroError(resp.StatusCode, bytes.NewReader(body))
		return nil, xerr
	}

	return body, nil
}

func (c *Client) doRaw(req *http.Request) (json.RawMessage, error) {
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		xerr := ParseXeroError(resp.StatusCode, bytes.NewReader(body))
		return nil, xerr
	}

	return json.RawMessage(body), nil
}

func (c *Client) doWithHeaders(req *http.Request) (json.RawMessage, http.Header, error) {
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		xerr := ParseXeroError(resp.StatusCode, bytes.NewReader(body))
		return nil, resp.Header, xerr
	}

	if len(body) == 0 {
		return json.RawMessage("{}"), resp.Header, nil
	}

	return json.RawMessage(body), resp.Header, nil
}
