package output

import (
	"bytes"
	"testing"

	"github.com/tidwall/gjson"
)

// Regression for the "CSV/TSV ignore Column.Format" bug: previously csv.go and tsv.go only ran
// convertXeroDate and never looked at col.Format, so a "currency" column exported the raw
// unrounded JSON number (e.g. "123.4") instead of matching what the table formatter shows for
// the same data (formatCurrency's "%.2f", e.g. "123.40").
func TestCSVFormatter_FormatList_AppliesCurrencyFormat(t *testing.T) {
	columns := []Column{
		{Header: "AMOUNT", Path: "Amount", Format: "currency"},
	}
	items := gjson.Parse(`[{"Amount": 123.4}]`)

	var buf bytes.Buffer
	f := &CSVFormatter{}
	if err := f.FormatList(&buf, items, columns); err != nil {
		t.Fatalf("FormatList returned error: %v", err)
	}

	want := "AMOUNT\n123.40\n"
	if buf.String() != want {
		t.Errorf("FormatList output = %q, want %q", buf.String(), want)
	}
}

func TestCSVFormatter_FormatOne_AppliesCurrencyFormat(t *testing.T) {
	columns := []Column{
		{Header: "AMOUNT", Path: "Amount", Format: "currency"},
	}
	item := gjson.Parse(`{"Amount": 123.4}`)

	var buf bytes.Buffer
	f := &CSVFormatter{}
	if err := f.FormatOne(&buf, item, columns); err != nil {
		t.Fatalf("FormatOne returned error: %v", err)
	}

	want := "AMOUNT\n123.40\n"
	if buf.String() != want {
		t.Errorf("FormatOne output = %q, want %q", buf.String(), want)
	}
}

func TestTSVFormatter_FormatList_AppliesCurrencyFormat(t *testing.T) {
	columns := []Column{
		{Header: "AMOUNT", Path: "Amount", Format: "currency"},
	}
	items := gjson.Parse(`[{"Amount": 123.4}]`)

	var buf bytes.Buffer
	f := &TSVFormatter{}
	if err := f.FormatList(&buf, items, columns); err != nil {
		t.Fatalf("FormatList returned error: %v", err)
	}

	want := "AMOUNT\n123.40\n"
	if buf.String() != want {
		t.Errorf("FormatList output = %q, want %q", buf.String(), want)
	}
}
