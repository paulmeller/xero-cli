package auth

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type CallbackServer struct {
	redirectURI string
	state       string
	server      *http.Server
	listener    net.Listener
	codeCh      chan string
	errCh       chan error
}

func NewCallbackServer(redirectURI string, state string) *CallbackServer {
	return &CallbackServer{
		redirectURI: redirectURI,
		state:       state,
		codeCh:      make(chan string, 1),
		errCh:       make(chan error, 1),
	}
}

func (s *CallbackServer) Start() error {
	u, err := url.Parse(s.redirectURI)
	if err != nil {
		return fmt.Errorf("invalid redirect URI: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc(u.Path, s.handleCallback)

	s.server = &http.Server{
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	s.listener, err = net.Listen("tcp", u.Host)
	if err != nil {
		return fmt.Errorf("cannot listen on %s: %w", u.Host, err)
	}

	go func() {
		if err := s.server.Serve(s.listener); err != nil && err != http.ErrServerClosed {
			s.errCh <- err
		}
	}()

	return nil
}

func (s *CallbackServer) handleCallback(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	if errMsg := query.Get("error"); errMsg != "" {
		desc := query.Get("error_description")
		s.errCh <- fmt.Errorf("authorization failed: %s: %s", errMsg, desc)
		http.Error(w, "Authorization failed. You can close this window.", http.StatusBadRequest)
		return
	}

	state := query.Get("state")
	if state != s.state {
		s.errCh <- fmt.Errorf("state mismatch")
		http.Error(w, "State mismatch. You can close this window.", http.StatusBadRequest)
		return
	}

	code := query.Get("code")
	if code == "" {
		s.errCh <- fmt.Errorf("no code in callback")
		http.Error(w, "No authorization code received. You can close this window.", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `<html><body><h2>Authentication successful!</h2><p>You can close this window and return to the terminal.</p></body></html>`)

	s.codeCh <- code
}

func (s *CallbackServer) WaitForCode(ctx context.Context) (string, error) {
	select {
	case code := <-s.codeCh:
		return code, nil
	case err := <-s.errCh:
		return "", err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func (s *CallbackServer) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}

// ExtractCode parses a callback URL and extracts the authorization code.
func ExtractCode(callbackURL string, expectedState string) (string, error) {
	callbackURL = strings.TrimSpace(callbackURL)
	u, err := url.Parse(callbackURL)
	if err != nil {
		return "", fmt.Errorf("invalid callback URL: %w", err)
	}

	if errMsg := u.Query().Get("error"); errMsg != "" {
		desc := u.Query().Get("error_description")
		return "", fmt.Errorf("authorization failed: %s: %s", errMsg, desc)
	}

	state := u.Query().Get("state")
	if state != expectedState {
		return "", fmt.Errorf("state mismatch in callback URL")
	}

	code := u.Query().Get("code")
	if code == "" {
		return "", fmt.Errorf("no authorization code in callback URL")
	}

	return code, nil
}
