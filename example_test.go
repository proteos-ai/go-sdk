package sdk_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	sdk "go.proteos.ai/sdk"
	"go.proteos.ai/sdk/meta"
)

// Example_listEntities is a runnable end-to-end smoke check: it boots an
// httptest server, builds a *sdk.Client against it, and exercises every
// layer (options -> HTTP -> query encoding -> JSON decode -> model
// type) via meta.New(c).Entities.ListPage.
func Example_listEntities() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"meta": map[string]any{"page": 0, "page_size": 100, "items_total": 1, "pages_total": 1},
			"data": []map[string]any{{
				"slug":        "customer",
				"name":        "Customer",
				"description": "people who buy stuff",
				"is_remote":   false,
				"module_slug": "core",
				"attributes":  []any{},
				"created_by":  map[string]any{"type": "person", "id": "platform"},
				"updated_by":  map[string]any{"type": "person", "id": "platform"},
			}},
		})
	}))
	defer srv.Close()

	c, err := sdk.NewClient(sdk.WithBaseURL(srv.URL), sdk.WithToken("t"))
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	page, err := meta.New(c).Entities.ListPage(context.Background(), nil)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	fmt.Println(page.Data[0].Slug)
	// Output: customer
}
