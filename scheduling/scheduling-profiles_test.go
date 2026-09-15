package scheduling_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	conversationmodel "go.proteos.ai/model/conversation"
	schedulingmodel "go.proteos.ai/model/scheduling"
	schedulingapi "go.proteos.ai/model/scheduling/api"
)

func TestSchedulingProfileService_Paths(t *testing.T) {
	var method, path, body string
	s := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		body = string(raw)
		_ = json.NewEncoder(w).Encode(schedulingmodel.SchedulingProfile{Id: "sp1", UserId: "u1", Timezone: "Europe/Berlin"})
	})
	ctx := context.Background()
	req := schedulingapi.PutSchedulingProfileRequest{
		Timezone:    "Europe/Berlin",
		WeeklyHours: []conversationmodel.WindowDay{{Day: conversationmodel.WeekdayMonday, From: "09:00", Until: "17:00"}},
		DisplayName: "Tonio",
		IsBookable:  true,
	}

	got, err := s.SchedulingProfiles.GetMine(ctx)
	require.NoError(t, err)
	require.Equal(t, http.MethodGet, method)
	require.Equal(t, "/scheduling/v1/me/scheduling-profile", path)
	require.Equal(t, "Europe/Berlin", got.Timezone)

	_, err = s.SchedulingProfiles.PutMine(ctx, req)
	require.NoError(t, err)
	require.Equal(t, http.MethodPut, method)
	require.Equal(t, "/scheduling/v1/me/scheduling-profile", path)
	require.JSONEq(t, `{"timezone":"Europe/Berlin","weekly_hours":[{"day":"monday","from":"09:00","until":"17:00"}],"display_name":"Tonio","headline":"","is_bookable":true}`, body)

	_, err = s.SchedulingProfiles.GetForUser(ctx, "u1")
	require.NoError(t, err)
	require.Equal(t, http.MethodGet, method)
	require.Equal(t, "/scheduling/v1/users/u1/scheduling-profile", path)

	_, err = s.SchedulingProfiles.PutForUser(ctx, "u1", req)
	require.NoError(t, err)
	require.Equal(t, http.MethodPut, method)
	require.Equal(t, "/scheduling/v1/users/u1/scheduling-profile", path)
}
