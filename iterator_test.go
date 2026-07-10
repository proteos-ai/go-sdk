package sdk

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.proteos.ai/model/common"
)

type listOpts struct {
	PageSize int
}

func makeFetch(t *testing.T, pages [][]string, pagesTotal int, expectPageSize int) (ListPageFunc[string, listOpts], *[]int) {
	t.Helper()
	seenPages := []int{}
	return func(ctx context.Context, page int, opts listOpts) (ListResult[string], error) {
		seenPages = append(seenPages, page)
		require.Equal(t, expectPageSize, opts.PageSize, "iterator should pass PageSize through")
		require.NoError(t, ctx.Err())
		var data []string
		if page < len(pages) {
			data = pages[page]
		}
		return ListResult[string]{
			Meta: common.ResponseMeta{
				Pagination: common.Pagination{Page: page, PageSize: opts.PageSize},
				PagesTotal: pagesTotal,
				ItemsTotal: pagesTotal * opts.PageSize,
			},
			Data: data,
		}, nil
	}, &seenPages
}

func TestPageIterator_Next_WalksAllPages(t *testing.T) {
	pages := [][]string{{"a", "b"}, {"c", "d"}, {"e", "f"}}
	fetch, seen := makeFetch(t, pages, 3, 2)
	it := NewPageIterator(fetch, listOpts{PageSize: 2})
	var got []string
	for {
		item, ok, err := it.Next(context.Background())
		require.NoError(t, err)
		if !ok {
			break
		}
		got = append(got, item)
	}
	require.Equal(t, []string{"a", "b", "c", "d", "e", "f"}, got)
	require.Equal(t, []int{0, 1, 2}, *seen, "pages must be 0-indexed")
}

func TestPageIterator_All_Collects(t *testing.T) {
	pages := [][]string{{"x"}, {"y"}, {"z"}}
	fetch, _ := makeFetch(t, pages, 3, 1)
	it := NewPageIterator(fetch, listOpts{PageSize: 1})
	all, err := it.All(context.Background())
	require.NoError(t, err)
	require.Equal(t, []string{"x", "y", "z"}, all)
}

func TestPageIterator_StopsAtPagesTotal(t *testing.T) {
	pages := [][]string{{"a"}, {"b"}}
	fetch, seen := makeFetch(t, pages, 2, 1)
	it := NewPageIterator(fetch, listOpts{PageSize: 1})
	all, err := it.All(context.Background())
	require.NoError(t, err)
	require.Equal(t, []string{"a", "b"}, all)
	// Iterator should not request a 3rd page even though we'd have time to.
	require.Equal(t, []int{0, 1}, *seen)
}

func TestPageIterator_StopsOnEmptyData(t *testing.T) {
	called := 0
	fetch := func(ctx context.Context, page int, opts listOpts) (ListResult[string], error) {
		called++
		return ListResult[string]{
			Meta: common.ResponseMeta{Pagination: common.Pagination{Page: page, PageSize: opts.PageSize}, PagesTotal: 99},
			Data: nil, // server reports more pages but returns nothing
		}, nil
	}
	it := NewPageIterator(fetch, listOpts{PageSize: 10})
	all, err := it.All(context.Background())
	require.NoError(t, err)
	require.Empty(t, all)
	require.Equal(t, 1, called, "must stop after the first empty page")
}

func TestPageIterator_PropagatesError(t *testing.T) {
	wantErr := errors.New("boom")
	fetch := func(ctx context.Context, page int, opts listOpts) (ListResult[string], error) {
		return ListResult[string]{}, wantErr
	}
	it := NewPageIterator(fetch, listOpts{PageSize: 10})
	_, _, err := it.Next(context.Background())
	require.ErrorIs(t, err, wantErr)
}

func TestPageIterator_ApplyDefaultPageSize_AtServiceLayer(t *testing.T) {
	// PageIterator itself doesn't set defaults — service layer is responsible.
	// This test documents the contract: opts is passed through verbatim.
	fetch := func(ctx context.Context, page int, opts listOpts) (ListResult[string], error) {
		require.Equal(t, 0, opts.PageSize)
		return ListResult[string]{Meta: common.ResponseMeta{PagesTotal: 0}}, nil
	}
	it := NewPageIterator(fetch, listOpts{})
	_, _, err := it.Next(context.Background())
	require.NoError(t, err)
}

func TestPageIterator_RespectsContextCancellation(t *testing.T) {
	fetch := func(ctx context.Context, page int, opts listOpts) (ListResult[string], error) {
		return ListResult[string]{}, ctx.Err()
	}
	it := NewPageIterator(fetch, listOpts{PageSize: 10})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, err := it.Next(ctx)
	require.ErrorIs(t, err, context.Canceled)
}

func TestDefaultPageSize_IsHundred(t *testing.T) {
	require.Equal(t, 100, DefaultPageSize)
}
