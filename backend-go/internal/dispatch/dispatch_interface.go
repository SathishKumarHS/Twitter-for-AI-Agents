// Package dispatch manages sending agent turns asynchronously to the Python AI Worker via gRPC.
package dispatch

import (
	"context"

	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
)

// Typed deliberation mode for a single agent LLM turn ("respond" vs "proactive")
type Mode string

const (
	ModeRespond   Mode = "respond"
	ModeProactive Mode = "proactive"
)

// Returns the string representation of the mode
func (m Mode) String() string {
	return string(m)
}

// Holds the decision outcome returned after an agent deliberation turn completes
type Decision struct {
	Executed bool // True if a post or reply was published on the backend during the turn
}

// Callback handler invoked asynchronously when the AI Worker completes a Think RPC call
type DecisionHandler func(decision *Decision, err error)

// Storage and RPC contract for dispatching agent turns to the AI Worker
type DispatcherInterface interface {
	// Dispatches an agent turn asynchronously to the AI Worker and invokes handler on completion
	Act(ctx context.Context, mode Mode, agent model.Agent, post model.Post, handler DecisionHandler)
	// Closes the underlying gRPC client connection
	Close() error
}
