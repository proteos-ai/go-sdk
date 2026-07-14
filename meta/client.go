package meta

import sdk "go.proteos.ai/sdk"

// Client groups the metadata services. Construct with New, then access the
// resource services via the public fields:
//
//	m := meta.New(c)
//	entity, err := m.Entities.Get(ctx, "customer")
type Client struct {
	Entities           *EntityService
	Modules            *ModuleService
	Variables          *VariableService
	Components         *ComponentService
	Lists              *ListService
	ListViews          *ListViewService
	Pages              *PageService
	MenuConfigurations *MenuConfigurationService
	Apps               *AppService
	DesignReferences   *DesignReferenceService
}

// New builds a Client backed by the given *sdk.Client.
func New(c *sdk.Client) *Client {
	return &Client{
		Entities:           &EntityService{c: c},
		Modules:            &ModuleService{c: c},
		Variables:          &VariableService{c: c},
		Components:         &ComponentService{c: c},
		Lists:              &ListService{c: c},
		ListViews:          &ListViewService{c: c},
		Pages:              &PageService{c: c},
		MenuConfigurations: &MenuConfigurationService{c: c},
		Apps:               &AppService{c: c},
		DesignReferences:   &DesignReferenceService{c: c},
	}
}
