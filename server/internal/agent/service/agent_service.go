package service

import (
	"iot-platform/internal/agent/model"
	"iot-platform/internal/agent/repository"
	errs "iot-platform/pkg/errors"

	"github.com/google/uuid"
)

// AgentService 代理/运营商业务逻辑层
type AgentService struct {
	repo *repository.AgentRepository
}

func NewAgentService(repo *repository.AgentRepository) *AgentService {
	return &AgentService{repo: repo}
}

// ========== Agent ==========

func (s *AgentService) CreateAgent(name, contact, phone, email, address string, commissionRate float64) (*model.Agent, error) {
	agent := &model.Agent{
		ID:             uuid.New().String(),
		Name:           name,
		Contact:        contact,
		Phone:          phone,
		Email:          email,
		Address:        address,
		CommissionRate: commissionRate,
		Status:         "active",
	}
	if err := s.repo.Create(agent); err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}
	return agent, nil
}

func (s *AgentService) GetAgent(id string) (*model.Agent, error) {
	agent, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errs.NotFound("代理商", id)
	}
	return agent, nil
}

func (s *AgentService) ListAgents(page, pageSize int) ([]model.Agent, int64, error) {
	return s.repo.List(page, pageSize)
}

func (s *AgentService) UpdateAgent(id string, req map[string]interface{}) (*model.Agent, error) {
	if _, err := s.repo.GetByID(id); err != nil {
		return nil, errs.NotFound("代理商", id)
	}
	updates := make(map[string]interface{})
	for k, v := range req {
		updates[k] = v
	}
	if err := s.repo.Update(id, updates); err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}
	return s.repo.GetByID(id)
}

func (s *AgentService) DeleteAgent(id string) error {
	if _, err := s.repo.GetByID(id); err != nil {
		return errs.NotFound("代理商", id)
	}
	return s.repo.Delete(id)
}

// ========== Operator ==========

func (s *AgentService) CreateOperator(agentID, name, contact, phone, address string) (*model.Operator, error) {
	// 验证代理商存在
	if agentID != "" {
		if _, err := s.repo.GetByID(agentID); err != nil {
			return nil, errs.NotFound("代理商", agentID)
		}
	}
	op := &model.Operator{
		ID:      uuid.New().String(),
		AgentID: agentID,
		Name:    name,
		Contact: contact,
		Phone:   phone,
		Address: address,
		Status:  "active",
	}
	if err := s.repo.CreateOperator(op); err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}
	return op, nil
}

func (s *AgentService) GetOperator(id string) (*model.Operator, error) {
	op, err := s.repo.GetOperatorByID(id)
	if err != nil {
		return nil, errs.NotFound("运营商", id)
	}
	return op, nil
}

func (s *AgentService) ListOperators(agentID string, page, pageSize int) ([]model.Operator, int64, error) {
	return s.repo.ListOperators(agentID, page, pageSize)
}

func (s *AgentService) UpdateOperator(id string, req map[string]interface{}) (*model.Operator, error) {
	if _, err := s.repo.GetOperatorByID(id); err != nil {
		return nil, errs.NotFound("运营商", id)
	}
	if err := s.repo.UpdateOperator(id, req); err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}
	return s.repo.GetOperatorByID(id)
}

func (s *AgentService) DeleteOperator(id string) error {
	if _, err := s.repo.GetOperatorByID(id); err != nil {
		return errs.NotFound("运营商", id)
	}
	return s.repo.DeleteOperator(id)
}
