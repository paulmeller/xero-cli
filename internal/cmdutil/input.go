package cmdutil

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// ReadInput reads input from --file flag. Returns nil if no --file specified.
// Supports --file <path> and --file - (stdin).
func ReadInput(cmd *cobra.Command) (json.RawMessage, error) {
	filePath, _ := cmd.Flags().GetString("file")
	if filePath == "" {
		return nil, nil
	}

	var r io.Reader
	if filePath == "-" {
		r = os.Stdin
	} else {
		f, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("cannot open file %s: %w", filePath, err)
		}
		defer f.Close()
		r = f
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("cannot read input: %w", err)
	}

	data = []byte(strings.TrimSpace(string(data)))
	if len(data) == 0 {
		return nil, fmt.Errorf("empty input")
	}

	// Validate it's valid JSON
	if !json.Valid(data) {
		return nil, fmt.Errorf("invalid JSON input")
	}

	// Reject bare scalars (null, true, 42, "foo"). Every create/update command wraps this value
	// into a {JSONKey: [...]} envelope or sends it directly as the request body, so a scalar
	// would silently become something like {"Accounts":[null]} and produce a confusing
	// server-side error instead of a clear one here.
	if data[0] != '{' && data[0] != '[' {
		preview := string(data)
		if len(preview) > 40 {
			preview = preview[:40] + "..."
		}
		return nil, fmt.Errorf("input must be a JSON object or array, got: %s", preview)
	}

	return json.RawMessage(data), nil
}

// IsBatchInput checks if the input is a JSON array (for batch operations).
func IsBatchInput(data json.RawMessage) bool {
	if len(data) == 0 {
		return false
	}
	return data[0] == '['
}
