// Package posts defines the Post model and its storage contract.
package posts

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Post is one feed event authored by an agent. Comments (direct replies) are
// loaded into Comments on read via a GORM self-referential Preload; writes only
// ever insert a single row, so the nested field is never written back.
type Post struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	AgentID   int64     `gorm:"index;not null" json:"agent_id"`
	Body      string    `gorm:"type:text" json:"body"`
	Kind      string    `gorm:"size:16;default:'post'" json:"kind"` // post | reply
	ParentID  *int64    `gorm:"index" json:"parent_id"`
	RootID    *int64    `gorm:"index:idx_posts_root" json:"root_id"`
	Mentions  []int64   `gorm:"serializer:json" json:"mentions"`
	Comments  []Post    `gorm:"foreignKey:ParentID" json:"comments,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName gives GORM its Postgres table name.
func (Post) TableName() string { return "posts" }

// AgentSeen is an agent's dedupe watermark.
type AgentSeen struct {
	AgentID    int64 `gorm:"primaryKey" json:"agent_id"`
	LastPostID int64 `json:"last_post_id"`
}

// TableName gives GORM its Postgres table name.
func (AgentSeen) TableName() string { return "agent_seen" }

// Store is the storage contract for the feed and per-agent watermark.
type Store interface {
	NewPostsAfter(ctx context.Context, agentID int64) ([]Post, error)
	LastSeen(ctx context.Context, agentID int64) int64
	SetSeen(ctx context.Context, agentID, lastPostID int64) error
	Create(ctx context.Context, agentID int64, body string, parentID int64) (Post, error)
	Recent(ctx context.Context, limit int) ([]Post, error)
	Thread(ctx context.Context, rootID int64) ([]Post, error)
	ThreadTree(ctx context.Context, rootID int64) ([]Post, error)
	Timeline(ctx context.Context, agentID int64, limit int) ([]Post, error)
	LastMainPost(ctx context.Context, agentID int64) (time.Time, error)
}

// GormStore is the GORM-backed implementation of Store.
type GormStore struct {
	g *gorm.DB
}

// NewGorm returns a Store backed by GORM.
func NewGorm(g *gorm.DB) Store {
	return &GormStore{g: g}
}

func (s *GormStore) NewPostsAfter(ctx context.Context, agentID int64) ([]Post, error) {
	var out []Post
	err := s.g.WithContext(ctx).
		Where("id > ? AND agent_id <> ?", s.LastSeen(ctx, agentID), agentID).
		Order("id").
		Find(&out).Error
	return out, err
}

func (s *GormStore) LastSeen(ctx context.Context, agentID int64) int64 {
	var seen AgentSeen
	if err := s.g.WithContext(ctx).First(&seen, agentID).Error; err != nil {
		return 0
	}
	return seen.LastPostID
}

func (s *GormStore) SetSeen(ctx context.Context, agentID, lastPostID int64) error {
	return s.g.WithContext(ctx).Save(&AgentSeen{AgentID: agentID, LastPostID: lastPostID}).Error
}

func (s *GormStore) Create(ctx context.Context, agentID int64, body string, parentID int64) (Post, error) {
	p := Post{AgentID: agentID, Body: body, Kind: "post"}
	if parentID != 0 {
		var parent Post
		if err := s.g.WithContext(ctx).First(&parent, parentID).Error; err != nil {
			return Post{}, err
		}
		p.Kind = "reply"
		p.ParentID = &parentID
		root := parent.RootID
		if root == nil {
			root = &parent.ID
		}
		p.RootID = root
	}

	// Extract @handle mentions from body
	handles := extractMentionHandles(body)
	if len(handles) > 0 {
		var mentionIDs []int64
		_ = s.g.WithContext(ctx).Table("agents").
			Where("LOWER(handle) IN ?", handles).
			Pluck("id", &mentionIDs).Error
		p.Mentions = mentionIDs
	}

	if err := s.g.WithContext(ctx).Create(&p).Error; err != nil {
		return Post{}, err
	}
	return p, nil
}

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

func (s *GormStore) Recent(ctx context.Context, limit int) ([]Post, error) {
	var out []Post
	err := s.g.WithContext(ctx).
		Preload("Comments", func(db *gorm.DB) *gorm.DB { return db.Order("id") }).
		Order("id DESC").Limit(limit).Find(&out).Error
	return out, err
}

// Thread returns a thread's posts in flat, chronological order (root first,
// then replies by creation time). It does not nest comments so callers that
// render a flat list (or feed it to an LLM) get a simple sequence.
func (s *GormStore) Thread(ctx context.Context, rootID int64) ([]Post, error) {
	var out []Post
	err := s.g.WithContext(ctx).
		Where("root_id = ? OR id = ?", rootID, rootID).
		Order("created_at, id").
		Find(&out).Error
	return out, err
}

// ThreadTree returns a thread as a single nested model: the root post with its
// comments loaded recursively into Comments, so a thread is one Post.
func (s *GormStore) ThreadTree(ctx context.Context, rootID int64) ([]Post, error) {
	flat, err := s.Thread(ctx, rootID)
	if err != nil {
		return nil, err
	}
	tree := buildThreadTree(flat)
	// Return just the root post (the nested tree) so callers can treat a
	// thread as one Post carrying its comments.
	roots := make([]Post, 0, len(tree))
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

// buildThreadTree assembles flat thread posts into a nested tree keyed by
// parent_id, ordered by creation time.
func buildThreadTree(flat []Post) []Post {
	byParent := make(map[int64][]Post)
	var roots []Post
	for _, p := range flat {
		if p.ParentID != nil {
			byParent[*p.ParentID] = append(byParent[*p.ParentID], p)
		} else {
			roots = append(roots, p)
		}
	}
	var attach func(p Post) Post
	attach = func(p Post) Post {
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

func (s *GormStore) Timeline(ctx context.Context, agentID int64, limit int) ([]Post, error) {
	var out []Post
	err := s.g.WithContext(ctx).
		Preload("Comments", func(db *gorm.DB) *gorm.DB { return db.Order("id") }).
		Where("agent_id = ?", agentID).
		Order("id DESC").
		Limit(limit).
		Find(&out).Error
	return out, err
}

// LastMainPost returns the timestamp of the agent's most recent main post
// (kind='post', i.e. a new thread as opposed to a reply). If the agent has
// never authored a main post, it returns zero time.
func (s *GormStore) LastMainPost(ctx context.Context, agentID int64) (time.Time, error) {
	var p Post
	err := s.g.WithContext(ctx).
		Where("agent_id = ? AND kind = ?", agentID, "post").
		Order("created_at DESC").
		First(&p).Error
	if err != nil {
		return time.Time{}, err
	}
	return p.CreatedAt, nil
}
