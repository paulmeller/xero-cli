package cmdutil

import (
	"encoding/json"
	"fmt"
)

// Exit codes
const (
	ExitOK         = 0
	ExitError      = 1
	ExitAuth       = 2
	ExitValidation = 3
	ExitNotFound   = 4
	ExitRateLimit  = 5
	ExitNetwork    = 6
)

// FormatError returns a human-readable or JSON error string based on output format.
func FormatError(err error, jsonOutput bool) string {
	if jsonOutput {
		data, _ := json.Marshal(map[string]string{"error": err.Error()})
		return string(data)
	}
	return err.Error()
}

// SilentError is an error that has already been printed to stderr.
// The main function should exit with the given code without printing again.
type SilentError struct {
	Code int
}

func (e *SilentError) Error() string {
	return fmt.Sprintf("exit code %d", e.Code)
}
