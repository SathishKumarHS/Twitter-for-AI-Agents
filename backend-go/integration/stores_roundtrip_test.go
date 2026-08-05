//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
)

func TestStoresRoundTrip(t *testing.T) {
	ctx := context.Background()
	reset(ctx)

	id, err := database.AgentsDB.Create(ctx, model.Agent{
		Handle:    "int-agent",
		Name:      "Int Agent",
		Persona:   "Precise and factual.",
		Goals:     []string{"Inform users"},
		Policies:  []string{"Be accurate"},
		Interests: []string{"tech"},
		Active:    true,
	})
	if err != nil {
		t.Fatalf("create agent: %v", err)
	}

	got, err := database.AgentsDB.Get(ctx, id)
	if err != nil {
		t.Fatalf("get agent: %v", err)
	}
	if got.Handle != "int-agent" || !got.Active {
		t.Fatalf("agent mismatch: %+v", got)
	}

	byHandle, err := database.AgentsDB.GetByHandle(ctx, "int-agent")
	if err != nil || byHandle == nil || byHandle.ID != id {
		t.Fatalf("get by handle failed: err=%v agent=%+v", err, byHandle)
	}

	list, err := database.AgentsDB.ListAll(ctx)
	if err != nil || len(list) != 1 {
		t.Fatalf("list agents: err=%v n=%d", err, len(list))
	}

	mainPost, err := database.PostDB.Create(ctx, id, "Main integration post", 0)
	if err != nil {
		t.Fatalf("create main post: %v", err)
	}
	if _, err := database.PostDB.Create(ctx, id, "Reply from the same agent", mainPost.ID); err != nil {
		t.Fatalf("create reply: %v", err)
	}

	thread, err := database.PostDB.Thread(ctx, mainPost.ID)
	if err != nil || len(thread) != 2 {
		t.Fatalf("thread: err=%v n=%d", err, len(thread))
	}

	tree, err := database.PostDB.ThreadTree(ctx, mainPost.ID)
	if err != nil || len(tree) != 1 || len(tree[0].Comments) != 1 {
		t.Fatalf("thread tree: err=%v roots=%d comments=%d", err, len(tree), len(tree[0].Comments))
	}

	if _, err := database.PostDB.Timeline(ctx, id, 10); err != nil {
		t.Fatalf("timeline: %v", err)
	}

	if err := database.LastSeenDB.SetSeen(ctx, id, mainPost.ID); err != nil {
		t.Fatalf("set seen: %v", err)
	}
	if seen := database.LastSeenDB.LastSeen(ctx, id); seen != mainPost.ID {
		t.Fatalf("seen = %d, want %d", seen, mainPost.ID)
	}

	if err := database.SubscriptionsDB.Subscribe(ctx, id, mainPost.ID); err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	subs, err := database.SubscriptionsDB.ListSubscribers(ctx, mainPost.ID)
	if err != nil || len(subs) != 1 || subs[0] != id {
		t.Fatalf("subscribers: err=%v subs=%v", err, subs)
	}
}

func TestMemoriesRoundTrip(t *testing.T) {
	ctx := context.Background()
	reset(ctx)

	id, err := database.AgentsDB.Create(ctx, model.Agent{
		Handle: "memory-agent", Name: "Memory Agent", Persona: "p", Active: false,
	})
	if err != nil {
		t.Fatalf("create agent: %v", err)
	}

	if err := database.MemoriesDB.Add(ctx, id, "learning", "remember-one"); err != nil {
		t.Fatalf("add memory one: %v", err)
	}
	if err := database.MemoriesDB.Add(ctx, id, "learning", "remember-two"); err != nil {
		t.Fatalf("add memory two: %v", err)
	}
	if err := database.MemoriesDB.Add(ctx, id, "", "remember-default-kind"); err != nil {
		t.Fatalf("add memory default kind: %v", err)
	}

	list, err := database.MemoriesDB.List(ctx, id, 10)
	if err != nil {
		t.Fatalf("list memories: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("memories = %d, want 3", len(list))
	}
	if list[0].Summary != "remember-default-kind" {
		t.Errorf("first memory = %q, want remember-default-kind", list[0].Summary)
	}
	if list[0].Kind != "observation" {
		t.Errorf("default kind = %q, want observation", list[0].Kind)
	}

	limited, err := database.MemoriesDB.List(ctx, id, 2)
	if err != nil {
		t.Fatalf("list limited memories: %v", err)
	}
	if len(limited) != 2 {
		t.Fatalf("limited memories = %d, want 2", len(limited))
	}
}

func TestAgentsListActive(t *testing.T) {
	ctx := context.Background()
	reset(ctx)

	activeID, err := database.AgentsDB.Create(ctx, model.Agent{
		Handle: "active-agent", Name: "Active Agent", Persona: "p", Active: true,
	})
	if err != nil {
		t.Fatalf("create active agent: %v", err)
	}
	if _, err := database.AgentsDB.Create(ctx, model.Agent{
		Handle: "paused-agent", Name: "Paused Agent", Persona: "p", Active: false,
	}); err != nil {
		t.Fatalf("create paused agent: %v", err)
	}

	list, err := database.AgentsDB.ListActive(ctx)
	if err != nil {
		t.Fatalf("list active: %v", err)
	}
	if len(list) != 1 || list[0].ID != activeID {
		t.Fatalf("list active = %+v, want [%d]", list, activeID)
	}
}

func TestAgentsSetActive(t *testing.T) {
	ctx := context.Background()
	reset(ctx)

	id, err := database.AgentsDB.Create(ctx, model.Agent{
		Handle: "toggle-agent", Name: "Toggle Agent", Persona: "p", Active: true,
	})
	if err != nil {
		t.Fatalf("create agent: %v", err)
	}

	if err := database.AgentsDB.SetActive(ctx, id, false); err != nil {
		t.Fatalf("set inactive: %v", err)
	}
	got, err := database.AgentsDB.Get(ctx, id)
	if err != nil {
		t.Fatalf("get agent: %v", err)
	}
	if got.Active {
		t.Error("agent active after SetActive(false)")
	}

	if err := database.AgentsDB.SetActive(ctx, id, true); err != nil {
		t.Fatalf("set active: %v", err)
	}
	got, err = database.AgentsDB.Get(ctx, id)
	if err != nil {
		t.Fatalf("get agent: %v", err)
	}
	if !got.Active {
		t.Error("agent inactive after SetActive(true)")
	}
}

func TestPostStoreNewPostsAfter(t *testing.T) {
	ctx := context.Background()
	reset(ctx)

	a1, err := database.AgentsDB.Create(ctx, model.Agent{
		Handle: "npa-one", Name: "NPA One", Persona: "p", Active: false,
	})
	if err != nil {
		t.Fatalf("create a1: %v", err)
	}
	a2, err := database.AgentsDB.Create(ctx, model.Agent{
		Handle: "npa-two", Name: "NPA Two", Persona: "p", Active: false,
	})
	if err != nil {
		t.Fatalf("create a2: %v", err)
	}

	p1, err := database.PostDB.Create(ctx, a2, "first post", 0)
	if err != nil {
		t.Fatalf("create p1: %v", err)
	}

	after, err := database.PostDB.NewPostsAfter(ctx, a1)
	if err != nil {
		t.Fatalf("new posts after: %v", err)
	}
	if len(after) != 1 || after[0].ID != p1.ID {
		t.Fatalf("new posts after = %+v, want [%d]", after, p1.ID)
	}

	if err := database.LastSeenDB.SetSeen(ctx, a1, p1.ID); err != nil {
		t.Fatalf("set seen: %v", err)
	}

	p2, err := database.PostDB.Create(ctx, a2, "second post", 0)
	if err != nil {
		t.Fatalf("create p2: %v", err)
	}
	if _, err := database.PostDB.Create(ctx, a1, "own post", 0); err != nil {
		t.Fatalf("create own post: %v", err)
	}

	after, err = database.PostDB.NewPostsAfter(ctx, a1)
	if err != nil {
		t.Fatalf("new posts after: %v", err)
	}
	if len(after) != 1 || after[0].ID != p2.ID {
		t.Fatalf("new posts after = %+v, want [%d]", after, p2.ID)
	}
}

func TestPostStoreLastMainPost(t *testing.T) {
	ctx := context.Background()
	reset(ctx)

	id, err := database.AgentsDB.Create(ctx, model.Agent{
		Handle: "lmp-agent", Name: "LMP Agent", Persona: "p", Active: false,
	})
	if err != nil {
		t.Fatalf("create agent: %v", err)
	}

	if _, err := database.PostDB.LastMainPost(ctx, id); err == nil {
		t.Fatal("expected error before any main post")
	}

	main, err := database.PostDB.Create(ctx, id, "main post", 0)
	if err != nil {
		t.Fatalf("create main post: %v", err)
	}
	if _, err := database.PostDB.Create(ctx, id, "a reply", main.ID); err != nil {
		t.Fatalf("create reply: %v", err)
	}

	ts, err := database.PostDB.LastMainPost(ctx, id)
	if err != nil {
		t.Fatalf("last main post: %v", err)
	}
	if ts.IsZero() {
		t.Error("last main post time is zero")
	}
}
