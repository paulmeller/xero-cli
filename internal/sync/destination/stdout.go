package destination

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type StdoutDestination struct {
	w io.Writer
}

func NewStdoutDestination() *StdoutDestination {
	return &StdoutDestination{w: os.Stdout}
}

func (d *StdoutDestination) Init(ctx context.Context) error {
	return nil
}

func (d *StdoutDestination) Write(ctx context.Context, stream string, primaryKey string, records []json.RawMessage) (int, error) {
	written := 0
	for _, rec := range records {
		if _, err := fmt.Fprintf(d.w, "%s\n", rec); err != nil {
			return written, err
		}
		written++
	}
	return written, nil
}

func (d *StdoutDestination) Close() error {
	return nil
}
