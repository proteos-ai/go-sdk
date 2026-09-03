package data

import (
	"context"
	"encoding/json"
	"maps"
	"net/http"

	datamodel "go.proteos.ai/model/data"
	sdk "go.proteos.ai/sdk"
)

const (
	recordsBasePath      = "/data/v1/records"
	batchRecordsBasePath = "/data/v1/batch/records"
)

// RecordServiceAPI is the contract a RecordService satisfies.
type RecordServiceAPI interface {
	List(entitySlug string, opts *ListRecordsOptions) *sdk.PageIterator[datamodel.Record, ListRecordsOptions]
	ListPage(ctx context.Context, entitySlug string, opts *ListRecordsOptions) (sdk.ListResult[datamodel.Record], error)
	Get(ctx context.Context, entitySlug, id string) (datamodel.Record, error)
	Create(ctx context.Context, entitySlug string, data datamodel.Record) (datamodel.Record, error)
	Update(ctx context.Context, entitySlug, id string, data datamodel.Record) (datamodel.Record, error)
	Delete(ctx context.Context, entitySlug, id string) error
	BatchUpsert(ctx context.Context, entitySlug string, txns []BatchUpsertTransaction) (BatchUpsertRecordsResponse, error)
}

// RecordService manages records (per-entity rows) via the data-service.
type RecordService struct{ c *sdk.Client }

var _ RecordServiceAPI = (*RecordService)(nil)

// List returns a PageIterator over records for the given entity.
func (s *RecordService) List(entitySlug string, opts *ListRecordsOptions) *sdk.PageIterator[datamodel.Record, ListRecordsOptions] {
	o := ListRecordsOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListRecordsOptions) (sdk.ListResult[datamodel.Record], error) {
		in.Page = page
		return s.ListPage(ctx, entitySlug, &in)
	}, o)
}

// ListPage fetches a single page of records for the given entity.
func (s *RecordService) ListPage(ctx context.Context, entitySlug string, opts *ListRecordsOptions) (sdk.ListResult[datamodel.Record], error) {
	var out sdk.ListResult[datamodel.Record]
	opts, err := withEncodedFilter(opts)
	if err != nil {
		return out, err
	}
	err = s.c.DoWithQuery(ctx, http.MethodGet, recordsBasePath+"/"+entitySlug, opts, nil, &out)
	return out, err
}

// withEncodedFilter folds a nested Filter tree into the flat Filters map as
// the `_filter` JSON param (the query encoder only emits primitives). The
// input options are never mutated — callers may reuse them across pages.
func withEncodedFilter(opts *ListRecordsOptions) (*ListRecordsOptions, error) {
	if opts == nil || opts.Filter == nil {
		return opts, nil
	}
	encoded, err := json.Marshal(opts.Filter)
	if err != nil {
		return nil, err
	}
	out := *opts
	out.Filter = nil
	out.Filters = make(map[string]any, len(opts.Filters)+1)
	maps.Copy(out.Filters, opts.Filters)
	out.Filters["_filter"] = string(encoded)
	return &out, nil
}

// Get returns a single record.
func (s *RecordService) Get(ctx context.Context, entitySlug, id string) (datamodel.Record, error) {
	var out datamodel.Record
	err := s.c.Do(ctx, http.MethodGet, recordsBasePath+"/"+entitySlug+"/"+id, nil, &out)
	return out, err
}

// Create posts a new record.
func (s *RecordService) Create(ctx context.Context, entitySlug string, data datamodel.Record) (datamodel.Record, error) {
	var out datamodel.Record
	err := s.c.Do(ctx, http.MethodPost, recordsBasePath+"/"+entitySlug, data, &out)
	return out, err
}

// Update patches a record.
func (s *RecordService) Update(ctx context.Context, entitySlug, id string, data datamodel.Record) (datamodel.Record, error) {
	var out datamodel.Record
	err := s.c.Do(ctx, http.MethodPatch, recordsBasePath+"/"+entitySlug+"/"+id, data, &out)
	return out, err
}

// Delete removes a record.
func (s *RecordService) Delete(ctx context.Context, entitySlug, id string) error {
	return s.c.Do(ctx, http.MethodDelete, recordsBasePath+"/"+entitySlug+"/"+id, nil, nil)
}

// BatchUpsert posts a batch of upsert transactions; each transaction
// succeeds or fails independently and is reported per-row in the response.
func (s *RecordService) BatchUpsert(ctx context.Context, entitySlug string, txns []BatchUpsertTransaction) (BatchUpsertRecordsResponse, error) {
	var out BatchUpsertRecordsResponse
	err := s.c.Do(ctx, http.MethodPost, batchRecordsBasePath+"/"+entitySlug+"/upsert", txns, &out)
	return out, err
}
