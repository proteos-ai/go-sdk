package scheduling

import (
	"context"
	"net/http"
	"net/url"

	schedulingmodel "go.proteos.ai/model/scheduling"
	schedulingapi "go.proteos.ai/model/scheduling/api"
	sdk "go.proteos.ai/sdk"
)

// SchedulingProfileService reads and replaces per-user scheduling settings
// (one per org + user; the service answers defaults when none is stored).
type SchedulingProfileService struct{ c *sdk.Client }

func mineSchedulingProfilePath() string {
	return basePath + "/me/scheduling-profile"
}

func userSchedulingProfilePath(userId string) string {
	return basePath + "/users/" + url.PathEscape(userId) + "/scheduling-profile"
}

func (s *SchedulingProfileService) get(ctx context.Context, path string) (schedulingmodel.SchedulingProfile, error) {
	var out schedulingmodel.SchedulingProfile
	err := s.c.Do(ctx, http.MethodGet, path, nil, &out)
	return out, err
}

func (s *SchedulingProfileService) put(ctx context.Context, path string, req schedulingapi.PutSchedulingProfileRequest) (schedulingmodel.SchedulingProfile, error) {
	var out schedulingmodel.SchedulingProfile
	err := s.c.Do(ctx, http.MethodPut, path, req, &out)
	return out, err
}

func (s *SchedulingProfileService) GetMine(ctx context.Context) (schedulingmodel.SchedulingProfile, error) {
	return s.get(ctx, mineSchedulingProfilePath())
}

// PutMine replaces the caller's scheduling profile (PUT semantics: every
// field is the new value).
func (s *SchedulingProfileService) PutMine(ctx context.Context, req schedulingapi.PutSchedulingProfileRequest) (schedulingmodel.SchedulingProfile, error) {
	return s.put(ctx, mineSchedulingProfilePath(), req)
}

func (s *SchedulingProfileService) GetForUser(ctx context.Context, userId string) (schedulingmodel.SchedulingProfile, error) {
	return s.get(ctx, userSchedulingProfilePath(userId))
}

func (s *SchedulingProfileService) PutForUser(ctx context.Context, userId string, req schedulingapi.PutSchedulingProfileRequest) (schedulingmodel.SchedulingProfile, error) {
	return s.put(ctx, userSchedulingProfilePath(userId), req)
}
