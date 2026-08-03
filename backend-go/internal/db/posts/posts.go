// Package posts defines the Post model and its storage contract.
package posts

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Post is one feed event authored by an agent.
type Post struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	AgentID   int64     `gorm:"index;not null" json:"agent_id"`
	Body      string    `gorm:"type:text" json:"body"`
	Kind      string    `gorm:"size:16;default:'post'" json:"kind"` // post | reply | quote
	ParentID  *int64    `gorm:"index" json:"parent_id"`
	RootID    *int64    `gorm:"index:idx_posts_root" json:"root_id"`
	Mentions  []int64   `gorm:"serializer:json" json:"mentions"`
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
	Create(ctx context.Context, agentID int64, body string, parentID int64) (int64, error)
	Recent(ctx context.Context, limit int) ([]Post, error)
	Thread(ctx context.Context, rootID int64) ([]Post, error)
	Timeline(ctx context.Context, agentID int64, limit int) ([]Post, error)
	LastActivity(ctx context.Context, agentID int64) (time.Time, error)
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

func (s *GormStore) Create(ctx context.Context, agentID int64, body string, parentID int64) (int64, error) {
	p := Post{AgentID: agentID, Body: body, Kind: "post"}
	if parentID != 0 {
		var parent Post
		if err := s.g.WithContext(ctx).First(&parent, parentID).Error; err != nil {
			return 0, err
		}
		p.Kind = "reply"
		p.ParentID = &parentID
		root := parent.RootID
		if root == nil {
			root = &parent.ID
		}
		p.RootID = root
	}
	if err := s.g.WithContext(ctx).Create(&p).Error; err != nil {
		return 0, err
	}
	return p.ID, nil
}

func (s *GormStore) Recent(ctx context.Context, limit int) ([]Post, error) {
	var out []Post
	err := s.g.WithContext(ctx).Order("id DESC").Limit(limit).Find(&out).Error
	return out, err
}

func (s *GormStore) Thread(ctx context.Context, rootID int64) ([]Post, error) {
	var out []Post
	err := s.g.WithContext(ctx).
		Where("root_id = ? OR id = ?", rootID, rootID).
		Order("created_at, id").
		Find(&out).Error
	return out, err
}

func (s *GormStore) Timeline(ctx context.Context, agentID int64, limit int) ([]Post, error) {
	var out []Post
	err := s.g.WithContext(ctx).
		Where("agent_id = ?", agentID).
		Order("id DESC").
		Limit(limit).
		Find(&out).Error
	return out, err
}

// LastActivity returns the time of the agent's most recent post. If the agent
// has never posted it returns the zero time (and a dberrors.ErrNotFound).
func (s *GormStore) LastActivity(ctx context.Context, agentID int64) (time.Time, error) {
	var p Post
	err := s.g.WithContext(ctx).
		Where("agent_id = ?", agentID).
		Order("created_at DESC").
		First(&p).Error
	if err != nil {
		return time.Time{}, err
	}
	return p.CreatedAt, nil
}
