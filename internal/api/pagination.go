package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/tidwall/gjson"
)

// defaultPageSize is Xero's server-side default when no pageSize param is sent.
const defaultPageSize = 100

// maxPages is a hard safety cap on PaginateAll's loop. It exists purely as a backstop against
// endpoints that don't paginate the way callers expect (see the stall check below) - any real
// Xero collection is expected to terminate long before this.
const maxPages = 1000

// PaginateAll fetches all pages and returns a merged gjson array.
// key is the JSON key containing the array (e.g. "Invoices", "Contacts").
func PaginateAll(ctx context.Context, client *Client, path string, params url.Values, key string, pageSize int) (gjson.Result, error) {
	if params == nil {
		params = url.Values{}
	}

	// pageSize<=0 means "don't override Xero's server-side default" for the request, but the
	// termination check below still needs a concrete number to compare against - otherwise
	// `len(arr) < pageSize` becomes `len(arr) < 0`, which is never true, and the loop never
	// terminates (e.g. `--page-size 0`).
	effectivePageSize := pageSize
	if effectivePageSize <= 0 {
		effectivePageSize = defaultPageSize
	}

	var allItems []json.RawMessage
	var prevPageRaw string
	page := 1

	for {
		if page > maxPages {
			return gjson.Result{}, fmt.Errorf("PaginateAll: exceeded %d pages fetching %s - this endpoint may not support page/pageSize pagination", maxPages, path)
		}

		params.Set("page", fmt.Sprintf("%d", page))
		if pageSize > 0 {
			params.Set("pageSize", fmt.Sprintf("%d", pageSize))
		}

		data, err := client.Get(ctx, path, params)
		if err != nil {
			return gjson.Result{}, err
		}

		parsed := gjson.ParseBytes(data)
		items := parsed.Get(key)
		if !items.Exists() || !items.IsArray() {
			break
		}

		arr := items.Array()
		if len(arr) == 0 {
			break
		}

		// Some Xero collections (e.g. BankTransfers) silently ignore page/pageSize and return
		// the full result set on every call, so len(arr) < effectivePageSize never fires. If
		// consecutive "pages" come back byte-identical, the endpoint isn't actually paginating -
		// stop instead of re-fetching the same data forever.
		if items.Raw == prevPageRaw {
			break
		}
		prevPageRaw = items.Raw

		for _, item := range arr {
			allItems = append(allItems, json.RawMessage(item.Raw))
		}

		if len(arr) < effectivePageSize {
			break
		}

		page++
	}

	result, err := json.Marshal(allItems)
	if err != nil {
		return gjson.Result{}, err
	}

	return gjson.ParseBytes(result), nil
}
