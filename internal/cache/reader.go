package cache

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/tidwall/gjson"
)

// ReadStream reads a JSONL file, deduplicates by primaryKey, and returns
// the result wrapped in an API envelope: {"<jsonKey>": [...]}.
// Skips malformed lines (tolerance for partial writes during concurrent sync).
func ReadStream(jsonlPath, jsonKey, primaryKey string) (json.RawMessage, error) {
	f, err := os.Open(jsonlPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Dedup map: primaryKey value -> last-seen raw JSON line
	dedup := make(map[string]json.RawMessage)
	var order []string // preserve insertion order for stable output

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024) // 10MB max line

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		if !json.Valid(line) {
			continue
		}

		key := gjson.GetBytes(line, primaryKey).String()
		if key == "" {
			continue
		}

		if _, exists := dedup[key]; !exists {
			order = append(order, key)
		}
		// Copy the line bytes since scanner reuses the buffer
		lineCopy := make(json.RawMessage, len(line))
		copy(lineCopy, line)
		dedup[key] = lineCopy
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading JSONL: %w", err)
	}

	// Build the array in insertion order (last-seen wins via dedup)
	items := make([]json.RawMessage, 0, len(order))
	for _, k := range order {
		items = append(items, dedup[k])
	}

	arr, err := json.Marshal(items)
	if err != nil {
		return nil, err
	}

	// Wrap in envelope: {"Invoices": [...]}
	return json.RawMessage(`{"` + jsonKey + `":` + string(arr) + `}`), nil
}

// ReadByID reads a JSONL file and returns the single record matching id.
// Wraps in envelope: {"<jsonKey>": [{...}]}.
// Only tracks the matching record (last occurrence wins), no full dedup needed.
func ReadByID(jsonlPath, jsonKey, primaryKey, id string) (json.RawMessage, error) {
	f, err := os.Open(jsonlPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var match json.RawMessage

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		if !json.Valid(line) {
			continue
		}

		key := gjson.GetBytes(line, primaryKey).String()
		if key == id {
			match = make(json.RawMessage, len(line))
			copy(match, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading JSONL: %w", err)
	}

	if match == nil {
		return nil, fmt.Errorf("%s %s not found in cache", primaryKey, id)
	}

	return json.RawMessage(`{"` + jsonKey + `":[` + string(match) + `]}`), nil
}
