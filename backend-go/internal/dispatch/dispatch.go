package dispatch

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	agentfeedpb "github.com/agent-feed/backend-go/gen/agentfeedpb"
	"github.com/agent-feed/backend-go/internal/db/agents"
	"github.com/agent-feed/backend-go/internal/db/posts"
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
		grpc.WithBlock(),
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
// so replies feel aware of the conversation. Returns a structured decision.
func (d *Dispatcher) Think(ctx context.Context, agent agents.Agent, post posts.Post, recent []posts.Post) (*Decision, error) {
	return d.think(ctx, "respond", agent, post, recent)
}

// Proactive asks the worker to author fresh, original content from the agent.
// It is the "empty feed" path: with no post to react to, the agent posts on its own.
func (d *Dispatcher) Proactive(ctx context.Context, agent agents.Agent, recent []posts.Post) (*Decision, error) {
	return d.think(ctx, "proactive", agent, posts.Post{}, recent)
}

func (d *Dispatcher) think(ctx context.Context, mode string, agent agents.Agent, post posts.Post, recent []posts.Post) (*Decision, error) {
	protoRecent := make([]*agentfeedpb.Post, 0, len(recent))
	for i := range recent {
		protoRecent = append(protoRecent, toProtoPost(recent[i]))
	}
	resp, err := d.cli.Think(ctx, &agentfeedpb.ThinkRequest{
		Agent:  toProtoAgent(agent),
		Post:   toProtoPost(post),
		Recent: protoRecent,
		Mode:   mode,
	})
	if err != nil {
		return nil, fmt.Errorf("worker think: %w", err)
	}
	return &Decision{Action: resp.Action, Text: resp.Text}, nil
}

func toProtoAgent(a agents.Agent) *agentfeedpb.Agent {
	return &agentfeedpb.Agent{
		Id:        a.ID,
		Handle:    a.Handle,
		Name:      a.Name,
		Persona:   a.Persona,
		Goals:     a.Goals,
		Policies:  a.Policies,
		Interests: a.Interests,
		AvatarUrl: a.AvatarURL,
		Active:    a.Active,
	}
}

func toProtoPost(p posts.Post) *agentfeedpb.Post {
	parent := int64(0)
	if p.ParentID != nil {
		parent = *p.ParentID
	}
	root := int64(0)
	if p.RootID != nil {
		root = *p.RootID
	}
	return &agentfeedpb.Post{
		Id:       p.ID,
		AgentId:  p.AgentID,
		Body:     p.Body,
		Kind:     p.Kind,
		ParentId: parent,
		RootId:   root,
		Mentions: p.Mentions,
	}
}
