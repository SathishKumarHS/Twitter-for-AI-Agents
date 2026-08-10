package tools

import (
	"context"

	twitterforaiworkerpb "github.com/twitter-for-ai-agents/backend-go/gen/twitterforaiworkerpb"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/agents"
	lastseen "github.com/twitter-for-ai-agents/backend-go/internal/db/last_seen"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/memories"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/posts"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/subscriptions"
	"github.com/twitter-for-ai-agents/backend-go/internal/events"
	"github.com/twitter-for-ai-agents/backend-go/internal/redis"
	"github.com/twitter-for-ai-agents/backend-go/internal/scheduler/notifier"
)

// gRPC server implementation for the ToolExecutor service
type toolServer struct {
	twitterforaiworkerpb.UnimplementedToolExecutorServer
	agents          agents.AgentsInterface
	posts           posts.PostInterface
	memories        memories.MemoriesInterface
	lastSeen        lastseen.LastSeenInterface
	subscriptions   subscriptions.SubscriptionInterface
	hub             events.HubInterface
	feedCache       redis.RedisInterface
	notifier        notifier.NotifierInterface
	feedDefaultSize int
	feedMaxSize     int
}

// Instantiates a new ToolExecutorServer for handling gRPC tool execution requests from the AI Worker
func NewToolGRPCServer(
	agents agents.AgentsInterface,
	posts posts.PostInterface,
	memories memories.MemoriesInterface,
	lastSeen lastseen.LastSeenInterface,
	subscriptions subscriptions.SubscriptionInterface,
	hub events.HubInterface,
	feedCache redis.RedisInterface,
	notifier notifier.NotifierInterface,
	feedDefaultSize, feedMaxSize int,
) twitterforaiworkerpb.ToolExecutorServer {
	return &toolServer{
		agents:          agents,
		posts:           posts,
		memories:        memories,
		lastSeen:        lastSeen,
		subscriptions:   subscriptions,
		hub:             hub,
		feedCache:       feedCache,
		notifier:        notifier,
		feedDefaultSize: feedDefaultSize,
		feedMaxSize:     feedMaxSize,
	}
}

// Handles gRPC ExecuteTool RPC requests from the AI Worker, validating the agent and invoking the Executor
func (t *toolServer) ExecuteTool(ctx context.Context, req *twitterforaiworkerpb.ToolCall) (*twitterforaiworkerpb.ToolResult, error) {
	agentPtr, err := t.agents.Get(ctx, req.AgentId)
	if err != nil || agentPtr == nil || !agentPtr.Active {
		return &twitterforaiworkerpb.ToolResult{
			Name:   req.Name,
			Output: "unknown or inactive agent",
			Error:  true,
		}, nil
	}
	agent := *agentPtr

	agentMap := make(map[int64]model.Agent, 1)
	agentMap[agent.ID] = agent
	if active, aerr := t.agents.ListActive(ctx); aerr == nil {
		for _, a := range active {
			agentMap[a.ID] = a
		}
	}

	ex := New(
		t.posts,
		t.memories,
		t.lastSeen,
		t.subscriptions,
		t.hub,
		t.feedCache,
		t.notifier,
		agent,
		agentMap,
		t.feedDefaultSize,
		t.feedMaxSize,
		req.Mode,
	)
	return ex.Execute(ctx, req.Name, []byte(req.Arguments))
}
