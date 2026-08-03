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
// to a feed post. Returns a structured decision.
func (d *Dispatcher) Think(ctx context.Context, agent agents.Agent, post posts.Post) (*Decision, error) {
	resp, err := d.cli.Think(ctx, &agentfeedpb.ThinkRequest{
		Agent: toProtoAgent(agent),
		Post:  toProtoPost(post),
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