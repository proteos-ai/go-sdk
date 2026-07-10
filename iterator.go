package sdk

import (
	"context"
	"go.proteos.ai/model/common"
)

// DefaultPageSize is applied to a list request when the caller leaves
// PageSize at zero.
const DefaultPageSize = 100

// ListResult is the standard {meta, data} envelope returned by every list
// endpoint. Pages are 0-indexed.
type ListResult[T any] struct {
	Meta common.ResponseMeta `json:"meta"`
	Data []T                `json:"data"`
}

// ListPageFunc fetches a single page. Implementations should pass page
// straight through to the wire request; opts provides everything else.
type ListPageFunc[T any, O any] func(ctx context.Context, page int, opts O) (ListResult[T], error)

// PageIterator walks a paginated resource one item at a time. Pages are
// 0-indexed. Construct via NewPageIterator; the zero value is unusable.
type PageIterator[T any, O any] struct {
	fetch      ListPageFunc[T, O]
	opts       O
	page       int
	buf        []T
	bufIdx     int
	pagesTotal int
	knownTotal bool
	done       bool
}

// NewPageIterator returns an iterator that calls fetch to load each page.
func NewPageIterator[T any, O any](fetch ListPageFunc[T, O], opts O) *PageIterator[T, O] {
	return &PageIterator[T, O]{fetch: fetch, opts: opts}
}

// Next returns the next item. (zero, false, nil) signals normal end of
// iteration; (zero, false, err) signals an error from the underlying fetch.
//
//	for {
//	    item, ok, err := it.Next(ctx)
//	    if err != nil { return err }
//	    if !ok { break }
//	    // use item
//	}
func (it *PageIterator[T, O]) Next(ctx context.Context) (T, bool, error) {
	var zero T
	if it.done {
		return zero, false, nil
	}
	if it.bufIdx < len(it.buf) {
		item := it.buf[it.bufIdx]
		it.bufIdx++
		return item, true, nil
	}
	if it.knownTotal && it.page >= it.pagesTotal {
		it.done = true
		return zero, false, nil
	}
	res, err := it.fetch(ctx, it.page, it.opts)
	if err != nil {
		return zero, false, err
	}
	it.pagesTotal = res.Meta.PagesTotal
	it.knownTotal = true
	if len(res.Data) == 0 {
		it.done = true
		return zero, false, nil
	}
	it.buf = res.Data
	it.bufIdx = 1
	it.page++
	return res.Data[0], true, nil
}

// All collects every remaining item into a slice. Convenient for CLI
// commands that just print the full list.
func (it *PageIterator[T, O]) All(ctx context.Context) ([]T, error) {
	var out []T
	for {
		item, ok, err := it.Next(ctx)
		if err != nil {
			return nil, err
		}
		if !ok {
			return out, nil
		}
		out = append(out, item)
	}
}
