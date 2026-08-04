// Package agents defines the Agent model and its storage contract.
package agents

import (
	"context"

	"gorm.io/gorm"

	dberrors "github.com/agent-feed/backend-go/internal/db/db_errors"
	"github.com/agent-feed/backend-go/internal/db/model"
)

type AgentsDB struct {
	g *gorm.DB
}

func NewAgentsDB(g *gorm.DB) AgentsInterface {
	return &AgentsDB{g: g}
}

func (s *AgentsDB) ListAll(ctx context.Context) ([]model.Agent, error) {
	var out []model.Agent
	if err := s.g.WithContext(ctx).Order("id").Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (s *AgentsDB) ListActive(ctx context.Context) ([]model.Agent, error) {
	var out []model.Agent
	if err := s.g.WithContext(ctx).Where("active = ?", true).Order("id").Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

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

func (s *AgentsDB) GetByHandle(ctx context.Context, handle string) (*model.Agent, error) {
	var a model.Agent
	if err := s.g.WithContext(ctx).Where("LOWER(handle) = ?", handle).First(&a).Error; err != nil {
		if dberrors.IsNoRows(err) {
			return nil, dberrors.ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (s *AgentsDB) Create(ctx context.Context, a model.Agent) (int64, error) {
	if err := s.g.WithContext(ctx).Create(&a).Error; err != nil {
		return 0, err
	}
	return a.ID, nil
}

func (s *AgentsDB) SetActive(ctx context.Context, id int64, active bool) error {
	return s.g.WithContext(ctx).Model(&model.Agent{}).Where("id = ?", id).Update("active", active).Error
}
