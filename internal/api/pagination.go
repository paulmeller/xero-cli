package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/tidwall/gjson"
)

// PaginateAll fetches all pages and returns a merged gjson array.
// key is the JSON key containing the array (e.g. "Invoices", "Contacts").
func PaginateAll(ctx context.Context, client *Client, path string, params url.Values, key string, pageSize int) (gjson.Result, error) {
	if params == nil {
		params = url.Values{}
	}

	var allItems []json.RawMessage
	page := 1

	for {
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

		for _, item := range arr {
			allItems = append(allItems, json.RawMessage(item.Raw))
		}

		if len(arr) < pageSize {
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
