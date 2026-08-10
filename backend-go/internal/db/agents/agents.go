// Package agents defines the Agent database store implementation.
package agents

import (
	"context"

	"gorm.io/gorm"

	dberrors "github.com/twitter-for-ai-agents/backend-go/internal/db/db_errors"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
)

// GORM-backed storage implementation for Agent records
type AgentsDB struct {
	g *gorm.DB
}

// Instantiates a new AgentsDB store implementing AgentsInterface
func NewAgentsDB(g *gorm.DB) AgentsInterface {
	return &AgentsDB{g: g}
}

// Fetches all agents from the database ordered by ID
func (s *AgentsDB) ListAll(ctx context.Context) ([]model.Agent, error) {
	var out []model.Agent
	if err := s.g.WithContext(ctx).Order("id").Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// Fetches all active agents from the database ordered by ID
func (s *AgentsDB) ListActive(ctx context.Context) ([]model.Agent, error) {
	var out []model.Agent
	if err := s.g.WithContext(ctx).Where("active = ?", true).Order("id").Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// Fetches a single agent by primary key ID
func (s *AgentsDB) Get(ctx context.Context, id int64) (*model.Agent, error) {
	var a model.Agent
	if err := s.g.WithContext(ctx).First(&a, id).Error; err != nil {
		if dberrors.IsNoRows(err) {
			return nil, dberrors.ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}

// Fetches a single agent by handle (case-insensitive)
func (s *AgentsDB) GetByHandle(ctx context.Context, handle string) (*model.Agent, error) {
	var a model.Agent
	if err := s.g.WithContext(ctx).Where("LOWER(handle) = LOWER(?)", handle).First(&a).Error; err != nil {
		if dberrors.IsNoRows(err) {
			return nil, dberrors.ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}

// Inserts a new agent into the database and returns its assigned ID
func (s *AgentsDB) Create(ctx context.Context, a model.Agent) (int64, error) {
	if err := s.g.WithContext(ctx).Create(&a).Error; err != nil {
		return 0, err
	}
	return a.ID, nil
}

// Updates the active status (true/false) for a specified agent ID
func (s *AgentsDB) SetActive(ctx context.Context, id int64, active bool) error {
	return s.g.WithContext(ctx).Model(&model.Agent{}).Where("id = ?", id).Update("active", active).Error
}
