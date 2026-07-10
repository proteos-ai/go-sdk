// Package auth provides services for users, roles, organizations, and
// the permissions/role-assignment surface of the Proteos auth API.
package account

import (
	"time"

	"go.proteos.ai/model/common"
)

// AuditFields are the createdAt/updatedAt/createdBy/updatedBy columns
// embedded in every auth resource. createdBy/updatedBy hold the common.UserRef
// ({type, id}) of the user who made the change (the "platform" sentinel id for
// system writes).
type AuditFields struct {
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	CreatedBy common.UserRef `json:"created_by"`
	UpdatedBy common.UserRef `json:"updated_by"`
}

// ListOptions are pagination + sort fields shared across every list endpoint.
// Page and PageSize are always sent (page is 0-indexed; the server treats a
// missing value differently than 0, so we send it explicitly).
type ListOptions struct {
	Page      int    `query:"page"`
	PageSize  int    `query:"page_size"`
	SortBy    string `query:"sort_by,omitempty"`
	SortOrder string `query:"sort_direction,omitempty"`
}

// Permission is the access level granted to a role for an entity.
type Permission string

const (
	PermissionRead   Permission = "read"
	PermissionWrite  Permission = "write"
	PermissionDelete Permission = "delete"
)

// User is a person with access to the platform.
type User struct {
	AuditFields
	ID           string `json:"id"`
	Email        string `json:"email"`
	GivenName    string `json:"given_name"`
	FamilyName   string `json:"family_name"`
	ExternalID   string `json:"external_id"`
	DefaultOrgID string `json:"default_org_id"`
}

// CreateUserRequest is the body for POST /accounts/v1/users.
type CreateUserRequest struct {
	GivenName    string `json:"given_name"`
	FamilyName   string `json:"family_name"`
	Email        string `json:"email"`
	DefaultOrgID string `json:"default_org_id,omitempty"`
}

// UpdateUserRequest is the body for PATCH /accounts/v1/users/{id}.
// Pointer fields distinguish "absent" from "empty"; nil pointers are omitted.
type UpdateUserRequest struct {
	GivenName    *string `json:"given_name,omitempty"`
	FamilyName   *string `json:"family_name,omitempty"`
	DefaultOrgID *string `json:"default_org_id,omitempty"`
}

// ListUsersOptions filters and paginates GET /accounts/v1/users.
type ListUsersOptions struct {
	ListOptions
	ID           string `query:"id,omitempty"`
	GivenName    string `query:"given_name,omitempty"`
	FamilyName   string `query:"family_name,omitempty"`
	Email        string `query:"email,omitempty"`
	DefaultOrgID string `query:"default_org_id,omitempty"`
}

// ApiKey is a user-scoped pak_ token. Only display hints (prefix/suffix) are
// readable after creation — the full token appears once in CreatedApiKey.
type ApiKey struct {
	AuditFields
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	OrgID       string     `json:"org_id"`
	Name        string     `json:"name"`
	TokenPrefix string     `json:"token_prefix"`
	TokenSuffix string     `json:"token_suffix"`
	ExpiresAt   *time.Time `json:"expires_at"`
	LastUsedAt  *time.Time `json:"last_used_at"`
}

// CreatedApiKey is the create response: the key plus the full token, returned
// exactly once. Store it immediately — it is not retrievable afterwards.
type CreatedApiKey struct {
	ApiKey
	Token string `json:"token"`
}

// CreateApiKeyRequest is the body for POST /accounts/v1/users/{id}/api-keys.
// A nil ExpiresAt creates a key that never expires; an empty OrgID binds the
// key to the user's default org.
type CreateApiKeyRequest struct {
	Name      string     `json:"name"`
	OrgID     string     `json:"org_id,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// Role is a named bundle of permissions in an organization.
type Role struct {
	AuditFields
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	OrgID       string `json:"org_id"`
	Description string `json:"description"`
}

type CreateRoleRequest struct {
	OrgID       string `json:"org_id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type UpdateRoleRequest struct {
	OrgID       *string `json:"org_id,omitempty"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

type ListRolesOptions struct {
	ListOptions
	OrgID       string `query:"org_id,omitempty"`
	Slug        string `query:"slug,omitempty"`
	Name        string `query:"name,omitempty"`
	Description string `query:"description,omitempty"`
}

// Organization is a tenant in the platform.
type Organization struct {
	AuditFields
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateOrganizationRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type UpdateOrganizationRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

type ListOrganizationsOptions struct {
	ListOptions
	ID          string `query:"id,omitempty"`
	Name        string `query:"name,omitempty"`
	Description string `query:"description,omitempty"`
}

// RoleEntityPermission is a permission granted to a role for a specific entity.
type RoleEntityPermission struct {
	AuditFields
	ID         string     `json:"id"`
	RoleSlug   string     `json:"role_slug"`
	EntitySlug string     `json:"entity_slug"`
	Permission Permission `json:"permission"`
}

// AssignPermissionRequest is the body for assigning a permission to a role.
type AssignPermissionRequest struct {
	EntitySlug string     `json:"entity_slug"`
	Permission Permission `json:"permission"`
}

// ListRolePermissionsOptions filters role permissions.
type ListRolePermissionsOptions struct {
	ListOptions
	ID         string     `query:"id,omitempty"`
	EntitySlug string     `query:"entity_slug,omitempty"`
	Permission Permission `query:"permission,omitempty"`
}

// UserRoleAssignment links a user to a role.
type UserRoleAssignment struct {
	AuditFields
	ID       string `json:"id"`
	UserID   string `json:"user_id"`
	RoleSlug string `json:"role_slug"`
	OrgID    string `json:"org_id"`
}

// AssignRoleRequest is the body for POST /accounts/v1/users/{id}/roles.
type AssignRoleRequest struct {
	RoleSlug string `json:"role_slug"`
	// OrgID optionally targets the org the assignment lands in. Honored only for
	// platform admins (and the system); a regular caller's value is ignored and
	// pinned to their token org. Empty = use the token org.
	OrgID string `json:"org_id,omitempty"`
}

// ListUserRoleAssignmentsOptions filters user role assignments.
type ListUserRoleAssignmentsOptions struct {
	ListOptions
	ID       string `query:"id,omitempty"`
	RoleSlug string `query:"role_slug,omitempty"`
}
