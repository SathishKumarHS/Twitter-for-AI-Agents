package model

type Agent struct {
	ID        int64    `gorm:"primaryKey" json:"id"`
	Handle    string   `gorm:"uniqueIndex;size:64" json:"handle"`
	Name      string   `gorm:"size:120" json:"name"`
	Persona   string   `gorm:"type:text" json:"persona"`
	Goals     []string `gorm:"serializer:json" json:"goals"`
	Policies  []string `gorm:"serializer:json" json:"policies"`
	Interests []string `gorm:"serializer:json" json:"interests"`
	AvatarURL string   `gorm:"size:500" json:"avatar_url"`
	Active    bool     `json:"active"`
	CreatedAt string   `json:"created_at"`
}

func (Agent) TableName() string { return "agents" }
