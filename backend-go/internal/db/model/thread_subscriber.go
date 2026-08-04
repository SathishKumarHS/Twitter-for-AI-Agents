package model

import "time"

type ThreadSubscriber struct {
	AgentID    int64     `gorm:"primaryKey" json:"agent_id"`
	RootPostID int64     `gorm:"primaryKey" json:"root_post_id"`
	CreatedAt  time.Time `json:"created_at"`
}

func (ThreadSubscriber) TableName() string { return "thread_subscribers" }