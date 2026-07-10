package agent

import sdk "go.proteos.ai/sdk"

// Client groups the agent-service resource services. Construct with New, then
// access them via the public fields:
//
//	a := agent.New(c)
//	ag, err := a.Agents.Get(ctx, "support-bot")
type Client struct {
	Agents     *AgentService
	Prompts    *PromptService
	Tools      *ToolService
	McpServers *McpServerService
	Skills     *SkillService
}

// New builds a Client backed by the given *sdk.Client.
func New(c *sdk.Client) *Client {
	return &Client{
		Agents:     &AgentService{c: c},
		Prompts:    &PromptService{c: c},
		Tools:      &ToolService{c: c},
		McpServers: &McpServerService{c: c},
		Skills:     &SkillService{c: c},
	}
}
