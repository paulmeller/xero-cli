package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/tidwall/gjson"
)

// TSVFormatter outputs tab-separated values — no quoting, no alignment.
// Designed for piping to awk, cut, sort, etc.
type TSVFormatter struct{}

func (f *TSVFormatter) FormatList(w io.Writer, items gjson.Result, columns []Column) error {
	// Header
	headers := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = col.Header
	}
	fmt.Fprintln(w, strings.Join(headers, "\t"))

	// Rows
	items.ForEach(func(_, item gjson.Result) bool {
		row := make([]string, len(columns))
		for i, col := range columns {
			val := item.Get(col.Path)
			row[i] = convertXeroDate(val.String())
		}
		fmt.Fprintln(w, strings.Join(row, "\t"))
		return true
	})

	return nil
}

func (f *TSVFormatter) FormatOne(w io.Writer, item gjson.Result, columns []Column) error {
	headers := make([]string, len(columns))
	values := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = col.Header
		val := item.Get(col.Path)
		values[i] = convertXeroDate(val.String())
	}

	fmt.Fprintln(w, strings.Join(headers, "\t"))
	fmt.Fprintln(w, strings.Join(values, "\t"))
	return nil
}
