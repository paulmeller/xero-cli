package output

import (
	"io"

	"github.com/tidwall/gjson"
)

// Column defines how to extract and display a field from JSON data.
type Column struct {
	Header string // Display header, e.g. "INVOICE #"
	Path   string // gjson path, e.g. "InvoiceNumber", "Contact.Name"
	Format string // Optional format hint: "currency", "date", "status"
	Width  int    // Minimum width (0 = auto)
}

// Formatter formats JSON data for output.
type Formatter interface {
	FormatList(w io.Writer, items gjson.Result, columns []Column) error
	FormatOne(w io.Writer, item gjson.Result, columns []Column) error
}
