package data

import (
	"context"
	"net/http"

	sdk "go.proteos.ai/sdk"
)

const recordDraftsBasePath = "/data/v1/record-drafts"

// RecordDraftServiceAPI is the contract a RecordDraftService satisfies.
type RecordDraftServiceAPI interface {
	ListPage(ctx context.Context, opts *ListRecordDraftsOptions) (sdk.ListResult[RecordDraft], error)
	Get(ctx context.Context, id string) (RecordDraft, error)
	Create(ctx context.Context, req CreateRecordDraftRequest) (RecordDraft, error)
	Approve(ctx context.Context, id string, req ApproveRecordDraftRequest) (RecordDraft, error)
	Reject(ctx context.Context, id string, req RejectRecordDraftRequest) (RecordDraft, error)
}

// RecordDraftService manages record drafts — proposed record writes awaiting
// a verdict — via the data-service. Create needs record-drafts:write; Approve
// additionally needs write (create / update / upsert drafts) or delete
// (delete drafts) on the entity, because the writes run as the caller.
// approved vs approved_with_changes is derived server-side from the diff.
type RecordDraftService struct{ c *sdk.Client }

var _ RecordDraftServiceAPI = (*RecordDraftService)(nil)

// ListPage fetches one page of the org's drafts — proposed by default (the
// review queue).
func (s *RecordDraftService) ListPage(ctx context.Context, opts *ListRecordDraftsOptions) (sdk.ListResult[RecordDraft], error) {
	var out sdk.ListResult[RecordDraft]
	err := s.c.DoWithQuery(ctx, http.MethodGet, recordDraftsBasePath, opts, nil, &out)
	return out, err
}

// Get returns a single draft.
func (s *RecordDraftService) Get(ctx context.Context, id string) (RecordDraft, error) {
	var out RecordDraft
	err := s.c.Do(ctx, http.MethodGet, recordDraftsBasePath+"/"+id, nil, &out)
	return out, err
}

// Create proposes a write — nothing is written until the draft is approved.
func (s *RecordDraftService) Create(ctx context.Context, req CreateRecordDraftRequest) (RecordDraft, error) {
	var out RecordDraft
	err := s.c.Do(ctx, http.MethodPost, recordDraftsBasePath, req, &out)
	return out, err
}

// Approve applies the draft as the caller. An empty request applies it as
// proposed; Items carries the reviewer's edited proposal. Always answers the
// draft: status approved / approved_with_changes when every item succeeded,
// else still proposed with per-item Results to fix and re-approve. 409
// record_draft_not_proposed once resolved.
func (s *RecordDraftService) Approve(ctx context.Context, id string, req ApproveRecordDraftRequest) (RecordDraft, error) {
	var out RecordDraft
	err := s.c.Do(ctx, http.MethodPost, recordDraftsBasePath+"/"+id+"/approve", req, &out)
	return out, err
}

// Reject resolves the draft without applying it. 409
// record_draft_not_proposed once resolved.
func (s *RecordDraftService) Reject(ctx context.Context, id string, req RejectRecordDraftRequest) (RecordDraft, error) {
	var out RecordDraft
	err := s.c.Do(ctx, http.MethodPost, recordDraftsBasePath+"/"+id+"/reject", req, &out)
	return out, err
}
