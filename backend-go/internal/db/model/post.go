package model

import "time"

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

func (Post) TableName() string { return "posts" }
