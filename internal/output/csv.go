package output

import (
	"encoding/csv"
	"io"

	"github.com/tidwall/gjson"
)

type CSVFormatter struct{}

func (f *CSVFormatter) FormatList(w io.Writer, items gjson.Result, columns []Column) error {
	cw := csv.NewWriter(w)

	// Header
	headers := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = col.Header
	}
	if err := cw.Write(headers); err != nil {
		return err
	}

	// Rows
	var writeErr error
	items.ForEach(func(_, item gjson.Result) bool {
		row := make([]string, len(columns))
		for i, col := range columns {
			val := item.Get(col.Path)
			row[i] = convertXeroDate(val.String())
		}
		if err := cw.Write(row); err != nil {
			writeErr = err
			return false
		}
		return true
	})
	if writeErr != nil {
		return writeErr
	}

	cw.Flush()
	return cw.Error()
}

func (f *CSVFormatter) FormatOne(w io.Writer, item gjson.Result, columns []Column) error {
	cw := csv.NewWriter(w)

	headers := make([]string, len(columns))
	values := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = col.Header
		val := item.Get(col.Path)
		values[i] = convertXeroDate(val.String())
	}

	if err := cw.Write(headers); err != nil {
		return err
	}
	if err := cw.Write(values); err != nil {
		return err
	}

	cw.Flush()
	return cw.Error()
}
