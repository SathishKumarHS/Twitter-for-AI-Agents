// Package posts implements storage and queries for posts, replies, threads, and timelines.
package posts

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"

	lastseen "github.com/twitter-for-ai-agents/backend-go/internal/db/last_seen"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
)

// GORM-backed storage implementation for posts and replies
type PostsDB struct {
	g          *gorm.DB
	LastSeenDB lastseen.LastSeenInterface
}

// Instantiates a new PostsDB store implementing PostInterface
func NewPostsDB(g *gorm.DB, lastSeenDB lastseen.LastSeenInterface) PostInterface {
	return &PostsDB{
		g:          g,
		LastSeenDB: lastSeenDB,
	}
}

// Fetches posts created after an agent's last-seen ID (excluding the agent's own posts)
func (s *PostsDB) NewPostsAfter(ctx context.Context, agentID int64) ([]model.Post, error) {
	var out []model.Post
	err := s.g.WithContext(ctx).
		Where("id > ? AND agent_id <> ?", s.LastSeenDB.LastSeen(ctx, agentID), agentID).
		Order("id").
		Find(&out).Error
	return out, err
}

// Creates a new post or reply, resolves @mentions, sets root thread ID, and persists to DB
func (s *PostsDB) Create(ctx context.Context, agentID int64, body string, parentID int64) (model.Post, error) {
	p := model.Post{AgentID: agentID, Body: body, Kind: "post"}
	if parentID != 0 {
		var parent model.Post
		if err := s.g.WithContext(ctx).First(&parent, parentID).Error; err != nil {
			return model.Post{}, err
		}
		p.Kind = "reply"
		p.ParentID = &parentID
		root := parent.RootID
		if root == nil {
			root = &parent.ID
		}
		p.RootID = root
	}

	handles := extractMentionHandles(body)
	if len(handles) > 0 {
		var mentionIDs []int64
		_ = s.g.WithContext(ctx).Table("agents").
			Where("LOWER(handle) IN ?", handles).
			Pluck("id", &mentionIDs).Error
		p.Mentions = mentionIDs
	}

	if err := s.g.WithContext(ctx).Create(&p).Error; err != nil {
		return model.Post{}, err
	}
	return p, nil
}

// Helper: extracts unique @handle mention strings from post body text
func extractMentionHandles(body string) []string {
	words := strings.Fields(body)
	seen := make(map[string]bool)
	var handles []string
	for _, w := range words {
		w = strings.Trim(w, ".,!?:;\"'()[]{}#|-")
		if strings.HasPrefix(w, "@") && len(w) > 1 {
			handle := strings.ToLower(strings.TrimPrefix(w, "@"))
			if !seen[handle] {
				seen[handle] = true
				handles = append(handles, handle)
			}
		}
	}
	return handles
}

// Fetches the most recent posts on the network timeline up to limit
func (s *PostsDB) Recent(ctx context.Context, limit int) ([]model.Post, error) {
	var out []model.Post
	err := s.g.WithContext(ctx).
		Preload("Comments", func(db *gorm.DB) *gorm.DB { return db.Order("id") }).
		Order("id DESC").Limit(limit).Find(&out).Error
	return out, err
}

// Fetches all posts belonging to a conversation thread (root post and all replies)
func (s *PostsDB) Thread(ctx context.Context, rootID int64) ([]model.Post, error) {
	var out []model.Post
	err := s.g.WithContext(ctx).
		Where("root_id = ? OR id = ?", rootID, rootID).
		Order("created_at, id").
		Find(&out).Error
	return out, err
}

// Fetches a thread and nests replies under parent posts to form a tree structure
func (s *PostsDB) ThreadTree(ctx context.Context, rootID int64) ([]model.Post, error) {
	flat, err := s.Thread(ctx, rootID)
	if err != nil {
		return nil, err
	}
	tree := buildThreadTree(flat)
	roots := make([]model.Post, 0, len(tree))
	for i := range tree {
		if tree[i].ID == rootID || tree[i].RootID == nil {
			roots = append(roots, tree[i])
		}
	}
	if len(roots) == 0 && len(tree) > 0 {
		roots = append(roots, tree[0])
	}
	return roots, nil
}

// Helper: builds a nested tree of comments from a flat list of thread posts
func buildThreadTree(flat []model.Post) []model.Post {
	byParent := make(map[int64][]model.Post)
	var roots []model.Post
	for _, p := range flat {
		if p.ParentID != nil {
			byParent[*p.ParentID] = append(byParent[*p.ParentID], p)
		} else {
			roots = append(roots, p)
		}
	}
	var attach func(p model.Post) model.Post
	attach = func(p model.Post) model.Post {
		if kids, ok := byParent[p.ID]; ok {
			for _, k := range kids {
				p.Comments = append(p.Comments, attach(k))
			}
		}
		return p
	}
	for i := range roots {
		roots[i] = attach(roots[i])
	}
	return roots
}

// Fetches an agent's personal timeline posts up to limit
func (s *PostsDB) Timeline(ctx context.Context, agentID int64, limit int) ([]model.Post, error) {
	var out []model.Post
	err := s.g.WithContext(ctx).
		Preload("Comments", func(db *gorm.DB) *gorm.DB { return db.Order("id") }).
		Where("agent_id = ?", agentID).
		Order("id DESC").
		Limit(limit).
		Find(&out).Error
	return out, err
}

// Fetches the timestamp of the last top-level main post created by an agent
func (s *PostsDB) LastMainPost(ctx context.Context, agentID int64) (time.Time, error) {
	var p model.Post
	err := s.g.WithContext(ctx).
		Where("agent_id = ? AND kind = ?", agentID, "post").
		Order("created_at DESC").
		First(&p).Error
	if err != nil {
		return time.Time{}, err
	}
	return p.CreatedAt, nil
}
