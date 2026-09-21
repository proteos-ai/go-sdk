package agent

import (
	"context"
	"net/http"

	agentmodel "go.proteos.ai/model/agent"
	agentapi "go.proteos.ai/model/agent/api"
	sdk "go.proteos.ai/sdk"
	"go.proteos.ai/sdk/internal/httpx"
)

const modelsBasePath = "/agents/v1/models"

// ModelService exposes the stateless model calls — one request in, one reply
// out, no session:
//
//   - Generate: free text or schema-constrained structured output from an LLM.
//   - Decide: calibrated typed answers (boolean | choice | score) with
//     probabilities and confidence for snap judgments inside code paths. No
//     prose comes back; compose the answers in code and threshold on confidence.
//
// Both endpoints return BARE objects (no {data} envelope) and need the
// agent-sessions:write permission. DecisionResult.Answers decodes into the
// typed variant union (BooleanAnswer / ChoiceAnswer / ScoreAnswer).
type ModelService struct{ c *sdk.Client }

// Generate runs a single model call. A generation can legitimately take minutes
// on a long structured reply — pass sdk request options (e.g. a per-call timeout)
// when the client default is too tight.
func (s *ModelService) Generate(ctx context.Context, req agentapi.GenerateRequest, opts ...httpx.RequestOption) (agentmodel.GenerationResult, error) {
	var out agentmodel.GenerationResult
	err := s.c.Do(ctx, http.MethodPost, modelsBasePath+"/generate", req, &out, opts...)
	return out, err
}

// Decide answers typed questions about a text state with calibrated
// probabilities. Semantic limits (option / level counts, empty questions,
// unknown model ids) are the provider's and surface as a 400
// decision_provider_rejected_request carrying its message; an unconfigured
// provider answers 503 decision_provider_unavailable.
func (s *ModelService) Decide(ctx context.Context, req agentapi.DecideRequest, opts ...httpx.RequestOption) (agentmodel.DecisionResult, error) {
	var out agentmodel.DecisionResult
	err := s.c.Do(ctx, http.MethodPost, modelsBasePath+"/decide", req, &out, opts...)
	return out, err
}
