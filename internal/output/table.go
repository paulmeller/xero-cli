package output

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/tidwall/gjson"
)

var xeroDateRe = regexp.MustCompile(`^/Date\((\d+)([+-]\d{4})?\)/$`)

type TableFormatter struct {
	w      io.Writer
	color  bool
}

func NewTableFormatter(w io.Writer, color bool) *TableFormatter {
	return &TableFormatter{w: w, color: color}
}

func (f *TableFormatter) FormatList(w io.Writer, items gjson.Result, columns []Column) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)

	// Header
	headers := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = col.Header
	}
	fmt.Fprintln(tw, strings.Join(headers, "\t"))

	// Rows
	items.ForEach(func(_, item gjson.Result) bool {
		vals := make([]string, len(columns))
		for i, col := range columns {
			val := item.Get(col.Path)
			vals[i] = f.formatValue(val.String(), col.Format)
		}
		fmt.Fprintln(tw, strings.Join(vals, "\t"))
		return true
	})

	return tw.Flush()
}

func (f *TableFormatter) FormatOne(w io.Writer, item gjson.Result, columns []Column) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)

	for _, col := range columns {
		val := item.Get(col.Path)
		formatted := f.formatValue(val.String(), col.Format)
		fmt.Fprintf(tw, "%s:\t%s\n", col.Header, formatted)
	}

	return tw.Flush()
}

func (f *TableFormatter) formatValue(val string, format string) string {
	// Convert Xero dates
	val = convertXeroDate(val)

	switch format {
	case "status":
		return f.colorStatus(val)
	case "currency":
		return formatCurrency(val)
	case "date":
		return val
	default:
		return val
	}
}

func (f *TableFormatter) colorStatus(status string) string {
	if !f.color {
		return status
	}

	switch strings.ToUpper(status) {
	case "PAID", "ACTIVE", "AUTHORISED":
		return "\033[32m" + status + "\033[0m" // green
	case "DRAFT":
		return "\033[33m" + status + "\033[0m" // yellow
	case "OVERDUE", "DELETED", "VOIDED":
		return "\033[31m" + status + "\033[0m" // red
	case "SUBMITTED", "SENT":
		return "\033[36m" + status + "\033[0m" // cyan
	default:
		return status
	}
}

func formatCurrency(val string) string {
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return val
	}
	return fmt.Sprintf("%.2f", f)
}

// convertXeroDate converts Xero's /Date(1639094400000+0000)/ format to ISO 8601.
func convertXeroDate(val string) string {
	matches := xeroDateRe.FindStringSubmatch(val)
	if matches == nil {
		return val
	}
	ms, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return val
	}
	t := time.UnixMilli(ms).UTC()
	return t.Format("2006-01-02")
}
