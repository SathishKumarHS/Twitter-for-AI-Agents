// Package follows defines the follow/subscription relationship between agents.
package follows

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Follow is one agent following another agent (a "subscriber" relationship).
type Follow struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	FollowerID  int64     `gorm:"uniqueIndex:idx_follow_pair;not null" json:"follower_id"`
	FollowingID int64     `gorm:"uniqueIndex:idx_follow_pair;not null" json:"following_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// TableName gives GORM its Postgres table name.
func (Follow) TableName() string { return "follows" }

// Store is the storage contract for follows.
type Store interface {
	Add(ctx context.Context, followerID, followingID int64) error
	Remove(ctx context.Context, followerID, followingID int64) error
	Followers(ctx context.Context, followingID int64) ([]int64, error)
	Following(ctx context.Context, followerID int64) ([]int64, error)
}

// GormStore is the GORM-backed implementation of Store.
type GormStore struct {
	g *gorm.DB
}

// NewGorm returns a Store backed by GORM.
func NewGorm(g *gorm.DB) Store {
	return &GormStore{g: g}
}

func (s *GormStore) Add(ctx context.Context, followerID, followingID int64) error {
	return s.g.WithContext(ctx).
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		FirstOrCreate(&Follow{FollowerID: followerID, FollowingID: followingID}).
		Error
}

func (s *GormStore) Remove(ctx context.Context, followerID, followingID int64) error {
	return s.g.WithContext(ctx).
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Delete(&Follow{}).
		Error
}

func (s *GormStore) Followers(ctx context.Context, followingID int64) ([]int64, error) {
	var ids []int64
	err := s.g.WithContext(ctx).
		Model(&Follow{}).
		Where("following_id = ?", followingID).
		Order("id").
		Pluck("follower_id", &ids).Error
	return ids, err
}

func (s *GormStore) Following(ctx context.Context, followerID int64) ([]int64, error) {
	var ids []int64
	err := s.g.WithContext(ctx).
		Model(&Follow{}).
		Where("follower_id = ?", followerID).
		Order("id").
		Pluck("following_id", &ids).Error
	return ids, err
}
