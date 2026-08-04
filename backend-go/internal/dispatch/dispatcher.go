package dispatch

import (
	"context"
	"fmt"

	agentfeedpb "github.com/agent-feed/backend-go/gen/agentfeedpb"
	"github.com/agent-feed/backend-go/internal/constants"
	"github.com/agent-feed/backend-go/internal/db/model"
	"github.com/agent-feed/backend-go/internal/logs"
)

const (
	respondNamespace = "respond"
	proactiveNamespace = "proactive"
)

func (d *Dispatcher) think(ctx context.Context, mode string, agent model.Agent, post model.Post, recent []model.Post, memories []string, agentMap map[int64]model.Agent) (*Decision, error) {
	protoRecent := make([]*agentfeedpb.Post, 0, len(recent))
	for i := range recent {
		protoRecent = append(protoRecent, toProtoPost(recent[i], agentMap))
	}
	logs.AgentInfo(agent.ID, agent.Name, agent.Handle, constants.AiWorkerNamespace, fmt.Sprintf("Dispatched LLM deliberation request to worker (mode: %s)", mode))
	resp, err := d.cli.Think(ctx, &agentfeedpb.ThinkRequest{
		Agent:    toProtoAgent(agent),
		Post:     toProtoPost(post, agentMap),
		Recent:   protoRecent,
		Mode:     mode,
		Memories: memories,
	})
	if err != nil {
		logs.AgentError(agent.ID, agent.Name, agent.Handle, constants.AiWorkerNamespace, fmt.Sprintf("AI Worker gRPC error: %v", err))
		return nil, fmt.Errorf("worker think: %w", err)
	}
	logs.AgentInfo(agent.ID, agent.Name, agent.Handle, constants.AiWorkerNamespace, fmt.Sprintf("Received AI worker decision: action=%s", resp.Action))
	return &Decision{Action: ActionType(resp.Action), Text: resp.Text}, nil
}

func (d *Dispatcher) Think(ctx context.Context, agent model.Agent, post model.Post, recent []model.Post, memories []string, agentMap map[int64]model.Agent) (*Decision, error) {
	return d.think(ctx, respondNamespace, agent, post, recent, memories, agentMap)
}

func (d *Dispatcher) Proactive(ctx context.Context, agent model.Agent, recent []model.Post, memories []string, agentMap map[int64]model.Agent) (*Decision, error) {
	return d.think(ctx, proactiveNamespace, agent, model.Post{}, recent, memories, agentMap)
}


