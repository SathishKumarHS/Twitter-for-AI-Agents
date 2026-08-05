//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
	"github.com/twitter-for-ai-agents/backend-go/internal/logs"
)

func doGet(t *testing.T, url string) *http.Response {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("get %s: %v", url, err)
	}
	return resp
}

func postJSON(t *testing.T, url string, body []byte) *http.Response {
	t.Helper()
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("post %s: %v", url, err)
	}
	return resp
}

func createAgent(t *testing.T, handle string) model.Agent {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"handle": handle, "name": "Test Agent", "persona": "Curious, helpful, and concise.",
		"goals": []string{"Increase engagement"}, "policies": []string{"Be respectful"},
		"interests": []string{"food", "tech"},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	resp := postJSON(t, baseURL+"/api/agents", body)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create agent status = %d", resp.StatusCode)
	}
	var a model.Agent
	if err := json.NewDecoder(resp.Body).Decode(&a); err != nil {
		t.Fatalf("decode agent: %v", err)
	}
	return a
}

func TestCreateAgentPersistsToDBAndRedis(t *testing.T) {
	ctx := context.Background()
	reset(ctx)

	a := createAgent(t, "testagent")
	if a.Handle != "testagent" || !a.Active {
		t.Fatalf("created agent mismatch: %+v", a)
	}

	got, err := database.AgentsDB.Get(ctx, a.ID)
	if err != nil {
		t.Fatalf("get agent from db: %v", err)
	}
	if got.Handle != "testagent" || got.Name != "Test Agent" || !got.Active {
		t.Errorf("db row mismatch: %+v", got)
	}

	waitFor(t, 30*time.Second, "agent:active cache", func() bool {
		return raw.Get(ctx, fmt.Sprintf("agent:active:%d", a.ID)).Val() != ""
	})
}

func TestAgentPauseAndResumeLifecycle(t *testing.T) {
	ctx := context.Background()
	reset(ctx)

	a := createAgent(t, "pause_me")

	pause, err := http.Post(fmt.Sprintf("%s/api/agents/%d/pause", baseURL, a.ID), "application/json", nil)
	if err != nil {
		t.Fatalf("pause agent: %v", err)
	}
	pause.Body.Close()
	if pause.StatusCode != http.StatusOK {
		t.Fatalf("pause status = %d", pause.StatusCode)
	}

	got, err := database.AgentsDB.Get(ctx, a.ID)
	if err != nil {
		t.Fatalf("get agent: %v", err)
	}
	if got.Active {
		t.Error("agent still active in db after pause")
	}
	if active, found := rc.IsAgentActive(ctx, a.ID); !found || active {
		t.Errorf("redis active = %v/%v, want false/true", active, found)
	}

	resume, err := http.Post(fmt.Sprintf("%s/api/agents/%d/resume", baseURL, a.ID), "application/json", nil)
	if err != nil {
		t.Fatalf("resume agent: %v", err)
	}
	resume.Body.Close()
	if resume.StatusCode != http.StatusOK {
		t.Fatalf("resume status = %d", resume.StatusCode)
	}

	got, err = database.AgentsDB.Get(ctx, a.ID)
	if err != nil {
		t.Fatalf("get agent: %v", err)
	}
	if !got.Active {
		t.Error("agent not active in db after resume")
	}
	if active, found := rc.IsAgentActive(ctx, a.ID); !found || !active {
		t.Errorf("redis active = %v/%v, want true/true", active, found)
	}
}

func TestFeedCacheHitAndMiss(t *testing.T) {
	ctx := context.Background()
	reset(ctx)

	id, err := database.AgentsDB.Create(ctx, model.Agent{
		Handle: "feed-agent", Name: "Feed Agent", Persona: "p", Active: false,
	})
	if err != nil {
		t.Fatalf("create agent: %v", err)
	}
	mainPost, err := database.PostDB.Create(ctx, id, "Feed cache body", 0)
	if err != nil {
		t.Fatalf("create post: %v", err)
	}
	if _, err := database.PostDB.Create(ctx, id, "Feed reply", mainPost.ID); err != nil {
		t.Fatalf("create reply: %v", err)
	}

	first := doGet(t, baseURL+"/api/posts?limit=20")
	defer first.Body.Close()
	if first.StatusCode != http.StatusOK {
		t.Fatalf("feed status = %d", first.StatusCode)
	}
	if first.Header.Get("X-Cache") != "MISS" {
		t.Errorf("first feed X-Cache = %q, want MISS", first.Header.Get("X-Cache"))
	}
	var feed []model.Post
	if err := json.NewDecoder(first.Body).Decode(&feed); err != nil {
		t.Fatalf("decode feed: %v", err)
	}
	if len(feed) != 2 {
		t.Fatalf("feed has %d posts, want 2", len(feed))
	}

	second := doGet(t, baseURL+"/api/posts?limit=20")
	defer second.Body.Close()
	if second.Header.Get("X-Cache") != "HIT" {
		t.Errorf("second feed X-Cache = %q, want HIT", second.Header.Get("X-Cache"))
	}
}

func TestThreadEndpoint(t *testing.T) {
	ctx := context.Background()
	reset(ctx)

	id, err := database.AgentsDB.Create(ctx, model.Agent{
		Handle: "thread-agent", Name: "Thread Agent", Persona: "p", Active: false,
	})
	if err != nil {
		t.Fatalf("create agent: %v", err)
	}
	mainPost, err := database.PostDB.Create(ctx, id, "Thread root", 0)
	if err != nil {
		t.Fatalf("create root: %v", err)
	}
	if _, err := database.PostDB.Create(ctx, id, "Thread reply", mainPost.ID); err != nil {
		t.Fatalf("create reply: %v", err)
	}

	resp := doGet(t, fmt.Sprintf("%s/api/posts/thread/%d", baseURL, mainPost.ID))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("thread status = %d", resp.StatusCode)
	}

	var roots []model.Post
	if err := json.NewDecoder(resp.Body).Decode(&roots); err != nil {
		t.Fatalf("decode thread: %v", err)
	}
	if len(roots) != 1 {
		t.Fatalf("thread roots = %d, want 1", len(roots))
	}
	if roots[0].ID != mainPost.ID {
		t.Errorf("root id = %d, want %d", roots[0].ID, mainPost.ID)
	}
	if len(roots[0].Comments) != 1 {
		t.Errorf("root comments = %d, want 1", len(roots[0].Comments))
	}
}

func TestLogsEndpoint(t *testing.T) {
	ctx := context.Background()
	reset(ctx)

	a := createAgent(t, "log_agent")

	resp := doGet(t, baseURL+"/api/logs?limit=100")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("logs status = %d", resp.StatusCode)
	}

	var entries []logs.LogEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		t.Fatalf("decode logs: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("logs endpoint returned no entries")
	}
	for _, e := range entries {
		if strings.Contains(e.Message, a.Handle) {
			return
		}
	}
	t.Fatalf("no log entry mentions created agent @%s", a.Handle)
}

func TestListAgentsEndpoint(t *testing.T) {
	ctx := context.Background()
	reset(ctx)

	createAgent(t, "list_one")
	createAgent(t, "list_two")

	resp := doGet(t, baseURL+"/api/agents")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list agents status = %d", resp.StatusCode)
	}
	var agents []model.Agent
	if err := json.NewDecoder(resp.Body).Decode(&agents); err != nil {
		t.Fatalf("decode agents: %v", err)
	}
	if len(agents) != 2 {
		t.Fatalf("agents = %d, want 2", len(agents))
	}
}

func TestGetAgentEndpoint(t *testing.T) {
	ctx := context.Background()
	reset(ctx)

	a := createAgent(t, "get_me")

	resp := doGet(t, fmt.Sprintf("%s/api/agents/%d", baseURL, a.ID))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get agent status = %d", resp.StatusCode)
	}
	var got model.Agent
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode agent: %v", err)
	}
	if got.Handle != "get_me" {
		t.Errorf("handle = %q, want get_me", got.Handle)
	}
}

func TestAgentPostsEndpoint(t *testing.T) {
	ctx := context.Background()
	reset(ctx)

	id, err := database.AgentsDB.Create(ctx, model.Agent{
		Handle: "posts-agent", Name: "Posts Agent", Persona: "p", Active: false,
	})
	if err != nil {
		t.Fatalf("create agent: %v", err)
	}
	mainPost, err := database.PostDB.Create(ctx, id, "Agent main post", 0)
	if err != nil {
		t.Fatalf("create main post: %v", err)
	}
	if _, err := database.PostDB.Create(ctx, id, "Agent reply", mainPost.ID); err != nil {
		t.Fatalf("create reply: %v", err)
	}

	resp := doGet(t, fmt.Sprintf("%s/api/agents/%d/posts", baseURL, id))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("agent posts status = %d", resp.StatusCode)
	}
	var posts []model.Post
	if err := json.NewDecoder(resp.Body).Decode(&posts); err != nil {
		t.Fatalf("decode posts: %v", err)
	}
	if len(posts) != 2 {
		t.Fatalf("posts = %d, want 2", len(posts))
	}
}

func TestCreateAgentDuplicateHandleConflict(t *testing.T) {
	ctx := context.Background()
	reset(ctx)

	createAgent(t, "dup_handle")

	resp := postJSON(t, baseURL+"/api/agents", []byte(`{"handle":"@dup_handle","name":"Dup","persona":"p"}`))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("duplicate handle status = %d, want 409", resp.StatusCode)
	}
}

func TestCreateAgentInvalidBody(t *testing.T) {
	ctx := context.Background()
	reset(ctx)

	resp := postJSON(t, baseURL+"/api/agents", []byte(`{not json`))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid body status = %d, want 400", resp.StatusCode)
	}
}

func TestCreateAgentValidationErrors(t *testing.T) {
	ctx := context.Background()
	reset(ctx)

	cases := []string{
		`{"handle":"","name":"No Handle","persona":"p"}`,
		`{"handle":"bad handle!","name":"Bad Handle","persona":"p"}`,
		`{"handle":"noname","name":"","persona":"p"}`,
	}
	for _, body := range cases {
		resp := postJSON(t, baseURL+"/api/agents", []byte(body))
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("validation body %s status = %d, want 400", body, resp.StatusCode)
		}
	}
}

func TestGetAgentNotFoundAndInvalidID(t *testing.T) {
	ctx := context.Background()
	reset(ctx)

	missing := doGet(t, baseURL+"/api/agents/99999")
	missing.Body.Close()
	if missing.StatusCode != http.StatusNotFound {
		t.Errorf("missing agent status = %d, want 404", missing.StatusCode)
	}

	badID := doGet(t, baseURL+"/api/agents/abc")
	badID.Body.Close()
	if badID.StatusCode != http.StatusBadRequest {
		t.Errorf("bad id status = %d, want 400", badID.StatusCode)
	}

	pauseBad := postJSON(t, baseURL+"/api/agents/abc/pause", nil)
	pauseBad.Body.Close()
	if pauseBad.StatusCode != http.StatusBadRequest {
		t.Errorf("pause bad id status = %d, want 400", pauseBad.StatusCode)
	}
}
