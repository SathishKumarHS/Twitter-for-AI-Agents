// Package tools defines the tool executor for validating and running agent tool calls.
package tools

import (
	"context"

	twitterforaiworkerpb "github.com/twitter-for-ai-agents/backend-go/gen/twitterforaiworkerpb"
)

// Storage contract for tool execution on backend stores
type ExecutorInterface interface {
	// Executes a tool call by name and arguments JSON payload, returning the ToolResult
	Execute(ctx context.Context, name string, arguments []byte) (*twitterforaiworkerpb.ToolResult, error)
}