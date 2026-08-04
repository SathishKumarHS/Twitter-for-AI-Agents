// Package agents defines the Agent model and its storage contract.
package agents

import (
	"context"

	"gorm.io/gorm"

	dberrors "github.com/agent-feed/backend-go/internal/db/db_errors"
)

// Agent is the Registry definition of one AI agent.
type Agent struct {
	ID           int64    `gorm:"primaryKey" json:"id"`
	Handle       string   `gorm:"uniqueIndex;size:64" json:"handle"`
	Name         string   `gorm:"size:120" json:"name"`
	Description  string   `gorm:"type:text" json:"description"`
	Instructions string   `gorm:"type:text" json:"instructions"`
	Persona      string   `gorm:"type:text" json:"persona"`
	Goals        []string `gorm:"serializer:json" json:"goals"`
	Policies     []string `gorm:"serializer:json" json:"policies"`
	Interests    []string `gorm:"serializer:json" json:"interests"`
	AvatarURL    string   `gorm:"size:500" json:"avatar_url"`
	Active       bool     `json:"active"`
	CreatedAt    string   `json:"created_at"`
}

// TableName gives GORM the Postgres table name.
func (Agent) TableName() string { return "agents" }

// Store is the storage contract for agents. Consumers depend on this
// interface, not on a concrete database implementation.
type Store interface {
	ListAll(ctx context.Context) ([]Agent, error)
	ListActive(ctx context.Context) ([]Agent, error)
	Get(ctx context.Context, id int64) (*Agent, error)
	GetByHandle(ctx context.Context, handle string) (*Agent, error)
	Create(ctx context.Context, a Agent) (int64, error)
	SetActive(ctx context.Context, id int64, active bool) error
}

// GormStore is the GORM-backed implementation of Store.
type GormStore struct {
	g *gorm.DB
}

// NewGorm returns a Store backed by GORM.
func NewGorm(g *gorm.DB) Store {
	return &GormStore{g: g}
}

func (s *GormStore) ListAll(ctx context.Context) ([]Agent, error) {
	var out []Agent
	if err := s.g.WithContext(ctx).Order("id").Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (s *GormStore) ListActive(ctx context.Context) ([]Agent, error) {
	var out []Agent
	if err := s.g.WithContext(ctx).Where("active = ?", true).Order("id").Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (s *GormStore) Get(ctx context.Context, id int64) (*Agent, error) {
	var a Agent
	if err := s.g.WithContext(ctx).First(&a, id).Error; err != nil {
		if dberrors.IsNoRows(err) {
			return nil, dberrors.ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (s *GormStore) GetByHandle(ctx context.Context, handle string) (*Agent, error) {
	var a Agent
	if err := s.g.WithContext(ctx).Where("LOWER(handle) = ?", handle).First(&a).Error; err != nil {
		if dberrors.IsNoRows(err) {
			return nil, dberrors.ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (s *GormStore) Create(ctx context.Context, a Agent) (int64, error) {
	if err := s.g.WithContext(ctx).Create(&a).Error; err != nil {
		return 0, err
	}
	return a.ID, nil
}

func (s *GormStore) SetActive(ctx context.Context, id int64, active bool) error {
	return s.g.WithContext(ctx).Model(&Agent{}).Where("id = ?", id).Update("active", active).Error
}
