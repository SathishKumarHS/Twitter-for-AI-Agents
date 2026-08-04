package sched

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/agent-feed/backend-go/internal/db"
	"github.com/agent-feed/backend-go/internal/db/model"
	"github.com/agent-feed/backend-go/internal/dispatch"
	"github.com/agent-feed/backend-go/internal/events"
	"github.com/agent-feed/backend-go/internal/logs"
	"github.com/agent-feed/backend-go/internal/redis"
)

// turn is one unit of work for the worker pool: a single agent deliberation.
type turn struct {
	mode       turnMode
	agent      model.Agent
	agentMap   map[int64]model.Agent
	targetPost *model.Post
}

// turnMode selects what a turn does for an agent.
type turnMode string

const (
	modePost    turnMode = "author_main_post"
	modeReview  turnMode = "review_unread_feed"
	modeRecheck turnMode = "comment_recheck"
)

const lockTTLMargin = 60 * time.Second

// Scheduler orchestrates periodic agent deliberations and subscriber notifications.
type Scheduler struct {
	interval      time.Duration
	postCooldown  time.Duration
	turnTimeout   time.Duration
	maxConcurrent int
	queue         chan turn
	wg            sync.WaitGroup
	locks         redis.RedisInterface
	db            *db.DB
	dispatcher    dispatch.DispatcherInterface
	hub           events.HubInterface
	cache         redis.RedisInterface
	rootCtx       context.Context
	cancel        context.CancelFunc
	done          chan struct{}
}

// New builds a scheduler over the domain stores and event hub.
func New(interval, postCooldown, turnTimeout time.Duration, maxConcurrent int, locks redis.RedisInterface, db *db.DB, d dispatch.DispatcherInterface, hub events.HubInterface, c redis.RedisInterface) *Scheduler {
	return &Scheduler{
		interval:      interval,
		postCooldown:  postCooldown,
		turnTimeout:   turnTimeout,
		maxConcurrent: maxConcurrent,
		queue:         make(chan turn, maxConcurrent*2),
		locks:         locks,
		db:            db,
		dispatcher:    d,
		hub:           hub,
		cache:         c,
		done:          make(chan struct{}),
	}
}

func (s *Scheduler) invalidateFeedCache(ctx context.Context) {
	if s.cache != nil {
		_ = s.cache.InvalidateFeed(ctx)
	}
}

// Start launches the worker pool and the background review ticker loop.
func (s *Scheduler) Start(parent context.Context) {
	ctx, cancel := context.WithCancel(parent)
	s.rootCtx = ctx
	s.cancel = cancel
	for i := 0; i < s.maxConcurrent; i++ {
		s.wg.Add(1)
		go s.worker(ctx)
	}
	go s.run(ctx)
	msg := fmt.Sprintf("Scheduler started (review interval=%s, agent post cooldown=%s, workers=%d)", s.interval, s.postCooldown, s.maxConcurrent)
	log.Println(msg)
	logs.Info("system", msg)
}

// Stop terminates the scheduler loop and waits for in-flight turns to finish.
func (s *Scheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	<-s.done
	s.wg.Wait()
}

func (s *Scheduler) run(ctx context.Context) {
	defer close(s.done)
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Fire an immediate review on startup so the feed goes live without waiting
	// a full interval.
	s.tick(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("scheduler stopping")
			logs.Info("system", "Scheduler stopping")
			return
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

// worker is one consumer of the turn queue. Turns are enqueued (never dropped)
// and each is executed under the per-agent lock with a hard turn timeout.
func (s *Scheduler) worker(ctx context.Context) {
	defer s.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case t := <-s.queue:
			turnCtx, cancel := context.WithTimeout(ctx, s.turnTimeout)
			s.evaluateAgentTurn(turnCtx, t)
			cancel()
		}
	}
}

// tick runs one full 10-minute cycle in two phases:
//  1. Every agent authors a new main post (cooldown-gated).
//  2. Every agent reviews all new posts since its watermark and replies where
//     the AI decides to, so a feed of N agents yields N posts then up to N×N
//     review decisions.
//
// Both phases are enqueued on the same FIFO queue, so all post turns complete
// before the review turns start (workers pull posts first).
func (s *Scheduler) tick(ctx context.Context) {
	list, err := s.db.AgentsDB.ListActive(ctx)
	if err != nil || len(list) == 0 {
		if err != nil {
			log.Printf("scheduler: list active agents: %v", err)
		}
		return
	}

	agentMap := make(map[int64]model.Agent, len(list))
	for _, a := range list {
		agentMap[a.ID] = a
	}

	logs.Info("system", fmt.Sprintf("Scheduler review cycle started for %d active agents", len(list)))

	// Phase 1: main posts.
	for _, agent := range list {
		s.submitTurn(ctx, modePost, agent, agentMap, nil)
	}
	// Phase 2: review everything new.
	for _, agent := range list {
		s.submitTurn(ctx, modeReview, agent, agentMap, nil)
	}
}

// submitTurn enqueues one agent turn onto the queue. It never drops work: if
// the pool is saturated the turn simply waits for a free worker.
func (s *Scheduler) submitTurn(ctx context.Context, mode turnMode, agent model.Agent, agentMap map[int64]model.Agent, targetPost *model.Post) {
	if !agent.Active {
		return
	}
	select {
	case s.queue <- turn{mode: mode, agent: agent, agentMap: agentMap, targetPost: targetPost}:
	case <-ctx.Done():
	}
}

func (m turnMode) String() string {
	return string(m)
}

// evaluateAgentTurn is the unified deliberation pipeline for an AI agent,
// dispatched by turn mode:
//   - modePost: author a new main post when the 10-min cooldown has expired.
//   - modeReview: review every post since the watermark (reply or ignore each).
//   - modeRecheck: review a specific post/thread (comment-triggered).
//
// Guardrails:
//  1. Hard 10-minute cooldown on MAIN POSTS (at most 1 new thread per agent in
//     10 mins). Replies are not cooldown-gated: they are the burst of comments
//     that follow the posting phase.
//  2. The agent reads everything published since its last review (watermark).
//  3. Complete AI decision autonomy (AI determines post, reply, or ignore).
func (s *Scheduler) isAgentActive(ctx context.Context, id int64) bool {
	if s.cache != nil {
		if active, found := s.cache.IsAgentActive(ctx, id); found {
			return active
		}
	}
	latest, err := s.db.AgentsDB.Get(ctx, id)
	if err != nil {
		return false
	}
	if s.cache != nil {
		_ = s.cache.SetAgentActive(ctx, id, latest.Active)
	}
	return latest.Active
}

func (s *Scheduler) evaluateAgentTurn(ctx context.Context, t turn) {
	// Re-check active status via Redis/DB cache in case agent was paused while waiting in queue.
	if !s.isAgentActive(ctx, t.agent.ID) {
		logs.AgentInfo(t.agent.ID, t.agent.Name, t.agent.Handle, "turn", fmt.Sprintf("Turn skipped: agent @%s is currently paused", t.agent.Handle))
		return
	}

	// Serialize per agent (fleet-wide) so the cooldown check cannot race.
	release, acquired, err := s.locks.Acquire(ctx, t.agent.ID, s.turnTimeout+lockTTLMargin)
	if err != nil {
		log.Printf("scheduler: agent %d acquire turn lock: %v", t.agent.ID, err)
		return
	}
	if !acquired {
		log.Printf("scheduler: agent %d turn already in progress elsewhere, skipping", t.agent.ID)
		logs.AgentWarn(t.agent.ID, t.agent.Name, t.agent.Handle, "turn", fmt.Sprintf("Turn skipped: agent @%s already in progress elsewhere", t.agent.Handle))
		return
	}
	defer release()

	logs.AgentInfo(t.agent.ID, t.agent.Name, t.agent.Handle, "turn", fmt.Sprintf("Worker starting turn for @%s (mode: %s)", t.agent.Handle, t.mode.String()))

	switch t.mode {
	case modePost:
		s.authorMainPost(ctx, t.agent, t.agentMap)
	case modeReview:
		s.reviewUnread(ctx, t.agent, t.agentMap)
	case modeRecheck:
		if t.targetPost != nil {
			s.reviewTargetPost(ctx, t.agent, *t.targetPost, t.agentMap)
		}
	}
}

// authorMainPost runs the "post" phase: the agent authors a fresh main post,
// gated by the 10-minute main-post cooldown so each agent posts at most once
// per cycle. It never blocks on a full pool and never errors out the cycle.
func (s *Scheduler) authorMainPost(ctx context.Context, agent model.Agent, agentMap map[int64]model.Agent) {
	// Guardrail 1: at most 1 main post per 10 minutes.
	last, err := s.db.PostDB.LastMainPost(ctx, agent.ID)
	if err == nil && !last.IsZero() && time.Since(last) < s.postCooldown {
		return // already authored a main post recently
	}

	recent, rerr := s.db.PostDB.Recent(ctx, 10)
	if rerr != nil {
		log.Printf("scheduler: agent %d fetch recent posts: %v", agent.ID, rerr)
		return
	}
	decision, err := s.dispatcher.Proactive(ctx, agent, recent, s.memories(ctx, agent.ID), agentMap)
	if err != nil {
		log.Printf("scheduler: agent %d proactive deliberation failed: %v", agent.ID, err)
		return
	}

	text := cleanText(decision.Text)
	if decision.Action != dispatch.ActionPost || text == "" {
		return
	}

	created, cerr := s.db.PostDB.Create(ctx, agent.ID, text, 0)
	if cerr != nil {
		log.Printf("scheduler: agent %d persist post: %v", agent.ID, cerr)
		return
	}
	_ = s.db.LastSeenDB.SetSeen(ctx, agent.ID, created.ID)
	// The author subscribes to their own thread so they hear the replies.
	_ = s.db.SubscriptionsDB.Subscribe(ctx, agent.ID, created.ID)
	s.hub.Publish(events.Event{Type: events.EventPostNew, Data: map[string]any{
		"id": created.ID, "agent_id": agent.ID, "body": text, "kind": "post",
	}})
	s.invalidateFeedCache(ctx)
	_ = s.db.MemoriesDB.Add(ctx, agent.ID, "learning", "posted: "+truncate(text, 80))
	log.Printf("scheduler: agent %d (@%s) posted #%d", agent.ID, agent.Handle, created.ID)
	logs.AgentInfo(agent.ID, agent.Name, agent.Handle, "post", "Published main post #"+strconv.FormatInt(created.ID, 10)+": "+truncate(text, 80))
	go s.notifySubscribers(s.rootCtx, created, agentMap)
}

// reviewUnread runs the periodic "review" phase. It reads all new posts published
// since the agent's watermark and evaluates each post for a potential reply.
func (s *Scheduler) reviewUnread(ctx context.Context, agent model.Agent, agentMap map[int64]model.Agent) {
	newPosts, err := s.db.PostDB.NewPostsAfter(ctx, agent.ID)
	if err != nil {
		log.Printf("scheduler: agent %d fetch new posts: %v", agent.ID, err)
		return
	}

	seenMax := s.db.LastSeenDB.LastSeen(ctx, agent.ID)

	for _, p := range newPosts {
		if p.AgentID == agent.ID {
			continue
		}
		s.maybeReply(ctx, agent, p, agentMap)
		if p.ID > seenMax {
			seenMax = p.ID
		}
		// Advance the watermark incrementally so a long review that times out
		// still records progress and never re-reviews already-decided posts.
		if seenMax > s.db.LastSeenDB.LastSeen(ctx, agent.ID) {
			if err := s.db.LastSeenDB.SetSeen(ctx, agent.ID, seenMax); err != nil {
				log.Printf("scheduler: agent %d set seen: %v", agent.ID, err)
			}
		}
	}
}

// reviewTargetPost runs a comment-triggered recheck for a specific post.
func (s *Scheduler) reviewTargetPost(ctx context.Context, agent model.Agent, targetPost model.Post, agentMap map[int64]model.Agent) {
	if targetPost.AgentID == agent.ID {
		return
	}
	s.maybeReply(ctx, agent, targetPost, agentMap)
	seenMax := s.db.LastSeenDB.LastSeen(ctx, agent.ID)
	if targetPost.ID > seenMax {
		if err := s.db.LastSeenDB.SetSeen(ctx, agent.ID, targetPost.ID); err != nil {
			log.Printf("scheduler: agent %d set seen for target post: %v", agent.ID, err)
		}
	}
}

// maybeReply asks the AI how to react to a single post and, if it chooses to
// reply, persists the comment and notifies the thread's subscribers.
func (s *Scheduler) maybeReply(ctx context.Context, agent model.Agent, post model.Post, agentMap map[int64]model.Agent) {
	thread := s.threadContext(ctx, post)
	decision, err := s.dispatcher.Think(ctx, agent, post, thread, s.memories(ctx, agent.ID), agentMap)
	if err != nil {
		log.Printf("scheduler: agent %d deliberation on post %d failed: %v", agent.ID, post.ID, err)
		logs.AgentError(agent.ID, agent.Name, agent.Handle, "review", "Deliberation failed on post #"+strconv.FormatInt(post.ID, 10)+": "+err.Error())
		return
	}

	text := cleanText(decision.Text)
	if decision.Action != dispatch.ActionReply || text == "" {
		targetHandle := ""
		if author, ok := agentMap[post.AgentID]; ok {
			targetHandle = author.Handle
		}
		logs.AgentInfo(agent.ID, agent.Name, agent.Handle, "ignore", "Evaluated post #"+strconv.FormatInt(post.ID, 10)+" by @"+targetHandle+" — AI decided to ignore")
		return
	}

	created, cerr := s.db.PostDB.Create(ctx, agent.ID, text, post.ID)
	if cerr != nil {
		log.Printf("scheduler: agent %d persist reply: %v", agent.ID, cerr)
		return
	}
	if created.RootID != nil {
		_ = s.db.SubscriptionsDB.Subscribe(ctx, agent.ID, *created.RootID)
	}
	s.hub.Publish(events.Event{Type: events.EventPostNew, Data: map[string]any{
		"id": created.ID, "agent_id": agent.ID, "body": text, "kind": "reply", "parent_id": post.ID,
	}})
	s.invalidateFeedCache(ctx)
	_ = s.db.MemoriesDB.Add(ctx, agent.ID, "learning", "replied to post #"+itoa(post.ID))
	log.Printf("scheduler: agent %d (@%s) replied to post #%d with reply #%d", agent.ID, agent.Handle, post.ID, created.ID)
	logs.AgentInfo(agent.ID, agent.Name, agent.Handle, "reply", "Replied to post #"+strconv.FormatInt(post.ID, 10)+" with comment #"+strconv.FormatInt(created.ID, 10)+": "+truncate(text, 80))
	go s.notifySubscribers(s.rootCtx, created, agentMap)
}

// notifySubscribers alerts the right audience about a new post or comment:
//   - a new main post notifies the author's followers (subscribers) plus mentions;
//   - a thread comment notifies everyone subscribed to the thread plus mentions;
//
// and never re-notifies the agent that just authored it.
func (s *Scheduler) notifySubscribers(ctx context.Context, created model.Post, agentMap map[int64]model.Agent) {
	var subscriberIDs, followerIDs []int64
	if created.ParentID != nil {
		// Thread comment: notify thread subscribers.
		if created.RootID != nil {
			if subs, err := s.db.SubscriptionsDB.ListSubscribers(ctx, *created.RootID); err == nil {
				subscriberIDs = subs
			}
		}
	} else {
		// New main post: broadcast notification turn to all network agents.
		for id := range agentMap {
			if id != created.AgentID {
				followerIDs = append(followerIDs, id)
			}
		}
	}

	// Mentions are always notified and subscribed to the thread.
	rootID := created.RootID
	if rootID == nil && created.ParentID == nil {
		rootID = &created.ID
	}
	for _, mentionID := range created.Mentions {
		if mentionID != created.AgentID && rootID != nil {
			_ = s.db.SubscriptionsDB.Subscribe(ctx, mentionID, *rootID)
		}
	}

	for _, targetID := range selectNotifyTargets(created, subscriberIDs, followerIDs) {
		if targetAgent, ok := agentMap[targetID]; ok {
			if s.isAgentActive(ctx, targetID) {
				p := created
				s.submitTurn(ctx, modeRecheck, targetAgent, agentMap, &p)
			}
		}
	}
}

// selectNotifyTargets computes the deduped, author-excluded audience for a new
// post. A reply notifies thread subscribers; a main post notifies the author's
// followers. Mentioned agents are always included.
func selectNotifyTargets(created model.Post, subscriberIDs, followerIDs []int64) []int64 {
	var base []int64
	if created.ParentID != nil {
		base = subscriberIDs
	} else {
		base = followerIDs
	}
	targets := make([]int64, 0, len(base)+len(created.Mentions)+1)
	targets = append(targets, base...)
	targets = append(targets, created.Mentions...)
	targets = dedupeInts(targets)

	out := targets[:0]
	for _, t := range targets {
		if t != created.AgentID {
			out = append(out, t)
		}
	}
	return out
}

func (s *Scheduler) memories(ctx context.Context, agentID int64) []string {
	list, err := s.db.MemoriesDB.List(ctx, agentID, 10)
	if err != nil || len(list) == 0 {
		return nil
	}
	out := make([]string, 0, len(list))
	for i := range list {
		out = append(out, list[i].Summary)
	}
	return out
}

func (s *Scheduler) threadContext(ctx context.Context, post model.Post) []model.Post {
	root := post.RootID
	if root == nil && post.ParentID != nil {
		root = post.ParentID
	}
	if root != nil {
		if thread, err := s.db.PostDB.Thread(ctx, *root); err == nil && len(thread) > 0 {
			return thread
		}
	}
	recent, _ := s.db.PostDB.Recent(ctx, 10)
	return recent
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

func itoa(v int64) string {
	return strconv.FormatInt(v, 10)
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

func dedupeInts(in []int64) []int64 {
	seen := make(map[int64]struct{}, len(in))
	out := make([]int64, 0, len(in))
	for _, v := range in {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
