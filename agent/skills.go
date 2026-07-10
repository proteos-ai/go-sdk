package agent

import (
	"context"
	"io"
	"net/http"
	"strconv"

	agentmodel "go.proteos.ai/model/agent"
	sdk "go.proteos.ai/sdk"
	"go.proteos.ai/sdk/internal/httpx"
)

const skillsBasePath = "/agents/v1/skills"

// SkillService manages versioned skill bundles. Skills are created/versioned via a
// multipart bundle upload (Deploy); there is no JSON create/update.
type SkillService struct{ c *sdk.Client }

func (s *SkillService) List(opts *ListSkillsOptions) *sdk.PageIterator[agentmodel.Skill, ListSkillsOptions] {
	o := ListSkillsOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListSkillsOptions) (sdk.ListResult[agentmodel.Skill], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

func (s *SkillService) ListPage(ctx context.Context, opts *ListSkillsOptions) (sdk.ListResult[agentmodel.Skill], error) {
	var out sdk.ListResult[agentmodel.Skill]
	err := s.c.DoWithQuery(ctx, http.MethodGet, skillsBasePath, opts, nil, &out)
	return out, err
}

func (s *SkillService) Get(ctx context.Context, key string) (agentmodel.Skill, error) {
	var out agentmodel.Skill
	err := s.c.Do(ctx, http.MethodGet, skillsBasePath+"/"+key, nil, &out)
	return out, err
}

// Deploy uploads a skill bundle (a tar.gz / zip with SKILL.md at its root) as a
// multipart `bundle` part, with an optional display_name override and an optional
// moduleSlug that tags the skill with its owning module (empty for a standalone
// deploy). The skill key is derived server-side from the SKILL.md frontmatter
// `name`. Idempotent: re-deploying an identical bundle (same checksum) returns the
// existing skill without a new version. Calls `POST /agents/v1/skills/deploy`.
func (s *SkillService) Deploy(ctx context.Context, displayName string, moduleSlug string, bundle io.Reader, filename string) (agentmodel.Skill, error) {
	var out agentmodel.Skill
	fields := map[string]string{}
	if displayName != "" {
		fields["display_name"] = displayName
	}
	if moduleSlug != "" {
		fields["module_slug"] = moduleSlug
	}
	err := s.c.DoMultipart(ctx, http.MethodPost, skillsBasePath+"/deploy", fields,
		httpx.MultipartFile{
			FieldName:   "bundle",
			Filename:    filename,
			ContentType: "application/gzip",
			Reader:      bundle,
		},
		&out,
	)
	return out, err
}

func (s *SkillService) Delete(ctx context.Context, key string) error {
	return s.c.Do(ctx, http.MethodDelete, skillsBasePath+"/"+key, nil, nil)
}

// Versions returns every immutable version of a skill (newest first).
func (s *SkillService) Versions(ctx context.Context, key string) ([]agentmodel.SkillVersion, error) {
	var out struct {
		Data []agentmodel.SkillVersion `json:"data"`
	}
	err := s.c.Do(ctx, http.MethodGet, skillsBasePath+"/"+key+"/versions", nil, &out)
	return out.Data, err
}

// Version returns a single skill version by number.
func (s *SkillService) Version(ctx context.Context, key string, number uint32) (agentmodel.SkillVersion, error) {
	var out agentmodel.SkillVersion
	err := s.c.Do(ctx, http.MethodGet, skillsBasePath+"/"+key+"/versions/"+strconv.FormatUint(uint64(number), 10), nil, &out)
	return out, err
}
