package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/paulmeller/xero-cli/internal/cmdutil"
	"github.com/paulmeller/xero-cli/internal/output"
)

const sampleReportJSON = `{
	"Reports": [{
		"ReportName": "Profit and Loss",
		"Rows": [
			{"RowType": "Header", "Cells": [{"Value": "Account"}, {"Value": "Total"}]},
			{"RowType": "Section", "Title": "Income", "Rows": [
				{"RowType": "Row", "Cells": [{"Value": "Sales"}, {"Value": "100.5"}]}
			]},
			{"RowType": "SummaryRow", "Cells": [{"Value": "Total Income"}, {"Value": "100.5"}]}
		]
	}]
}`

func testFactory(buf *bytes.Buffer) *cmdutil.Factory {
	ios := &cmdutil.IOStreams{Out: buf, ErrOut: buf}
	return &cmdutil.Factory{
		IO: ios,
		Formatter: func(format string) output.Formatter {
			switch format {
			case "csv":
				return &output.CSVFormatter{}
			case "tsv":
				return &output.TSVFormatter{}
			default:
				return output.NewTableFormatter(ios.Out, false)
			}
		},
	}
}

// Regression for "reports -o csv emits a padded table instead of real CSV": renderReport used
// to hardcode f.Formatter("table") regardless of what the caller actually requested, and always
// interleaved prose (title, section headers) into the output stream, which would corrupt a real
// CSV/TSV parse even if the formatter itself were fixed.
func TestRenderReport_CSVFormat_IsRealCSVWithNoProse(t *testing.T) {
	var buf bytes.Buffer
	f := testFactory(&buf)

	if err := renderReport(f, []byte(sampleReportJSON), "csv"); err != nil {
		t.Fatalf("renderReport returned error: %v", err)
	}

	out := buf.String()
	if strings.Contains(out, "Profit and Loss") {
		t.Errorf("csv output must not contain the prose report title, got: %q", out)
	}
	if strings.Contains(out, "====") {
		t.Errorf("csv output must not contain the prose title underline, got: %q", out)
	}
	if strings.Contains(out, "Income\n") {
		t.Errorf("csv output must not contain the prose section header, got: %q", out)
	}
	if !strings.Contains(out, "ACCOUNT,TOTAL") {
		t.Errorf("csv output missing expected header row, got: %q", out)
	}
	if !strings.Contains(out, "Sales,100.5") {
		t.Errorf("csv output missing expected data row, got: %q", out)
	}
}

func TestRenderReport_TableFormat_StillShowsProse(t *testing.T) {
	var buf bytes.Buffer
	f := testFactory(&buf)

	if err := renderReport(f, []byte(sampleReportJSON), "table"); err != nil {
		t.Fatalf("renderReport returned error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Profit and Loss") {
		t.Errorf("table output should still show the report title, got: %q", out)
	}
	if !strings.Contains(out, "Income") {
		t.Errorf("table output should still show section headers, got: %q", out)
	}
}
