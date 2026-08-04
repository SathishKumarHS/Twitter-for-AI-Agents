package dispatch

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	agentfeedpb "github.com/agent-feed/backend-go/gen/agentfeedpb"
	"github.com/agent-feed/backend-go/internal/db/model"
	"github.com/agent-feed/backend-go/internal/logs"
)

// Decision is the structured output the AI worker returns for a lifecycle turn.
type Decision struct {
	Action string // post | reply | ignore
	Text   string
}
// Dispatcher forwards "think" jobs to the Python AI worker over gRPC.
type Dispatcher struct {
	conn *grpc.ClientConn
	cli  agentfeedpb.WorkerClient
}

// New opens a gRPC connection to the AI worker. Pass workerAddr like "ai-worker:50051".
func New(ctx context.Context, workerAddr string) (*Dispatcher, error) {
	conn, err := grpc.DialContext(ctx, workerAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("dial worker %s: %w", workerAddr, err)
	}
	return &Dispatcher{
		conn: conn,
		cli:  agentfeedpb.NewWorkerClient(conn),
	}, nil
}

// Close tears down the gRPC connection.
func (d *Dispatcher) Close() error {
	return d.conn.Close()
}

// Think asks the worker to reason about whether/how the agent should respond
// to a feed post. recent provides thread context (older posts in the thread)
// so replies feel aware of the conversation. memories are the agent's long-term
// history, closing the loop so decisions are informed by prior experience.
func (d *Dispatcher) Think(ctx context.Context, agent model.Agent, post model.Post, recent []model.Post, memories []string, agentMap map[int64]model.Agent) (*Decision, error) {
	return d.think(ctx, "respond", agent, post, recent, memories, agentMap)
}

// Proactive asks the worker to author fresh, original content from the agent.
// It is the "empty feed" path: with no post to react to, the agent posts on its own.
func (d *Dispatcher) Proactive(ctx context.Context, agent model.Agent, recent []model.Post, memories []string, agentMap map[int64]model.Agent) (*Decision, error) {
	return d.think(ctx, "proactive", agent, model.Post{}, recent, memories, agentMap)
}

func (d *Dispatcher) think(ctx context.Context, mode string, agent model.Agent, post model.Post, recent []model.Post, memories []string, agentMap map[int64]model.Agent) (*Decision, error) {
	protoRecent := make([]*agentfeedpb.Post, 0, len(recent))
	for i := range recent {
		protoRecent = append(protoRecent, toProtoPost(recent[i], agentMap))
	}
	logs.AgentInfo(agent.ID, agent.Name, agent.Handle, "ai-worker", fmt.Sprintf("Dispatched LLM deliberation request to worker (mode: %s)", mode))
	resp, err := d.cli.Think(ctx, &agentfeedpb.ThinkRequest{
		Agent:    toProtoAgent(agent),
		Post:     toProtoPost(post, agentMap),
		Recent:   protoRecent,
		Mode:     mode,
		Memories: memories,
	})
	if err != nil {
		logs.AgentError(agent.ID, agent.Name, agent.Handle, "ai-worker", fmt.Sprintf("AI Worker gRPC error: %v", err))
		return nil, fmt.Errorf("worker think: %w", err)
	}
	logs.AgentInfo(agent.ID, agent.Name, agent.Handle, "ai-worker", fmt.Sprintf("Received AI worker decision: action=%s", resp.Action))
	return &Decision{Action: resp.Action, Text: resp.Text}, nil
}

func toProtoAgent(a model.Agent) *agentfeedpb.Agent {
	return &agentfeedpb.Agent{
		Id:           a.ID,
		Handle:       a.Handle,
		Name:         a.Name,
		Persona:      a.Persona,
		Goals:        a.Goals,
		Policies:     a.Policies,
		Interests:    a.Interests,
		AvatarUrl:    a.AvatarURL,
		Active:       a.Active,
		Description:  a.Description,
		Instructions: a.Instructions,
	}
}

func toProtoPost(p model.Post, agentMap map[int64]model.Agent) *agentfeedpb.Post {
	parent := int64(0)
	if p.ParentID != nil {
		parent = *p.ParentID
	}
	root := int64(0)
	if p.RootID != nil {
		root = *p.RootID
	}
	authorHandle := ""
	authorName := ""
	if a, ok := agentMap[p.AgentID]; ok {
		authorHandle = a.Handle
		authorName = a.Name
	}
	return &agentfeedpb.Post{
		Id:           p.ID,
		AgentId:      p.AgentID,
		Body:         p.Body,
		Kind:         p.Kind,
		ParentId:     parent,
		RootId:       root,
		Mentions:     p.Mentions,
		AuthorHandle: authorHandle,
		AuthorName:   authorName,
	}
}
