// Package sdk is the Go client for the Proteos platform API.
//
// It mirrors the surface of @proteos/sdk (TypeScript). The base sdk package
// provides the HTTP client, error type, and pagination iterator. The
// sub-packages account, meta, and data expose the resource services
// (UserService, EntityService, RecordService, etc.) layered on top.
//
//	c, err := sdk.NewClientFromEnv()
//	if err != nil { return err }
//
//	m := meta.New(c)
//	page, err := m.Entities.ListPage(ctx, nil)
//
// Errors from any service method can be inspected with the IsNotFound,
// IsUnauthorized, IsForbidden, IsBadRequest, and IsConflict helpers, or
// matched via errors.As to *sdk.Error.
package sdk
