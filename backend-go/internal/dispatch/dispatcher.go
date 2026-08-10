package dispatch

import (
	"context"
	"fmt"
	"log/slog"

	twitterforaiworkerpb "github.com/twitter-for-ai-agents/backend-go/gen/twitterforaiworkerpb"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
)

// Asynchronously dispatches an agent turn to the AI Worker via gRPC Think RPC
func (d *Dispatcher) Act(ctx context.Context, mode Mode, agent model.Agent, post model.Post, handler DecisionHandler) {
	slog.Info("Turn dispatched to worker", "mode", mode, "agent_id", agent.ID, "agent_name", agent.Name, "agent_handle", agent.Handle, "post_id", post.ID)
	go func() {
		resp, err := d.cli.Think(ctx, &twitterforaiworkerpb.ThinkRequest{
			Agent: toProtoAgent(agent),
			Post:  toProtoPost(post),
			Mode:  string(mode),
		})
		if err != nil {
			slog.Error("AI Worker gRPC error", "error", err, "agent_id", agent.ID, "agent_name", agent.Name, "agent_handle", agent.Handle)
			if handler != nil {
				handler(nil, fmt.Errorf("worker think: %w", err))
			}
			return
		}
		if handler != nil {
			handler(&Decision{
				Executed: resp.Executed,
			}, nil)
		}
	}()
}
