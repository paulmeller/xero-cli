package output

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/tidwall/gjson"
)

type JSONFormatter struct{}

func (f *JSONFormatter) FormatList(w io.Writer, items gjson.Result, columns []Column) error {
	return writeIndentedJSON(w, items.Raw)
}

func (f *JSONFormatter) FormatOne(w io.Writer, item gjson.Result, columns []Column) error {
	return writeIndentedJSON(w, item.Raw)
}

func writeIndentedJSON(w io.Writer, raw string) error {
	if raw == "" {
		_, err := fmt.Fprintln(w, "null")
		return err
	}
	var buf json.RawMessage = []byte(raw)
	indented, err := json.MarshalIndent(buf, "", "  ")
	if err != nil {
		// Fallback to raw output
		_, err = fmt.Fprintln(w, raw)
		return err
	}
	_, err = fmt.Fprintln(w, string(indented))
	return err
}
