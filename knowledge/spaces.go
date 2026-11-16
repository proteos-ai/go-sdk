package knowledge

import (
	"context"
	"net/http"

	knowledgemodel "go.proteos.ai/model/knowledge"
	sdk "go.proteos.ai/sdk"
)

const spacesBasePath = "/knowledge/v1/spaces"

// SpaceServiceAPI is the contract a SpaceService satisfies.
type SpaceServiceAPI interface {
	EnsurePersonal(ctx context.Context) (knowledgemodel.KnowledgeSpace, error)
}

// SpaceService reaches knowledge spaces — the containers nodes live in.
type SpaceService struct{ c *sdk.Client }

var _ SpaceServiceAPI = (*SpaceService)(nil)

// EnsurePersonal returns the calling user's own personal space, creating it on
// first use. It takes no arguments: whose space it is comes from the token.
func (s *SpaceService) EnsurePersonal(ctx context.Context) (knowledgemodel.KnowledgeSpace, error) {
	var out knowledgemodel.KnowledgeSpace
	err := s.c.Do(ctx, http.MethodPost, spacesBasePath+"/actions/ensure-personal", nil, &out)
	return out, err
}
