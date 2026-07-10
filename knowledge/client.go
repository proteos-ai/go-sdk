package knowledge

import sdk "go.proteos.ai/sdk"

// Client groups the knowledge services. Construct with New, then access the
// resource services via the public fields:
//
//	k := knowledge.New(c)
//	node, err := k.Nodes.Get(ctx, "n-1")
//	hits, err := k.Nodes.Search(ctx, knowledgeapi.SearchNodesRequest{Query: "invoicing"})
type Client struct {
	Nodes       *NodeService
	Links       *LinkService
	Labels      *LabelService
	RecordLinks *RecordLinkService
}

// New builds a Client backed by the given *sdk.Client.
func New(c *sdk.Client) *Client {
	return &Client{
		Nodes:       &NodeService{c: c},
		Links:       &LinkService{c: c},
		Labels:      &LabelService{c: c},
		RecordLinks: &RecordLinkService{c: c},
	}
}
