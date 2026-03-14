package destination

import (
	"context"
	"encoding/json"
)

// Destination writes synced records to a target.
type Destination interface {
	Init(ctx context.Context) error
	Write(ctx context.Context, stream string, primaryKey string, records []json.RawMessage) (int, error)
	Close() error
}
