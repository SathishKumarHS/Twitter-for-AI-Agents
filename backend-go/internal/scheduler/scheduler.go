package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
	"github.com/twitter-for-ai-agents/backend-go/internal/logs"
	"github.com/twitter-for-ai-agents/backend-go/internal/redis"
	"github.com/twitter-for-ai-agents/backend-go/internal/db"
	"github.com/twitter-for-ai-agents/backend-go/internal/dispatch"
	"github.com/twitter-for-ai-agents/backend-go/internal/events"
)

const lockTTLMargin = 60 * time.Second

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

func New(interval, postCooldown, turnTimeout time.Duration, maxConcurrent int, locks redis.RedisInterface, db *db.DB, d dispatch.DispatcherInterface, hub events.HubInterface, c redis.RedisInterface) SchedulerInterface {
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
	logs.Info(logs.SystemNamespace, msg)
}

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

	s.tick(ctx)

	for {
		select {
		case <-ctx.Done():
			logs.Info(logs.SystemNamespace, "Scheduler stopping")
			return
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

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

func (s *Scheduler) tick(ctx context.Context) {
	list, err := s.db.AgentsDB.ListActive(ctx)
	if err != nil || len(list) == 0 {
		if err != nil {
			logs.Error(logs.SystemNamespace, fmt.Sprintf("scheduler: list active agents: %v", err))
		}
		return
	}

	agentMap := make(map[int64]model.Agent, len(list))
	for _, a := range list {
		agentMap[a.ID] = a
	}

	logs.Info(logs.SystemNamespace, fmt.Sprintf("Scheduler review cycle started for %d active agents", len(list)))

	for _, agent := range list {
		s.submitTurn(ctx, modePost, agent, agentMap, nil)
	}
	for _, agent := range list {
		s.submitTurn(ctx, modeReview, agent, agentMap, nil)
	}
}

func (s *Scheduler) submitTurn(ctx context.Context, mode turnMode, agent model.Agent, agentMap map[int64]model.Agent, targetPost *model.Post) {
	if !agent.Active {
		return
	}
	select {
	case s.queue <- turn{mode: mode, agent: agent, agentMap: agentMap, targetPost: targetPost}:
	case <-ctx.Done():
	}
}
