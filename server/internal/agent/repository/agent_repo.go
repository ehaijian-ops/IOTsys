package repository

import (
	"iot-platform/internal/agent/model"

	"gorm.io/gorm"
)

// AgentRepository 代理/运营商数据访问层
type AgentRepository struct {
	db *gorm.DB
}

func NewAgentRepository(db *gorm.DB) *AgentRepository {
	return &AgentRepository{db: db}
}

// ========== Agent ==========

func (r *AgentRepository) Create(agent *model.Agent) error {
	return r.db.Create(agent).Error
}

func (r *AgentRepository) GetByID(id string) (*model.Agent, error) {
	var a model.Agent
	err := r.db.First(&a, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AgentRepository) List(page, pageSize int) ([]model.Agent, int64, error) {
	var agents []model.Agent
	var total int64
	if err := r.db.Model(&model.Agent{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := r.db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&agents).Error
	return agents, total, err
}

func (r *AgentRepository) Update(id string, updates map[string]interface{}) error {
	return r.db.Model(&model.Agent{}).Where("id = ?", id).Updates(updates).Error
}

func (r *AgentRepository) Delete(id string) error {
	return r.db.Delete(&model.Agent{}, "id = ?", id).Error
}

func (r *AgentRepository) UpdateBalance(id string, amount float64) error {
	return r.db.Model(&model.Agent{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"balance":       gorm.Expr("balance + ?", amount),
			"total_revenue": gorm.Expr("total_revenue + ?", amount),
		}).Error
}

// ========== Operator ==========

func (r *AgentRepository) CreateOperator(op *model.Operator) error {
	return r.db.Create(op).Error
}

func (r *AgentRepository) GetOperatorByID(id string) (*model.Operator, error) {
	var o model.Operator
	err := r.db.First(&o, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *AgentRepository) ListOperators(agentID string, page, pageSize int) ([]model.Operator, int64, error) {
	var ops []model.Operator
	var total int64
	q := r.db.Model(&model.Operator{})
	if agentID != "" {
		q = q.Where("agent_id = ?", agentID)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&ops).Error
	return ops, total, err
}

func (r *AgentRepository) UpdateOperator(id string, updates map[string]interface{}) error {
	return r.db.Model(&model.Operator{}).Where("id = ?", id).Updates(updates).Error
}

func (r *AgentRepository) DeleteOperator(id string) error {
	return r.db.Delete(&model.Operator{}, "id = ?", id).Error
}
