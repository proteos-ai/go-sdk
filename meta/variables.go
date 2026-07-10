package meta

import (
	"context"
	"net/http"
	metamodel "go.proteos.ai/model/meta"
	sdk "go.proteos.ai/sdk"
)

const variablesBasePath = "/meta/v1/variables"

// VariableServiceAPI is the contract a VariableService satisfies.
type VariableServiceAPI interface {
	List(opts *ListVariablesOptions) *sdk.PageIterator[metamodel.Variable, ListVariablesOptions]
	ListPage(ctx context.Context, opts *ListVariablesOptions) (sdk.ListResult[metamodel.Variable], error)
	Get(ctx context.Context, id string) (metamodel.Variable, error)
	Create(ctx context.Context, req CreateVariableRequest) (metamodel.Variable, error)
	Update(ctx context.Context, id string, req UpdateVariableRequest) (metamodel.Variable, error)
	Delete(ctx context.Context, id string) error
}

// VariableService manages module configuration variables.
type VariableService struct{ c *sdk.Client }

var _ VariableServiceAPI = (*VariableService)(nil)

func (s *VariableService) List(opts *ListVariablesOptions) *sdk.PageIterator[metamodel.Variable, ListVariablesOptions] {
	o := ListVariablesOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListVariablesOptions) (sdk.ListResult[metamodel.Variable], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

func (s *VariableService) ListPage(ctx context.Context, opts *ListVariablesOptions) (sdk.ListResult[metamodel.Variable], error) {
	var out sdk.ListResult[metamodel.Variable]
	err := s.c.DoWithQuery(ctx, http.MethodGet, variablesBasePath, opts, nil, &out)
	return out, err
}

func (s *VariableService) Get(ctx context.Context, id string) (metamodel.Variable, error) {
	var out metamodel.Variable
	err := s.c.Do(ctx, http.MethodGet, variablesBasePath+"/"+id, nil, &out)
	return out, err
}

func (s *VariableService) Create(ctx context.Context, req CreateVariableRequest) (metamodel.Variable, error) {
	var out metamodel.Variable
	err := s.c.Do(ctx, http.MethodPost, variablesBasePath, req, &out)
	return out, err
}

func (s *VariableService) Update(ctx context.Context, id string, req UpdateVariableRequest) (metamodel.Variable, error) {
	var out metamodel.Variable
	err := s.c.Do(ctx, http.MethodPatch, variablesBasePath+"/"+id, req, &out)
	return out, err
}

func (s *VariableService) Delete(ctx context.Context, id string) error {
	return s.c.Do(ctx, http.MethodDelete, variablesBasePath+"/"+id, nil, nil)
}
