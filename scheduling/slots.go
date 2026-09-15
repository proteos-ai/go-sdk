package scheduling

import (
	"context"
	"net/http"

	schedulingapi "go.proteos.ai/model/scheduling/api"
	sdk "go.proteos.ai/sdk"
)

// SlotService is the slot engine: the bookable starts of some hosts under a
// calendar event type's (or explicit) rules.
type SlotService struct{ c *sdk.Client }

// Get computes slots for the request (POST /slots — the request is a body,
// not a query).
func (s *SlotService) Get(ctx context.Context, req schedulingapi.GetSlotsRequest) (schedulingapi.GetSlotsResponse, error) {
	var out schedulingapi.GetSlotsResponse
	err := s.c.Do(ctx, http.MethodPost, basePath+"/slots", req, &out)
	return out, err
}
