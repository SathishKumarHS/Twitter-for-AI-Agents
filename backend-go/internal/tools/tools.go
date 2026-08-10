// Package tools implements the tool executor: the authoritative place where an
// agent's tool calls are validated and executed against the real stores.
// Tool names MUST match ai-worker/internal/prompts/prompts.py build_tools() and
// the whitelist in ai-worker/internal/llm/llm.py ALLOWED_TOOLS.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	twitterforaiworkerpb "github.com/twitter-for-ai-agents/backend-go/gen/twitterforaiworkerpb"
	lastseen "github.com/twitter-for-ai-agents/backend-go/internal/db/last_seen"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/memories"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/posts"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/subscriptions"
	"github.com/twitter-for-ai-agents/backend-go/internal/events"
	"github.com/twitter-for-ai-agents/backend-go/internal/redis"
	"github.com/twitter-for-ai-agents/backend-go/internal/scheduler/notifier"
)

type Executor struct {
	posts           posts.PostInterface
	memories        memories.MemoriesInterface
	lastSeen        lastseen.LastSeenInterface
	subscriptions   subscriptions.SubscriptionInterface
	hub             events.HubInterface
	feedCache       redis.RedisInterface
	notifier        notifier.NotifierInterface
	agent           model.Agent
	agentMap        map[int64]model.Agent
	feedDefaultSize int
	feedMaxSize     int
	mode            string
}

func New(
	posts posts.PostInterface,
	memories memories.MemoriesInterface,
	lastSeen lastseen.LastSeenInterface,
	subscriptions subscriptions.SubscriptionInterface,
	hub events.HubInterface,
	feedCache redis.RedisInterface,
	notifier notifier.NotifierInterface,
	agent model.Agent,
	agentMap map[int64]model.Agent,
	feedDefaultSize int,
	feedMaxSize int,
	mode string,
) ExecutorInterface {
	return &Executor{
		posts:           posts,
		memories:        memories,
		lastSeen:        lastSeen,
		subscriptions:   subscriptions,
		hub:             hub,
		feedCache:       feedCache,
		notifier:        notifier,
		agent:           agent,
		agentMap:        agentMap,
		feedDefaultSize: feedDefaultSize,
		feedMaxSize:     feedMaxSize,
		mode:            mode,
	}
}

func (e *Executor) Execute(ctx context.Context, name string, arguments []byte) (*twitterforaiworkerpb.ToolResult, error) {
	if _, ok := allowedTools[allowedTool(name)]; !ok {
		return &twitterforaiworkerpb.ToolResult{Name: name, Output: fmt.Sprintf("unknown tool: %s", name), Error: true, Acted: false}, nil
	}

	switch allowedTool(name) {
	case TOOL_GET_FEED:
		return e.getFeed(ctx, arguments)
	case TOOL_GET_THREAD:
		return e.getThread(ctx, arguments)
	case TOOL_GET_MEMORIES:
		return e.getMemories(ctx)
	case TOOL_POST:
		return e.post(ctx, arguments)
	case TOOL_REPLY:
		return e.reply(ctx, arguments)
	default:
		return &twitterforaiworkerpb.ToolResult{Name: name, Output: fmt.Sprintf("unknown tool: %s", name), Error: true, Acted: false}, nil
	}
}


type postView struct {
	ID       int64  `json:"id"`
	AgentID  int64  `json:"agent_id"`
	Handle   string `json:"handle"`
	Name     string `json:"name"`
	Body     string `json:"body"`
	Kind     string `json:"kind"`
	ParentID int64  `json:"parent_id"`
	RootID   int64  `json:"root_id"`
}

func (e *Executor) view(p model.Post) postView {
	v := postView{
		ID:      p.ID,
		AgentID: p.AgentID,
		Body:    p.Body,
		Kind:    p.Kind,
	}
	if p.ParentID != nil {
		v.ParentID = *p.ParentID
	}
	if p.RootID != nil {
		v.RootID = *p.RootID
	}
	if a, ok := e.agentMap[p.AgentID]; ok {
		v.Handle = a.Handle
		v.Name = a.Name
	}
	return v
}

func (e *Executor) getFeed(ctx context.Context, arguments []byte) (*twitterforaiworkerpb.ToolResult, error) {
	var args struct {
		Limit *int `json:"limit"`
	}
	_ = json.Unmarshal(arguments, &args)
	limit := e.feedDefaultSize
	if args.Limit != nil {
		limit = *args.Limit
	}
	if limit < 1 {
		limit = 1
	}
	if limit > e.feedMaxSize {
		limit = e.feedMaxSize
	}
	posts, err := e.posts.Recent(ctx, limit)
	if err != nil {
		return &twitterforaiworkerpb.ToolResult{Name: "get_feed", Output: fmt.Sprintf("error: %v", err), Error: true, Acted: false}, nil
	}
	out := make([]postView, 0, len(posts))
	for i := range posts {
		out = append(out, e.view(posts[i]))
	}
	return jsonResult(TOOL_GET_FEED, out)
}

func (e *Executor) getThread(ctx context.Context, arguments []byte) (*twitterforaiworkerpb.ToolResult, error) {
	var args struct {
		RootID int64 `json:"root_id"`
	}
	if err := json.Unmarshal(arguments, &args); err != nil || args.RootID <= 0 {
		return &twitterforaiworkerpb.ToolResult{Name: TOOL_GET_THREAD.String(), Output: "error: root_id is required and must be positive", Error: true, Acted: false}, nil
	}
	posts, err := e.posts.Thread(ctx, args.RootID)
	if err != nil {
		return &twitterforaiworkerpb.ToolResult{Name: TOOL_GET_THREAD.String(), Output: fmt.Sprintf("error: %v", err), Error: true, Acted: false}, nil
	}
	out := make([]postView, 0, len(posts))
	for i := range posts {
		out = append(out, e.view(posts[i]))
	}
	return jsonResult(TOOL_GET_THREAD, out)
}

func (e *Executor) getMemories(ctx context.Context) (*twitterforaiworkerpb.ToolResult, error) {
	memories, err := e.memories.List(ctx, e.agent.ID, 10)
	if err != nil {
		return &twitterforaiworkerpb.ToolResult{Name: TOOL_GET_MEMORIES.String(), Output: fmt.Sprintf("error: %v", err), Error: true, Acted: false}, nil
	}
	out := make([]string, 0, len(memories))
	for i := range memories {
		out = append(out, memories[i].Summary)
	}
	return jsonResult(TOOL_GET_MEMORIES, out)
}

func (e *Executor) post(ctx context.Context, arguments []byte) (*twitterforaiworkerpb.ToolResult, error) {
	if e.mode == "respond" {
		return &twitterforaiworkerpb.ToolResult{Name: TOOL_POST.String(), Output: "error: post is not allowed while replying; use reply(parent_id, text) instead", Error: true, Acted: false}, nil
	}
	var args struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(arguments, &args); err != nil {
		return &twitterforaiworkerpb.ToolResult{Name: TOOL_POST.String(), Output: "error: invalid arguments", Error: true, Acted: false}, nil
	}
	text := cleanText(args.Text)
	if text == "" {
		return &twitterforaiworkerpb.ToolResult{Name: TOOL_POST.String(), Output: "error: text is required and must be non-empty", Error: true, Acted: false}, nil
	}
	created, err := e.createPost(ctx, text, 0)
	if err != nil {
		return &twitterforaiworkerpb.ToolResult{Name: TOOL_POST.String(), Output: fmt.Sprintf("error: %v", err), Error: true, Acted: false}, nil
	}
	return &twitterforaiworkerpb.ToolResult{
		Name:   TOOL_POST.String(),
		Output: fmt.Sprintf(`{"status": "created", "id": %d, "kind": "post"}`, created.ID),
		Acted:  true,
	}, nil
}

func (e *Executor) reply(ctx context.Context, arguments []byte) (*twitterforaiworkerpb.ToolResult, error) {
	if e.mode == "proactive" {
		return &twitterforaiworkerpb.ToolResult{Name: TOOL_REPLY.String(), Output: "error: reply is not allowed while authoring a new post; use post(text) instead", Error: true, Acted: false}, nil
	}
	var args struct {
		ParentID int64  `json:"parent_id"`
		Text     string `json:"text"`
	}
	if err := json.Unmarshal(arguments, &args); err != nil {
		return &twitterforaiworkerpb.ToolResult{Name: TOOL_REPLY.String(), Output: "error: invalid arguments", Error: true, Acted: false}, nil
	}
	text := cleanText(args.Text)
	if args.ParentID <= 0 {
		return &twitterforaiworkerpb.ToolResult{Name: TOOL_REPLY.String(), Output: "error: parent_id is required and must be positive", Error: true, Acted: false}, nil
	}
	if text == "" {
		return &twitterforaiworkerpb.ToolResult{Name: TOOL_REPLY.String(), Output: "error: text is required and must be non-empty", Error: true, Acted: false}, nil
	}
	created, err := e.createPost(ctx, text, args.ParentID)
	if err != nil {
		return &twitterforaiworkerpb.ToolResult{Name: TOOL_REPLY.String(), Output: fmt.Sprintf("error: %v", err), Error: true, Acted: false}, nil
	}
	return &twitterforaiworkerpb.ToolResult{
		Name:   TOOL_REPLY.String(),
		Output: fmt.Sprintf(`{"status": "created", "id": %d, "kind": "reply", "parent_id": %d}`, created.ID, args.ParentID),
		Acted:  true,
	}, nil
}

func (e *Executor) createPost(ctx context.Context, text string, parentID int64) (model.Post, error) {
	created, err := e.posts.Create(ctx, e.agent.ID, text, parentID)
	if err != nil {
		return model.Post{}, err
	}
	kind := TOOL_POST.String()
	if parentID == 0 {
		if e.lastSeen != nil {
			_ = e.lastSeen.SetSeen(ctx, e.agent.ID, created.ID)
		}
		if e.subscriptions != nil {
			_ = e.subscriptions.Subscribe(ctx, e.agent.ID, created.ID)
		}
	} else {
		kind = TOOL_REPLY.String()
		if created.RootID != nil && e.subscriptions != nil {
			_ = e.subscriptions.Subscribe(ctx, e.agent.ID, *created.RootID)
		}
	}
	if e.hub != nil {
		e.hub.Publish(events.Event{Type: events.EventPostNew, Data: map[string]any{
			"id": created.ID, "agent_id": e.agent.ID, "body": text, "kind": kind, "parent_id": parentID,
		}})
	}
	if e.feedCache != nil {
		_ = e.feedCache.InvalidateFeed(ctx)
	}
	if e.memories != nil {
		if parentID == 0 {
			_ = e.memories.Add(ctx, e.agent.ID, "learning", "posted: "+truncate(text, 80))
		} else {
			_ = e.memories.Add(ctx, e.agent.ID, "learning", "replied to post #"+itoa(parentID))
		}
	}
	if e.notifier != nil {
		go e.notifier.NotifySubscribers(ctx, created, e.agentMap)
	}
	return created, nil
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}

func itoa(i int64) string {
	return fmt.Sprintf("%d", i)
}

func jsonResult(name allowedTool, v any) (*twitterforaiworkerpb.ToolResult, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return &twitterforaiworkerpb.ToolResult{Name: name.String(), Output: fmt.Sprintf("error: %v", err), Error: true, Acted: false}, nil
	}
	return &twitterforaiworkerpb.ToolResult{Name: name.String(), Output: string(b), Error: false, Acted: false}, nil
}

func cleanText(s string) string {
	s = strings.TrimSpace(s)
	for _, marker := range []string{"...", "…", "---"} {
		s = strings.TrimPrefix(s, marker)
		s = strings.TrimSpace(s)
	}
	r := []rune(s)
	if len(r) > 280 {
		r = r[:280]
	}
	return strings.TrimSpace(string(r))
}
