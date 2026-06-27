package service

import (
	"time"

	"iot-platform/internal/advertisement/model"
	"iot-platform/internal/advertisement/repository"
	errs "iot-platform/pkg/errors"

	"github.com/google/uuid"
)

// AdvertisementService 广告/运营业务逻辑层
type AdvertisementService struct {
	repo *repository.AdvertisementRepository
}

func NewAdvertisementService(repo *repository.AdvertisementRepository) *AdvertisementService {
	return &AdvertisementService{repo: repo}
}

// ========== Advertisement ==========

func (s *AdvertisementService) CreateAd(title, imageURL, linkURL, platform string, sortOrder int) (*model.Advertisement, error) {
	ad := &model.Advertisement{
		ID:        uuid.New().String(),
		Title:     title,
		ImageURL:  imageURL,
		LinkURL:   linkURL,
		Platform:  platform,
		SortOrder: sortOrder,
		Status:    "active",
	}
	if err := s.repo.Create(ad); err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}
	return ad, nil
}

func (s *AdvertisementService) ListAds(platform string, page, pageSize int) ([]model.Advertisement, int64, error) {
	return s.repo.List(platform, page, pageSize)
}

func (s *AdvertisementService) UpdateAd(id string, req map[string]interface{}) error {
	if _, err := s.repo.GetByID(id); err != nil {
		return errs.NotFound("广告", id)
	}
	return s.repo.Update(id, req)
}

func (s *AdvertisementService) DeleteAd(id string) error {
	if _, err := s.repo.GetByID(id); err != nil {
		return errs.NotFound("广告", id)
	}
	return s.repo.Delete(id)
}

// ========== FranchiseApplication ==========

func (s *AdvertisementService) ApplyFranchise(name, phone, company, address, remark string) (*model.FranchiseApplication, error) {
	app := &model.FranchiseApplication{
		ID:      uuid.New().String(),
		Name:    name,
		Phone:   phone,
		Company: company,
		Address: address,
		Remark:  remark,
		Status:  "pending",
	}
	if err := s.repo.CreateFranchise(app); err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}
	return app, nil
}

func (s *AdvertisementService) ListFranchises(status string, page, pageSize int) ([]model.FranchiseApplication, int64, error) {
	return s.repo.ListFranchises(status, page, pageSize)
}

func (s *AdvertisementService) ProcessFranchise(id, status, processedBy string) error {
	app, err := s.repo.GetFranchise(id)
	if err != nil {
		return errs.NotFound("加盟申请", id)
	}
	if app.Status != "pending" {
		return errs.New("APPLICATION_ALREADY_PROCESSED", "申请已处理", 400)
	}
	now := time.Now()
	return s.repo.UpdateFranchise(id, map[string]interface{}{
		"status":       status,
		"processed_by":  processedBy,
		"processed_at":  now,
	})
}

// ========== WechatUser ==========

func (s *AdvertisementService) GetOrCreateWechatUser(openID string) (*model.WechatUser, error) {
	user, err := s.repo.GetWechatUserByOpenID(openID)
	if err == nil {
		return user, nil
	}
	user = &model.WechatUser{
		OpenID:  openID,
		Enabled: true,
	}
	if err := s.repo.CreateWechatUser(user); err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}
	return user, nil
}

func (s *AdvertisementService) ListWechatUsers(page, pageSize int) ([]model.WechatUser, int64, error) {
	return s.repo.ListWechatUsers(page, pageSize)
}

func (s *AdvertisementService) UpdateWechatUser(id uint, req map[string]interface{}) error {
	if _, err := s.repo.GetWechatUserByID(id); err != nil {
		return errs.NotFound("微信用户", "")
	}
	return s.repo.UpdateWechatUser(id, req)
}

func (s *AdvertisementService) FreezeWechatUser(id uint) error {
	return s.repo.UpdateWechatUser(id, map[string]interface{}{"enabled": false})
}
