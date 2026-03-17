package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadStream_BasicDedup(t *testing.T) {
	tmpDir := t.TempDir()
	jsonlPath := filepath.Join(tmpDir, "invoices.jsonl")

	lines := strings.Join([]string{
		`{"InvoiceID":"aaa","Total":100}`,
		`{"InvoiceID":"bbb","Total":200}`,
		`{"InvoiceID":"aaa","Total":150}`, // duplicate, should overwrite first
	}, "\n") + "\n"

	if err := os.WriteFile(jsonlPath, []byte(lines), 0644); err != nil {
		t.Fatal(err)
	}

	raw, err := ReadStream(jsonlPath, "Invoices", "InvoiceID")
	if err != nil {
		t.Fatalf("ReadStream error: %v", err)
	}

	var envelope map[string][]json.RawMessage
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	invoices := envelope["Invoices"]
	if len(invoices) != 2 {
		t.Fatalf("got %d invoices, want 2", len(invoices))
	}

	// First record should be "aaa" with the updated Total (last occurrence wins)
	var first map[string]interface{}
	if err := json.Unmarshal(invoices[0], &first); err != nil {
		t.Fatal(err)
	}
	if first["InvoiceID"] != "aaa" {
		t.Errorf("first record ID = %v, want 'aaa'", first["InvoiceID"])
	}
	if first["Total"].(float64) != 150 {
		t.Errorf("first record Total = %v, want 150 (last occurrence wins)", first["Total"])
	}

	var second map[string]interface{}
	if err := json.Unmarshal(invoices[1], &second); err != nil {
		t.Fatal(err)
	}
	if second["InvoiceID"] != "bbb" {
		t.Errorf("second record ID = %v, want 'bbb'", second["InvoiceID"])
	}
}

func TestReadStream_MalformedLines(t *testing.T) {
	tmpDir := t.TempDir()
	jsonlPath := filepath.Join(tmpDir, "contacts.jsonl")

	lines := strings.Join([]string{
		`{"ContactID":"c1","Name":"Alice"}`,
		`not valid json at all`,
		`{"ContactID":"c2","Name":"Bob"}`,
		``, // empty line
		`{broken`,
	}, "\n") + "\n"

	if err := os.WriteFile(jsonlPath, []byte(lines), 0644); err != nil {
		t.Fatal(err)
	}

	raw, err := ReadStream(jsonlPath, "Contacts", "ContactID")
	if err != nil {
		t.Fatalf("ReadStream error: %v", err)
	}

	var envelope map[string][]json.RawMessage
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	contacts := envelope["Contacts"]
	if len(contacts) != 2 {
		t.Fatalf("got %d contacts, want 2 (malformed lines skipped)", len(contacts))
	}
}

func TestReadStream_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	jsonlPath := filepath.Join(tmpDir, "empty.jsonl")

	if err := os.WriteFile(jsonlPath, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	raw, err := ReadStream(jsonlPath, "Invoices", "InvoiceID")
	if err != nil {
		t.Fatalf("ReadStream error: %v", err)
	}

	var envelope map[string][]json.RawMessage
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	invoices := envelope["Invoices"]
	if len(invoices) != 0 {
		t.Fatalf("got %d invoices, want 0 for empty file", len(invoices))
	}
}

func TestReadStream_Envelope(t *testing.T) {
	tmpDir := t.TempDir()
	jsonlPath := filepath.Join(tmpDir, "accounts.jsonl")

	lines := `{"AccountID":"a1","Name":"Sales"}` + "\n"
	if err := os.WriteFile(jsonlPath, []byte(lines), 0644); err != nil {
		t.Fatal(err)
	}

	raw, err := ReadStream(jsonlPath, "Accounts", "AccountID")
	if err != nil {
		t.Fatal(err)
	}

	// Check that the envelope key is "Accounts"
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatal(err)
	}
	if _, ok := envelope["Accounts"]; !ok {
		t.Error("missing 'Accounts' key in envelope")
	}
}

func TestReadByID_Found(t *testing.T) {
	tmpDir := t.TempDir()
	jsonlPath := filepath.Join(tmpDir, "invoices.jsonl")

	lines := strings.Join([]string{
		`{"InvoiceID":"aaa","Total":100}`,
		`{"InvoiceID":"bbb","Total":200}`,
		`{"InvoiceID":"ccc","Total":300}`,
	}, "\n") + "\n"

	if err := os.WriteFile(jsonlPath, []byte(lines), 0644); err != nil {
		t.Fatal(err)
	}

	raw, err := ReadByID(jsonlPath, "Invoices", "InvoiceID", "bbb")
	if err != nil {
		t.Fatalf("ReadByID error: %v", err)
	}

	var envelope map[string][]json.RawMessage
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatal(err)
	}

	invoices := envelope["Invoices"]
	if len(invoices) != 1 {
		t.Fatalf("got %d invoices, want 1", len(invoices))
	}

	var record map[string]interface{}
	if err := json.Unmarshal(invoices[0], &record); err != nil {
		t.Fatal(err)
	}
	if record["InvoiceID"] != "bbb" {
		t.Errorf("got ID %v, want 'bbb'", record["InvoiceID"])
	}
	if record["Total"].(float64) != 200 {
		t.Errorf("got Total %v, want 200", record["Total"])
	}
}

func TestReadByID_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	jsonlPath := filepath.Join(tmpDir, "invoices.jsonl")

	lines := `{"InvoiceID":"aaa","Total":100}` + "\n"
	if err := os.WriteFile(jsonlPath, []byte(lines), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := ReadByID(jsonlPath, "Invoices", "InvoiceID", "zzz")
	if err == nil {
		t.Fatal("ReadByID should return error for missing ID")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want 'not found' in message", err.Error())
	}
}

func TestReadByID_LastOccurrenceWins(t *testing.T) {
	tmpDir := t.TempDir()
	jsonlPath := filepath.Join(tmpDir, "invoices.jsonl")

	lines := strings.Join([]string{
		`{"InvoiceID":"aaa","Total":100}`,
		`{"InvoiceID":"aaa","Total":999}`, // updated record
	}, "\n") + "\n"

	if err := os.WriteFile(jsonlPath, []byte(lines), 0644); err != nil {
		t.Fatal(err)
	}

	raw, err := ReadByID(jsonlPath, "Invoices", "InvoiceID", "aaa")
	if err != nil {
		t.Fatal(err)
	}

	var envelope map[string][]json.RawMessage
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatal(err)
	}

	var record map[string]interface{}
	if err := json.Unmarshal(envelope["Invoices"][0], &record); err != nil {
		t.Fatal(err)
	}
	if record["Total"].(float64) != 999 {
		t.Errorf("Total = %v, want 999 (last occurrence wins)", record["Total"])
	}
}

func TestReadByID_SkipsMalformed(t *testing.T) {
	tmpDir := t.TempDir()
	jsonlPath := filepath.Join(tmpDir, "invoices.jsonl")

	lines := strings.Join([]string{
		`not valid json`,
		`{"InvoiceID":"aaa","Total":100}`,
	}, "\n") + "\n"

	if err := os.WriteFile(jsonlPath, []byte(lines), 0644); err != nil {
		t.Fatal(err)
	}

	raw, err := ReadByID(jsonlPath, "Invoices", "InvoiceID", "aaa")
	if err != nil {
		t.Fatalf("ReadByID should succeed despite malformed line: %v", err)
	}

	var envelope map[string][]json.RawMessage
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatal(err)
	}
	if len(envelope["Invoices"]) != 1 {
		t.Errorf("got %d invoices, want 1", len(envelope["Invoices"]))
	}
}
