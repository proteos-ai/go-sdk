package account

import sdk "go.proteos.ai/sdk"

// Client groups the account namespace services. Construct with New and then
// access the resource services via the public fields:
//
//	a := account.New(c)
//	user, err := a.Users.Get(ctx, id)
type Client struct {
	Users         *UserService
	Roles         *RoleService
	Organizations *OrganizationService
	PlatformRoles *PlatformRoleService
	Me            *MeService
}

// New builds a Client backed by the given *sdk.Client.
func New(c *sdk.Client) *Client {
	return &Client{
		Users:         &UserService{c: c},
		Roles:         &RoleService{c: c},
		Organizations: &OrganizationService{c: c},
		PlatformRoles: &PlatformRoleService{c: c},
		Me:            &MeService{c: c},
	}
}
