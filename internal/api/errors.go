package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// XeroError represents an error response from the Xero API.
type XeroError struct {
	StatusCode int
	ErrorNum   int       `json:"ErrorNumber"`
	Type       string    `json:"Type"`
	Message    string    `json:"Message"`
	Elements   []Element `json:"Elements"`
}

type Element struct {
	ValidationErrors []ValidationError `json:"ValidationErrors"`
}

type ValidationError struct {
	Message string `json:"Message"`
}

func (e *XeroError) Error() string {
	var b strings.Builder
	statusText := http.StatusText(e.StatusCode)
	if statusText == "" {
		statusText = "Unknown"
	}
	fmt.Fprintf(&b, "Error: %s (HTTP %d)", statusText, e.StatusCode)
	if e.Message != "" {
		fmt.Fprintf(&b, "\n  %s", e.Message)
	}
	for _, el := range e.Elements {
		for _, ve := range el.ValidationErrors {
			if ve.Message != "" {
				fmt.Fprintf(&b, "\n  • %s", ve.Message)
			}
		}
	}
	return b.String()
}

// ExitCode returns the appropriate exit code for this error.
func (e *XeroError) ExitCode() int {
	switch {
	case e.StatusCode == 401 || e.StatusCode == 403:
		return 2 // auth
	case e.StatusCode == 400 || e.Type == "ValidationException":
		return 3 // validation
	case e.StatusCode == 404:
		return 4 // not found
	case e.StatusCode == 429:
		return 5 // rate limited
	default:
		return 1 // general error
	}
}

// ParseXeroError parses an HTTP response body into a XeroError.
func ParseXeroError(statusCode int, body io.Reader) *XeroError {
	xerr := &XeroError{StatusCode: statusCode}
	data, err := io.ReadAll(body)
	if err != nil || len(data) == 0 {
		return xerr
	}

	_ = json.Unmarshal(data, xerr)

	if xerr.Message == "" {
		var simple struct {
			Detail  string `json:"Detail"`
			Title   string `json:"Title"`
			Message string `json:"message"`
		}
		if json.Unmarshal(data, &simple) == nil {
			if simple.Detail != "" {
				xerr.Message = simple.Detail
			} else if simple.Title != "" {
				xerr.Message = simple.Title
			} else if simple.Message != "" {
				xerr.Message = simple.Message
			}
		}
	}

	return xerr
}
