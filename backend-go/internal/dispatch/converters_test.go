package dispatch

import (
	"testing"

	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
)

func TestToProtoAgent(t *testing.T) {
	agent := model.Agent{
		ID:           10,
		Handle:       "tech_bot",
		Name:         "Tech Bot",
		Persona:      "Tech Enthusiast",
		Goals:        []string{"Share news", "Help users"},
		Policies:     []string{"No spam"},
		Interests:    []string{"Go", "AI"},
		AvatarURL:    "https://example.com/avatar.png",
		Active:       true,
		Description:  "AI news poster",
		Instructions: "Post about technology",
	}

	pbAgent := toProtoAgent(agent)

	if pbAgent.Id != 10 {
		t.Errorf("expected Id 10, got %d", pbAgent.Id)
	}
	if pbAgent.Handle != "tech_bot" {
		t.Errorf("expected Handle 'tech_bot', got '%s'", pbAgent.Handle)
	}
	if pbAgent.Name != "Tech Bot" {
		t.Errorf("expected Name 'Tech Bot', got '%s'", pbAgent.Name)
	}
	if pbAgent.Persona != "Tech Enthusiast" {
		t.Errorf("expected Persona 'Tech Enthusiast', got '%s'", pbAgent.Persona)
	}
	if len(pbAgent.Goals) != 2 || pbAgent.Goals[0] != "Share news" {
		t.Errorf("unexpected Goals: %v", pbAgent.Goals)
	}
	if len(pbAgent.Policies) != 1 || pbAgent.Policies[0] != "No spam" {
		t.Errorf("unexpected Policies: %v", pbAgent.Policies)
	}
	if len(pbAgent.Interests) != 2 || pbAgent.Interests[0] != "Go" {
		t.Errorf("unexpected Interests: %v", pbAgent.Interests)
	}
	if pbAgent.AvatarUrl != "https://example.com/avatar.png" {
		t.Errorf("expected AvatarUrl 'https://example.com/avatar.png', got '%s'", pbAgent.AvatarUrl)
	}
	if !pbAgent.Active {
		t.Errorf("expected Active true, got false")
	}
	if pbAgent.Description != "AI news poster" {
		t.Errorf("expected Description 'AI news poster', got '%s'", pbAgent.Description)
	}
	if pbAgent.Instructions != "Post about technology" {
		t.Errorf("expected Instructions 'Post about technology', got '%s'", pbAgent.Instructions)
	}
}

func TestToProtoPost(t *testing.T) {
	parentID := int64(5)
	rootID := int64(1)

	post := model.Post{
		ID:       100,
		AgentID:  2,
		Body:     "Test reply body",
		Kind:     "reply",
		ParentID: &parentID,
		RootID:   &rootID,
		Mentions: []int64{10},
	}

	agentMap := map[int64]model.Agent{
		2: {ID: 2, Handle: "alice", Name: "Alice Smith"},
	}

	pbPost := toProtoPost(post, agentMap)

	if pbPost.Id != 100 {
		t.Errorf("expected Id 100, got %d", pbPost.Id)
	}
	if pbPost.AgentId != 2 {
		t.Errorf("expected AgentId 2, got %d", pbPost.AgentId)
	}
	if pbPost.Body != "Test reply body" {
		t.Errorf("expected Body 'Test reply body', got '%s'", pbPost.Body)
	}
	if pbPost.Kind != "reply" {
		t.Errorf("expected Kind 'reply', got '%s'", pbPost.Kind)
	}
	if pbPost.ParentId != 5 {
		t.Errorf("expected ParentId 5, got %d", pbPost.ParentId)
	}
	if pbPost.RootId != 1 {
		t.Errorf("expected RootId 1, got %d", pbPost.RootId)
	}
	if len(pbPost.Mentions) != 1 || pbPost.Mentions[0] != 10 {
		t.Errorf("unexpected Mentions: %v", pbPost.Mentions)
	}
	if pbPost.AuthorHandle != "alice" {
		t.Errorf("expected AuthorHandle 'alice', got '%s'", pbPost.AuthorHandle)
	}
	if pbPost.AuthorName != "Alice Smith" {
		t.Errorf("expected AuthorName 'Alice Smith', got '%s'", pbPost.AuthorName)
	}

	// Test Nil ParentID and RootID, and missing agent in map
	postNil := model.Post{
		ID:      101,
		AgentID: 99,
		Body:    "Standalone post",
		Kind:    "post",
	}

	pbPostNil := toProtoPost(postNil, agentMap)
	if pbPostNil.ParentId != 0 {
		t.Errorf("expected ParentId 0 for nil parent, got %d", pbPostNil.ParentId)
	}
	if pbPostNil.RootId != 0 {
		t.Errorf("expected RootId 0 for nil root, got %d", pbPostNil.RootId)
	}
	if pbPostNil.AuthorHandle != "" || pbPostNil.AuthorName != "" {
		t.Errorf("expected empty author handle and name when missing in map")
	}
}
