package agent

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
)

type Client struct {
	api    anthropic.Client
	config ClientConfig
}

type ClientConfig struct {
	AgentID       string
	EnvironmentID string
	GitHubToken   string
}

func NewClient(cfg ClientConfig) *Client {
	return &Client{
		api:    anthropic.NewClient(),
		config: cfg,
	}
}

func (c *Client) EnsureAgent(ctx context.Context, name, system string) (string, error) {
	if c.config.AgentID != "" {
		return c.config.AgentID, nil
	}

	agent, err := c.api.Beta.Agents.New(ctx, anthropic.BetaAgentNewParams{
		Name: name,
		Model: anthropic.BetaManagedAgentsModelConfigParams{
			ID: anthropic.BetaManagedAgentsModelClaudeSonnet4_6,
		},
		System: anthropic.Opt(system),
		Tools: []anthropic.BetaAgentNewParamsToolUnion{{
			OfAgentToolset20260401: &anthropic.BetaManagedAgentsAgentToolset20260401Params{
				Type: anthropic.BetaManagedAgentsAgentToolset20260401ParamsTypeAgentToolset20260401,
			},
		}},
	})
	if err != nil {
		return "", fmt.Errorf("creating agent: %w", err)
	}
	c.config.AgentID = agent.ID
	return agent.ID, nil
}

func (c *Client) EnsureEnvironment(ctx context.Context) (string, error) {
	if c.config.EnvironmentID != "" {
		return c.config.EnvironmentID, nil
	}

	env, err := c.api.Beta.Environments.New(ctx, anthropic.BetaEnvironmentNewParams{
		Name: "agora-agent-env",
		Config: anthropic.BetaCloudConfigParams{
			Networking: anthropic.BetaCloudConfigParamsNetworkingUnion{
				OfUnrestricted: &anthropic.BetaUnrestrictedNetworkParam{},
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("creating environment: %w", err)
	}
	c.config.EnvironmentID = env.ID
	return env.ID, nil
}

type RunParams struct {
	Prompt string
	Repos  []RepoMount
}

type RepoMount struct {
	URL       string
	MountPath string
}

func (c *Client) Run(ctx context.Context, params RunParams) (<-chan Event, error) {
	agentID, err := c.EnsureAgent(ctx, "agora-agent", "You are a cross-repository analysis agent. You have access to the company's shared knowledge base (agora/) and can analyze codebases.")
	if err != nil {
		return nil, err
	}

	envID, err := c.EnsureEnvironment(ctx)
	if err != nil {
		return nil, err
	}

	// Build resources
	var resources []anthropic.BetaSessionNewParamsResourceUnion
	for _, repo := range params.Repos {
		resources = append(resources, anthropic.BetaSessionNewParamsResourceUnion{
			OfGitHubRepository: &anthropic.BetaManagedAgentsGitHubRepositoryResourceParams{
				Type:               anthropic.BetaManagedAgentsGitHubRepositoryResourceParamsTypeGitHubRepository,
				URL:                repo.URL,
				MountPath:          anthropic.String(repo.MountPath),
				AuthorizationToken: c.config.GitHubToken,
			},
		})
	}

	session, err := c.api.Beta.Sessions.New(ctx, anthropic.BetaSessionNewParams{
		Agent:         anthropic.BetaSessionNewParamsAgentUnion{OfString: anthropic.String(agentID)},
		EnvironmentID: envID,
		Resources:     resources,
	})
	if err != nil {
		return nil, fmt.Errorf("creating session: %w", err)
	}

	// Open stream
	stream := c.api.Beta.Sessions.Events.StreamEvents(ctx, session.ID, anthropic.BetaSessionEventStreamParams{})

	// Send user message
	_, err = c.api.Beta.Sessions.Events.Send(ctx, session.ID, anthropic.BetaSessionEventSendParams{
		Events: []anthropic.BetaManagedAgentsEventParamsUnion{{
			OfUserMessage: &anthropic.BetaManagedAgentsUserMessageEventParams{
				Type: anthropic.BetaManagedAgentsUserMessageEventParamsTypeUserMessage,
				Content: []anthropic.BetaManagedAgentsUserMessageEventParamsContentUnion{{
					OfText: &anthropic.BetaManagedAgentsTextBlockParam{
						Type: anthropic.BetaManagedAgentsTextBlockTypeText,
						Text: params.Prompt,
					},
				}},
			},
		}},
	})
	if err != nil {
		stream.Close()
		return nil, fmt.Errorf("sending message: %w", err)
	}

	// Stream events to channel
	events := make(chan Event, 100)
	go func() {
		defer close(events)
		defer stream.Close()
		for stream.Next() {
			raw := stream.Current()
			switch ev := raw.AsAny().(type) {
			case anthropic.BetaManagedAgentsAgentMessageEvent:
				for _, block := range ev.Content {
					events <- Event{Type: "message", Text: block.Text}
				}
			case anthropic.BetaManagedAgentsAgentToolUseEvent:
				events <- Event{Type: "tool_use", Text: ev.Name}
			case anthropic.BetaManagedAgentsSessionStatusIdleEvent:
				events <- Event{Type: "done"}
				return
			case anthropic.BetaManagedAgentsSessionStatusTerminatedEvent:
				events <- Event{Type: "error", Text: "session terminated"}
				return
			}
		}
		if err := stream.Err(); err != nil {
			events <- Event{Type: "error", Text: err.Error()}
		}
	}()

	return events, nil
}

type Event struct {
	Type string // "message", "tool_use", "done", "error"
	Text string
}
