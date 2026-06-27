package service

import (
	"iot-platform/internal/interconnect/model"
	"iot-platform/internal/interconnect/repository"
	errs "iot-platform/pkg/errors"

	"github.com/google/uuid"
)

// InterconnectService 互联互通业务逻辑层
type InterconnectService struct {
	repo *repository.InterconnectRepository
}

func NewInterconnectService(repo *repository.InterconnectRepository) *InterconnectService {
	return &InterconnectService{repo: repo}
}

func (s *InterconnectService) CreateOrg(name, orgCode, contact, phone, pushURL, reconciliationURL string) (*model.InterconnectOrg, error) {
	org := &model.InterconnectOrg{
		ID:                uuid.New().String(),
		Name:              name,
		OrgCode:           orgCode,
		Contact:           contact,
		Phone:             phone,
		PushURL:           pushURL,
		ReconciliationURL: reconciliationURL,
		Status:            "active",
	}
	if err := s.repo.CreateOrg(org); err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}
	return org, nil
}

func (s *InterconnectService) GetOrg(id string) (*model.InterconnectOrg, error) {
	org, err := s.repo.GetOrg(id)
	if err != nil {
		return nil, errs.NotFound("互联机构", id)
	}
	return org, nil
}

func (s *InterconnectService) ListOrgs(page, pageSize int) ([]model.InterconnectOrg, int64, error) {
	return s.repo.ListOrgs(page, pageSize)
}

func (s *InterconnectService) UpdateOrg(id string, req map[string]interface{}) error {
	if _, err := s.repo.GetOrg(id); err != nil {
		return errs.NotFound("互联机构", id)
	}
	return s.repo.UpdateOrg(id, req)
}

func (s *InterconnectService) DeleteOrg(id string) error {
	if _, err := s.repo.GetOrg(id); err != nil {
		return errs.NotFound("互联机构", id)
	}
	return s.repo.DeleteOrg(id)
}

// ========== InterconnectKey ==========

func (s *InterconnectService) CreateKey(orgID, keyType, publicKey, privateKey, secretKey, remark string) (*model.InterconnectKey, error) {
	if _, err := s.repo.GetOrg(orgID); err != nil {
		return nil, errs.NotFound("互联机构", orgID)
	}
	key := &model.InterconnectKey{
		ID:         uuid.New().String(),
		OrgID:      orgID,
		KeyType:    keyType,
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		SecretKey:  secretKey,
		Remark:     remark,
		Status:     "active",
	}
	if err := s.repo.CreateKey(key); err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}
	return key, nil
}

func (s *InterconnectService) ListKeys(orgID string, page, pageSize int) ([]model.InterconnectKey, int64, error) {
	return s.repo.ListKeys(orgID, page, pageSize)
}

func (s *InterconnectService) DeleteKey(id string) error {
	if _, err := s.repo.GetKey(id); err != nil {
		return errs.NotFound("互联密钥", id)
	}
	return s.repo.DeleteKey(id)
}
