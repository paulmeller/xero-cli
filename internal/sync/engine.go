package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"sort"
	"time"

	"github.com/tidwall/gjson"

	"github.com/paulmeller/xero-cli/internal/api"
	"github.com/paulmeller/xero-cli/internal/sync/destination"
)

type Engine struct {
	client   *api.Client
	config   *SyncConfig
	state    *SyncState
	dest     destination.Destination
	errOut   io.Writer
	dryRun   bool
	apiCalls int
}

func NewEngine(client *api.Client, config *SyncConfig, state *SyncState, dest destination.Destination, errOut io.Writer, dryRun bool) *Engine {
	return &Engine{
		client: client,
		config: config,
		state:  state,
		dest:   dest,
		errOut: errOut,
		dryRun: dryRun,
	}
}

func (e *Engine) Run(ctx context.Context, streamFilter []string) error {
	if err := e.dest.Init(ctx); err != nil {
		return fmt.Errorf("destination init failed: %w", err)
	}
	defer e.dest.Close()

	streams := e.orderedStreams(streamFilter)
	if len(streams) == 0 {
		fmt.Fprintf(e.errOut, "No streams to sync.\n")
		return nil
	}

	fmt.Fprintf(e.errOut, "Syncing %d streams...\n", len(streams))

	for _, sc := range streams {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if e.config.Sync.DailyBudget > 0 && e.apiCalls >= e.config.Sync.DailyBudget {
			fmt.Fprintf(e.errOut, "Daily API budget exhausted (%d calls). Stopping.\n", e.apiCalls)
			break
		}

		if err := e.syncStream(ctx, sc); err != nil {
			fmt.Fprintf(e.errOut, "Error syncing %s: %v\n", sc.Name, err)
			continue
		}
	}

	return nil
}

func (e *Engine) syncStream(ctx context.Context, sc StreamConfig) error {
	meta, ok := StreamRegistry[sc.Name]
	if !ok {
		return fmt.Errorf("unknown stream: %s", sc.Name)
	}

	start := time.Now()
	fmt.Fprintf(e.errOut, "  %s: syncing...", sc.Name)

	params := url.Values{}

	// For incremental sync, use the cursor/bookmark
	isIncremental := sc.SyncMode == "incremental" && sc.CursorField != ""
	if isIncremental {
		if ss, ok := e.state.Streams[sc.Name]; ok && ss.CursorValue != "" {
			params.Set("If-Modified-Since", ss.CursorValue)
		}
	}

	// Apply Xero API where filter if configured on the stream
	if sc.Where != "" {
		params.Set("where", sc.Where)
	}

	// For full_refresh, truncate the destination before writing
	if sc.SyncMode == "full_refresh" {
		type truncater interface {
			TruncateStream(stream string) error
		}
		if t, ok := e.dest.(truncater); ok {
			destName := sc.Name
			if sc.DestinationTable != "" {
				destName = sc.DestinationTable
			}
			if err := t.TruncateStream(destName); err != nil {
				return fmt.Errorf("truncate failed: %w", err)
			}
		}
	}

	if e.dryRun {
		fmt.Fprintf(e.errOut, " [dry-run] would fetch %s\n", meta.APIPath)
		return nil
	}

	// Paginate through all results
	var allRecords []json.RawMessage
	page := 1
	pageSize := 100

	for {
		params.Set("page", fmt.Sprintf("%d", page))
		params.Set("pageSize", fmt.Sprintf("%d", pageSize))

		data, err := e.client.Get(ctx, meta.APIPath, params)
		e.apiCalls++
		if err != nil {
			return err
		}

		parsed := gjson.ParseBytes(data)
		items := parsed.Get(meta.JSONKey)
		if !items.Exists() || !items.IsArray() {
			break
		}

		arr := items.Array()
		if len(arr) == 0 {
			break
		}

		for _, item := range arr {
			allRecords = append(allRecords, json.RawMessage(item.Raw))
		}

		if len(arr) < pageSize {
			break
		}
		page++
	}

	if len(allRecords) == 0 {
		fmt.Fprintf(e.errOut, " no new records (%.1fs)\n", time.Since(start).Seconds())
		return nil
	}

	// TODO: if sc.SelectedFields is set, filter each record to only include
	// those fields (would need gjson extraction + reconstruction per record).

	// Write to destination, using DestinationTable as the target name if set
	destName := sc.Name
	if sc.DestinationTable != "" {
		destName = sc.DestinationTable
	}
	written, err := e.dest.Write(ctx, destName, sc.PrimaryKey, allRecords)
	if err != nil {
		return fmt.Errorf("write failed: %w", err)
	}

	// Update cursor from the highest cursor value seen
	if isIncremental && sc.CursorField != "" {
		highestCursor := ""
		for _, rec := range allRecords {
			val := gjson.GetBytes(rec, sc.CursorField).String()
			if val > highestCursor {
				highestCursor = val
			}
		}
		if highestCursor != "" {
			e.state.Streams[sc.Name] = StreamState{
				CursorValue:   highestCursor,
				LastSync:      time.Now().UTC(),
				RecordsSynced: written,
			}
		}
	} else {
		e.state.Streams[sc.Name] = StreamState{
			LastSync:      time.Now().UTC(),
			RecordsSynced: written,
		}
	}

	fmt.Fprintf(e.errOut, " %d records (%.1fs)\n", written, time.Since(start).Seconds())
	return nil
}

func (e *Engine) orderedStreams(filter []string) []StreamConfig {
	filterSet := make(map[string]bool)
	for _, f := range filter {
		filterSet[f] = true
	}

	enabledMap := make(map[string]StreamConfig)
	for _, sc := range e.config.Streams {
		if !sc.Enabled {
			continue
		}
		if len(filterSet) > 0 && !filterSet[sc.Name] {
			continue
		}
		enabledMap[sc.Name] = sc
	}

	var result []StreamConfig
	for _, name := range StreamPriority {
		if sc, ok := enabledMap[name]; ok {
			result = append(result, sc)
			delete(enabledMap, name)
		}
	}
	// Add any remaining streams not in priority list (sorted for determinism)
	remaining := make([]string, 0, len(enabledMap))
	for name := range enabledMap {
		remaining = append(remaining, name)
	}
	sort.Strings(remaining)
	for _, name := range remaining {
		result = append(result, enabledMap[name])
	}

	return result
}
