package service

import (
	"iot-platform/internal/billing/model"
	"iot-platform/internal/billing/repository"
	errs "iot-platform/pkg/errors"

	"github.com/google/uuid"
)

// BillingService 收费方案业务逻辑层
type BillingService struct {
	repo *repository.BillingRepository
}

func NewBillingService(repo *repository.BillingRepository) *BillingService {
	return &BillingService{repo: repo}
}

// ========== BillingScheme ==========

func (s *BillingService) CreateScheme(name, schemeType, deviceType, siteID string,
	baseServiceFee, maxPrice, unitPrice float64, unit string) (*model.BillingScheme, error) {
	scheme := &model.BillingScheme{
		ID:             uuid.New().String(),
		Name:           name,
		Type:           schemeType,
		DeviceType:     deviceType,
		SiteID:         siteID,
		BaseServiceFee: baseServiceFee,
		MaxPrice:       maxPrice,
		UnitPrice:      unitPrice,
		Unit:           unit,
		Status:         "active",
	}
	if err := s.repo.CreateScheme(scheme); err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}
	return scheme, nil
}

func (s *BillingService) GetScheme(id string) (*model.BillingScheme, error) {
	scheme, err := s.repo.GetScheme(id)
	if err != nil {
		return nil, errs.NotFound("收费方案", id)
	}
	return scheme, nil
}

func (s *BillingService) ListSchemes(deviceType, siteID string, page, pageSize int) ([]model.BillingScheme, int64, error) {
	return s.repo.ListSchemes(deviceType, siteID, page, pageSize)
}

func (s *BillingService) UpdateScheme(id string, req map[string]interface{}) (*model.BillingScheme, error) {
	scheme, err := s.repo.GetScheme(id)
	if err != nil {
		return nil, errs.NotFound("收费方案", id)
	}
	if v, ok := req["name"]; ok {
		scheme.Name = v.(string)
	}
	if v, ok := req["base_service_fee"]; ok {
		scheme.BaseServiceFee = v.(float64)
	}
	if v, ok := req["max_price"]; ok {
		scheme.MaxPrice = v.(float64)
	}
	if v, ok := req["unit_price"]; ok {
		scheme.UnitPrice = v.(float64)
	}
	if v, ok := req["unit"]; ok {
		scheme.Unit = v.(string)
	}
	if v, ok := req["status"]; ok {
		scheme.Status = v.(string)
	}
	if err := s.repo.UpdateScheme(scheme); err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}
	return scheme, nil
}

func (s *BillingService) DeleteScheme(id string) error {
	if _, err := s.repo.GetScheme(id); err != nil {
		return errs.NotFound("收费方案", id)
	}
	return s.repo.DeleteScheme(id)
}

// ========== BillingPeriod ==========

func (s *BillingService) AddPeriod(schemeID, periodType, startTime, endTime string,
	pricePerKWh, serviceFee float64, sortOrder int) (*model.BillingPeriod, error) {
	if _, err := s.repo.GetScheme(schemeID); err != nil {
		return nil, errs.NotFound("收费方案", schemeID)
	}
	period := &model.BillingPeriod{
		SchemeID:    schemeID,
		PeriodType:  periodType,
		StartTime:   startTime,
		EndTime:     endTime,
		PricePerKWh: pricePerKWh,
		ServiceFee:  serviceFee,
		SortOrder:   sortOrder,
	}
	if err := s.repo.CreatePeriod(period); err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}
	return period, nil
}

func (s *BillingService) GetPeriods(schemeID string) ([]model.BillingPeriod, error) {
	return s.repo.GetPeriodsByScheme(schemeID)
}

func (s *BillingService) UpdatePeriod(id uint, req map[string]interface{}) error {
	// Get all periods to find the one (repo design limitation)
	// For now, let's just fetch and update
	periods, err := s.repo.GetPeriodsByScheme("")
	if err != nil {
		return errs.ErrInternalServer.Wrap(err)
	}
	var target *model.BillingPeriod
	for i := range periods {
		if periods[i].ID == id {
			target = &periods[i]
			break
		}
	}
	if target == nil {
		return errs.NotFound("时段配置", "")
	}
	if v, ok := req["period_type"]; ok {
		target.PeriodType = v.(string)
	}
	if v, ok := req["start_time"]; ok {
		target.StartTime = v.(string)
	}
	if v, ok := req["end_time"]; ok {
		target.EndTime = v.(string)
	}
	if v, ok := req["price_per_kwh"]; ok {
		target.PricePerKWh = v.(float64)
	}
	if v, ok := req["service_fee"]; ok {
		target.ServiceFee = v.(float64)
	}
	return s.repo.UpdatePeriod(target)
}

func (s *BillingService) DeletePeriod(id uint) error {
	return s.repo.DeletePeriod(id)
}

func (s *BillingService) BatchSetPeriods(schemeID string, periods []model.BillingPeriod) error {
	if _, err := s.repo.GetScheme(schemeID); err != nil {
		return errs.NotFound("收费方案", schemeID)
	}
	// 删除旧时段，批量创建新时段
	if err := s.repo.BatchDeletePeriods(schemeID); err != nil {
		return errs.ErrInternalServer.Wrap(err)
	}
	for i := range periods {
		periods[i].SchemeID = schemeID
		if err := s.repo.CreatePeriod(&periods[i]); err != nil {
			return errs.ErrInternalServer.Wrap(err)
		}
	}
	return nil
}

// ========== MonthlyCardScheme ==========

func (s *BillingService) CreateMonthlyScheme(name, deviceType string, price float64, durationDays int) (*model.MonthlyCardScheme, error) {
	scheme := &model.MonthlyCardScheme{
		ID:           uuid.New().String(),
		Name:         name,
		DeviceType:   deviceType,
		Price:        price,
		DurationDays: durationDays,
		Status:       "active",
	}
	if err := s.repo.CreateMonthlyScheme(scheme); err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}
	return scheme, nil
}

func (s *BillingService) ListMonthlySchemes(deviceType string, page, pageSize int) ([]model.MonthlyCardScheme, int64, error) {
	return s.repo.ListMonthlySchemes(deviceType, page, pageSize)
}

func (s *BillingService) UpdateMonthlyScheme(id string, req map[string]interface{}) (*model.MonthlyCardScheme, error) {
	scheme, err := s.repo.GetMonthlyScheme(id)
	if err != nil {
		return nil, errs.NotFound("月卡方案", id)
	}
	if v, ok := req["name"]; ok {
		scheme.Name = v.(string)
	}
	if v, ok := req["price"]; ok {
		scheme.Price = v.(float64)
	}
	if v, ok := req["duration_days"]; ok {
		scheme.DurationDays = int(v.(float64))
	}
	if v, ok := req["status"]; ok {
		scheme.Status = v.(string)
	}
	if err := s.repo.UpdateMonthlyScheme(scheme); err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}
	return scheme, nil
}

func (s *BillingService) DeleteMonthlyScheme(id string) error {
	if _, err := s.repo.GetMonthlyScheme(id); err != nil {
		return errs.NotFound("月卡方案", id)
	}
	return s.repo.DeleteMonthlyScheme(id)
}

// ========== RechargeScheme ==========

func (s *BillingService) CreateRechargeScheme(name string, amount, bonusAmount float64) (*model.RechargeScheme, error) {
	scheme := &model.RechargeScheme{
		ID:          uuid.New().String(),
		Name:        name,
		Amount:      amount,
		BonusAmount: bonusAmount,
		Status:      "active",
	}
	if err := s.repo.CreateRechargeScheme(scheme); err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}
	return scheme, nil
}

func (s *BillingService) ListRechargeSchemes(page, pageSize int) ([]model.RechargeScheme, int64, error) {
	return s.repo.ListRechargeSchemes(page, pageSize)
}

func (s *BillingService) UpdateRechargeScheme(id string, req map[string]interface{}) (*model.RechargeScheme, error) {
	scheme, err := s.repo.GetRechargeScheme(id)
	if err != nil {
		return nil, errs.NotFound("充值方案", id)
	}
	if v, ok := req["name"]; ok {
		scheme.Name = v.(string)
	}
	if v, ok := req["amount"]; ok {
		scheme.Amount = v.(float64)
	}
	if v, ok := req["bonus_amount"]; ok {
		scheme.BonusAmount = v.(float64)
	}
	if v, ok := req["status"]; ok {
		scheme.Status = v.(string)
	}
	if err := s.repo.UpdateRechargeScheme(scheme); err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}
	return scheme, nil
}

func (s *BillingService) DeleteRechargeScheme(id string) error {
	if _, err := s.repo.GetRechargeScheme(id); err != nil {
		return errs.NotFound("充值方案", id)
	}
	return s.repo.DeleteRechargeScheme(id)
}

// ========== BusinessConfig ==========

func (s *BillingService) GetConfig(key string) (*model.BusinessConfig, error) {
	cfg, err := s.repo.GetConfig(key)
	if err != nil {
		return nil, errs.NotFound("业务配置", key)
	}
	return cfg, nil
}

func (s *BillingService) SetConfig(key, value, desc string) error {
	return s.repo.SetConfig(key, value, desc)
}

func (s *BillingService) ListConfigs() ([]model.BusinessConfig, error) {
	return s.repo.ListConfigs()
}
