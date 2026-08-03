package conversation

import sdk "go.proteos.ai/sdk"

// Client groups the conversation-service resource services. Construct with
// New, then access them via the public fields:
//
//	c := conversation.New(client)
//	t, err := c.ConversationTypes.Get(ctx, "sales-discovery")
type Client struct {
	ConversationTypes *ConversationTypeService
}

// New builds a Client backed by the given *sdk.Client.
func New(c *sdk.Client) *Client {
	return &Client{
		ConversationTypes: &ConversationTypeService{c: c},
	}
}
