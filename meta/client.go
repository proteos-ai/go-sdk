package meta

import sdk "go.proteos.ai/sdk"

// Client groups the metadata services. Construct with New, then access the
// resource services via the public fields:
//
//	m := meta.New(c)
//	entity, err := m.Entities.Get(ctx, "customer")
type Client struct {
	Entities         *EntityService
	EntityExtensions *EntityExtensionService
	PageExtensions   *PageExtensionService
	ListExtensions   *ListExtensionService
	// MenuConfigurationExtensions are items contributed to a host menu by a
	// resource that does not own it.
	MenuConfigurationExtensions *MenuConfigurationExtensionService
	Modules                     *ModuleService
	Variables                   *VariableService
	Components                  *ComponentService
	Lists                       *ListService
	ListViews                   *ListViewService
	Pages                       *PageService
	MenuConfigurations          *MenuConfigurationService
	AppConfigurations           *AppConfigurationService
	Apps                        *AppService
	DesignReferences            *DesignReferenceService
}

// New builds a Client backed by the given *sdk.Client.
func New(c *sdk.Client) *Client {
	return &Client{
		Entities:                    &EntityService{c: c},
		EntityExtensions:            &EntityExtensionService{c: c},
		PageExtensions:              &PageExtensionService{c: c},
		ListExtensions:              &ListExtensionService{c: c},
		MenuConfigurationExtensions: &MenuConfigurationExtensionService{c: c},
		Modules:                     &ModuleService{c: c},
		Variables:                   &VariableService{c: c},
		Components:                  &ComponentService{c: c},
		Lists:                       &ListService{c: c},
		ListViews:                   &ListViewService{c: c},
		Pages:                       &PageService{c: c},
		MenuConfigurations:          &MenuConfigurationService{c: c},
		AppConfigurations:           &AppConfigurationService{c: c},
		Apps:                        &AppService{c: c},
		DesignReferences:            &DesignReferenceService{c: c},
	}
}
