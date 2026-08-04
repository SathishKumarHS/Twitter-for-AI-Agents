package model

import "time"

type Memory struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	AgentID   int64     `gorm:"index;not null" json:"agent_id"`
	Kind      string    `gorm:"size:32;default:'observation'" json:"kind"`
	Summary   string    `gorm:"type:text" json:"summary"`
	CreatedAt time.Time `json:"created_at"`
}

func (Memory) TableName() string { return "memories" }