package meta

import (
	"context"
	"fmt"
	"net/http"

	metamodel "go.proteos.ai/model/meta"
	sdk "go.proteos.ai/sdk"
)

const designReferencesBasePath = "/meta/v1/design-references"

// DesignReferenceServiceAPI is the contract a DesignReferenceService satisfies.
//
// The markdown body is intentionally split from the metadata methods: List/Get
// never carry `content` — read it with GetContent and write it with SetContent
// (Create may seed it in one shot).
type DesignReferenceServiceAPI interface {
	List(opts *ListDesignReferencesOptions) *sdk.PageIterator[metamodel.DesignReference, ListDesignReferencesOptions]
	ListPage(ctx context.Context, opts *ListDesignReferencesOptions) (sdk.ListResult[metamodel.DesignReference], error)
	Get(ctx context.Context, id string) (metamodel.DesignReference, error)
	GetBySlug(ctx context.Context, slug string) (metamodel.DesignReference, error)
	Create(ctx context.Context, req CreateDesignReferenceRequest) (metamodel.DesignReference, error)
	Upsert(ctx context.Context, req CreateDesignReferenceRequest) (metamodel.DesignReference, error)
	Update(ctx context.Context, id string, req UpdateDesignReferenceRequest) (metamodel.DesignReference, error)
	Delete(ctx context.Context, id string) error
	GetContent(ctx context.Context, id string) (string, error)
	SetContent(ctx context.Context, id string, content string) error
}

// DesignReferenceService manages an org's stored DESIGN.md documents.
type DesignReferenceService struct{ c *sdk.Client }

var _ DesignReferenceServiceAPI = (*DesignReferenceService)(nil)

func (s *DesignReferenceService) List(opts *ListDesignReferencesOptions) *sdk.PageIterator[metamodel.DesignReference, ListDesignReferencesOptions] {
	o := ListDesignReferencesOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListDesignReferencesOptions) (sdk.ListResult[metamodel.DesignReference], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

func (s *DesignReferenceService) ListPage(ctx context.Context, opts *ListDesignReferencesOptions) (sdk.ListResult[metamodel.DesignReference], error) {
	var out sdk.ListResult[metamodel.DesignReference]
	err := s.c.DoWithQuery(ctx, http.MethodGet, designReferencesBasePath, opts, nil, &out)
	return out, err
}

func (s *DesignReferenceService) Get(ctx context.Context, id string) (metamodel.DesignReference, error) {
	var out metamodel.DesignReference
	err := s.c.Do(ctx, http.MethodGet, designReferencesBasePath+"/"+id, nil, &out)
	return out, err
}

// GetBySlug resolves a reference by its per-org slug via the exact-match list
// filter (the metadata rows are content-free, same as List/Get).
func (s *DesignReferenceService) GetBySlug(ctx context.Context, slug string) (metamodel.DesignReference, error) {
	page, err := s.ListPage(ctx, &ListDesignReferencesOptions{Slug: slug})
	if err != nil {
		return metamodel.DesignReference{}, err
	}
	if len(page.Data) == 0 {
		return metamodel.DesignReference{}, fmt.Errorf("design reference with slug %q not found", slug)
	}
	return page.Data[0], nil
}

func (s *DesignReferenceService) Create(ctx context.Context, req CreateDesignReferenceRequest) (metamodel.DesignReference, error) {
	var out metamodel.DesignReference
	err := s.c.Do(ctx, http.MethodPost, designReferencesBasePath, req, &out)
	return out, err
}

func (s *DesignReferenceService) Upsert(ctx context.Context, req CreateDesignReferenceRequest) (metamodel.DesignReference, error) {
	var out metamodel.DesignReference
	err := s.c.Do(ctx, http.MethodPost, designReferencesBasePath+"/upsert", req, &out)
	return out, err
}

func (s *DesignReferenceService) Update(ctx context.Context, id string, req UpdateDesignReferenceRequest) (metamodel.DesignReference, error) {
	var out metamodel.DesignReference
	err := s.c.Do(ctx, http.MethodPatch, designReferencesBasePath+"/"+id, req, &out)
	return out, err
}

func (s *DesignReferenceService) Delete(ctx context.Context, id string) error {
	return s.c.Do(ctx, http.MethodDelete, designReferencesBasePath+"/"+id, nil, nil)
}

// GetContent fetches the markdown body for a reference.
func (s *DesignReferenceService) GetContent(ctx context.Context, id string) (string, error) {
	var out DesignReferenceContentResponse
	err := s.c.Do(ctx, http.MethodGet, designReferencesBasePath+"/"+id+"/content", nil, &out)
	return out.Content, err
}

// SetContent overwrites the markdown body for a reference (metadata untouched).
func (s *DesignReferenceService) SetContent(ctx context.Context, id string, content string) error {
	return s.c.Do(ctx, http.MethodPut, designReferencesBasePath+"/"+id+"/content", SetDesignReferenceContentRequest{Content: content}, nil)
}
