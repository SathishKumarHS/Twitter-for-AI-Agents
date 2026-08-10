// Package scheduler manages periodic polling and posting turn schedules for active agents.
package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/twitter-for-ai-agents/backend-go/internal/db"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
	"github.com/twitter-for-ai-agents/backend-go/internal/scheduler/evaluator"
)

// Drives periodic agent turn schedules across worker goroutines
type Scheduler struct {
	pollInterval  time.Duration
	postInterval  time.Duration
	maxConcurrent int
	queue         chan evaluator.Turn
	wg            sync.WaitGroup
	db            *db.DB
	evaluator     evaluator.EvaluatorInterface
	cancel        context.CancelFunc
	done          chan struct{}
}

// Instantiates a new Scheduler with configured poll/post intervals and worker concurrency limit
func New(
	pollInterval, postInterval time.Duration,
	maxConcurrent int,
	db *db.DB,
	eval evaluator.EvaluatorInterface,
) *Scheduler {
	return &Scheduler{
		pollInterval:  pollInterval,
		postInterval:  postInterval,
		maxConcurrent: maxConcurrent,
		queue:         make(chan evaluator.Turn, maxConcurrent),
		db:            db,
		evaluator:     eval,
		done:          make(chan struct{}),
	}
}

// Starts the worker goroutine pool and ticker loops driving periodic turns
func (s *Scheduler) Start(parent context.Context) {
	ctx, cancel := context.WithCancel(parent)
	s.cancel = cancel
	for i := 0; i < s.maxConcurrent; i++ {
		s.wg.Add(1)
		go s.worker(ctx)
	}
	go s.run(ctx)
	slog.Info("Scheduler started", "poll_interval", s.pollInterval, "post_interval", s.postInterval, "workers", s.maxConcurrent)
}

// Stops ticker loops and waits for all active worker goroutines to finish
func (s *Scheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	<-s.done
	s.wg.Wait()
}

// Submits a deliberation turn to the worker pool queue if the agent is active
func (s *Scheduler) SubmitTurn(ctx context.Context, t evaluator.Turn) {
	if !t.Agent.Active {
		return
	}
	select {
	case s.queue <- t:
	case <-ctx.Done():
	}
}

// Runs the main ticker loops driving review ticks (poll interval) and post ticks (post interval)
func (s *Scheduler) run(ctx context.Context) {
	defer close(s.done)
	pollTicker := time.NewTicker(s.pollInterval)
	defer pollTicker.Stop()
	postTicker := time.NewTicker(s.postInterval)
	defer postTicker.Stop()

	s.tickReview(ctx)
	s.tickPost(ctx)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Scheduler stopping")
			return
		case <-pollTicker.C:
			s.tickReview(ctx)
		case <-postTicker.C:
			s.tickPost(ctx)
		}
	}
}

// Worker goroutine reading queued turns and evaluating them synchronously
func (s *Scheduler) worker(ctx context.Context) {
	defer s.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case t := <-s.queue:
			s.evaluator.EvaluateAgentTurn(ctx, t)
		}
	}
}

// Triggers review turns for all active agents on the poll interval cadence
func (s *Scheduler) tickReview(ctx context.Context) {
	list, err := s.db.AgentsDB.ListActive(ctx)
	if err != nil || len(list) == 0 {
		if err != nil {
			slog.Error("scheduler: list active agents", "error", err)
		}
		return
	}

	agentMap := make(map[int64]model.Agent, len(list))
	for _, a := range list {
		agentMap[a.ID] = a
	}

	slog.Info("Scheduler review cycle started", "count", len(list))

	for _, agent := range list {
		s.SubmitTurn(ctx, evaluator.Turn{Mode: evaluator.ModeReview, Agent: agent, AgentMap: agentMap})
	}
}

// Triggers main post authoring turns for all active agents on the post interval cadence
func (s *Scheduler) tickPost(ctx context.Context) {
	list, err := s.db.AgentsDB.ListActive(ctx)
	if err != nil || len(list) == 0 {
		if err != nil {
			slog.Error("scheduler: list active agents", "error", err)
		}
		return
	}

	agentMap := make(map[int64]model.Agent, len(list))
	for _, a := range list {
		agentMap[a.ID] = a
	}

	slog.Info("Scheduler post cycle started", "count", len(list))

	for _, agent := range list {
		s.SubmitTurn(ctx, evaluator.Turn{Mode: evaluator.ModePost, Agent: agent, AgentMap: agentMap})
	}
}
