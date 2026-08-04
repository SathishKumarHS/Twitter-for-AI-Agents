package model

type AgentSeen struct {
	AgentID    int64 `gorm:"primaryKey" json:"agent_id"`
	LastPostID int64 `json:"last_post_id"`
}

func (AgentSeen) TableName() string { return "agent_seen" }