package api

import (
	"strings"
	"testing"
)

func TestParseXeroError_StatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantCode   int
		wantSubstr string
	}{
		{
			name:       "400 bad request with message",
			statusCode: 400,
			body:       `{"Message":"Invalid date format"}`,
			wantCode:   3,
			wantSubstr: "Bad Request",
		},
		{
			name:       "401 unauthorized",
			statusCode: 401,
			body:       `{"Message":"Token expired"}`,
			wantCode:   2,
			wantSubstr: "Unauthorized",
		},
		{
			name:       "404 not found",
			statusCode: 404,
			body:       `{"Message":"Invoice not found"}`,
			wantCode:   4,
			wantSubstr: "Not Found",
		},
		{
			name:       "429 rate limited",
			statusCode: 429,
			body:       `{"Message":"Rate limit exceeded"}`,
			wantCode:   5,
			wantSubstr: "Too Many Requests",
		},
		{
			name:       "500 server error",
			statusCode: 500,
			body:       `{"Message":"Internal error"}`,
			wantCode:   1,
			wantSubstr: "Internal Server Error",
		},
		{
			name:       "403 forbidden",
			statusCode: 403,
			body:       `{}`,
			wantCode:   2,
			wantSubstr: "Forbidden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xerr := ParseXeroError(tt.statusCode, strings.NewReader(tt.body))
			if xerr.StatusCode != tt.statusCode {
				t.Errorf("StatusCode = %d, want %d", xerr.StatusCode, tt.statusCode)
			}
			if xerr.ExitCode() != tt.wantCode {
				t.Errorf("ExitCode() = %d, want %d", xerr.ExitCode(), tt.wantCode)
			}
			errStr := xerr.Error()
			if !strings.Contains(errStr, tt.wantSubstr) {
				t.Errorf("Error() = %q, want substring %q", errStr, tt.wantSubstr)
			}
		})
	}
}

func TestParseXeroError_Message(t *testing.T) {
	body := `{"Message":"Invoice total exceeds limit"}`
	xerr := ParseXeroError(400, strings.NewReader(body))
	errStr := xerr.Error()
	if !strings.Contains(errStr, "Invoice total exceeds limit") {
		t.Errorf("Error() = %q, want Message in output", errStr)
	}
}

func TestParseXeroError_ValidationErrors(t *testing.T) {
	body := `{
		"Message":"Validation failed",
		"Elements":[{
			"ValidationErrors":[
				{"Message":"Contact name is required"},
				{"Message":"Email is invalid"}
			]
		}]
	}`
	xerr := ParseXeroError(400, strings.NewReader(body))
	errStr := xerr.Error()
	if !strings.Contains(errStr, "Contact name is required") {
		t.Errorf("Error() missing validation error: %q", errStr)
	}
	if !strings.Contains(errStr, "Email is invalid") {
		t.Errorf("Error() missing second validation error: %q", errStr)
	}
}

func TestParseXeroError_EmptyBody(t *testing.T) {
	xerr := ParseXeroError(500, strings.NewReader(""))
	if xerr.StatusCode != 500 {
		t.Errorf("StatusCode = %d, want 500", xerr.StatusCode)
	}
	if xerr.Message != "" {
		t.Errorf("Message = %q, want empty", xerr.Message)
	}
	errStr := xerr.Error()
	if !strings.Contains(errStr, "Internal Server Error") {
		t.Errorf("Error() = %q, want status text even with empty body", errStr)
	}
}

func TestParseXeroError_MalformedJSON(t *testing.T) {
	xerr := ParseXeroError(400, strings.NewReader("{not json"))
	if xerr.StatusCode != 400 {
		t.Errorf("StatusCode = %d, want 400", xerr.StatusCode)
	}
	// Should not panic, just return error with status text
	errStr := xerr.Error()
	if !strings.Contains(errStr, "Bad Request") {
		t.Errorf("Error() = %q, want status text", errStr)
	}
}

func TestParseXeroError_AlternativeFields(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantSubstr string
	}{
		{
			name:       "Detail field",
			body:       `{"Detail":"Organization not found"}`,
			wantSubstr: "Organization not found",
		},
		{
			name:       "Title field",
			body:       `{"Title":"Forbidden"}`,
			wantSubstr: "Forbidden",
		},
		{
			name:       "lowercase message field",
			body:       `{"message":"something went wrong"}`,
			wantSubstr: "something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xerr := ParseXeroError(400, strings.NewReader(tt.body))
			errStr := xerr.Error()
			if !strings.Contains(errStr, tt.wantSubstr) {
				t.Errorf("Error() = %q, want substring %q", errStr, tt.wantSubstr)
			}
		})
	}
}

func TestExitCode_ValidationException(t *testing.T) {
	xerr := &XeroError{StatusCode: 200, Type: "ValidationException"}
	if got := xerr.ExitCode(); got != 3 {
		t.Errorf("ExitCode() = %d, want 3 for ValidationException", got)
	}
}

func TestXeroError_ErrorFormat(t *testing.T) {
	xerr := &XeroError{StatusCode: 404}
	got := xerr.Error()
	want := "Error: Not Found (HTTP 404)"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestXeroError_UnknownStatusCode(t *testing.T) {
	xerr := &XeroError{StatusCode: 999}
	got := xerr.Error()
	if !strings.Contains(got, "Unknown") {
		t.Errorf("Error() = %q, want 'Unknown' for unrecognized status", got)
	}
	if xerr.ExitCode() != 1 {
		t.Errorf("ExitCode() = %d, want 1 for unrecognized status", xerr.ExitCode())
	}
}
