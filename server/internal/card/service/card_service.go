package service

import (
	"time"

	"iot-platform/internal/card/model"
	"iot-platform/internal/card/repository"
	errs "iot-platform/pkg/errors"

	"github.com/google/uuid"
)

// CardService 卡片管理业务逻辑层
type CardService struct {
	repo *repository.CardRepository
}

func NewCardService(repo *repository.CardRepository) *CardService {
	return &CardService{repo: repo}
}

// ========== IC Card ==========

func (s *CardService) CreateICCard(cardNo, cardUID string) (*model.ICCard, error) {
	if _, err := s.repo.GetICCardByNo(cardNo); err == nil {
		return nil, errs.New("CARD_EXISTS", "卡号已存在", 409)
	}
	card := &model.ICCard{
		ID:     uuid.New().String(),
		CardNo: cardNo,
		CardUID: cardUID,
		Status: model.CardStatusActive,
	}
	if err := s.repo.CreateICCard(card); err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}
	return card, nil
}

func (s *CardService) GetICCard(id string) (*model.ICCard, error) {
	card, err := s.repo.GetICCard(id)
	if err != nil {
		return nil, errs.NotFound("IC卡", id)
	}
	return card, nil
}

func (s *CardService) ListICCards(page, pageSize int) ([]model.ICCard, int64, error) {
	return s.repo.ListICCards(page, pageSize)
}

func (s *CardService) RechargeICCard(cardID string, amount float64, createdBy string) error {
	card, err := s.repo.GetICCard(cardID)
	if err != nil {
		return errs.NotFound("IC卡", cardID)
	}
	if card.Status != model.CardStatusActive {
		return errs.New("CARD_NOT_ACTIVE", "卡状态异常", 400)
	}

	// 更新卡内余额
	if err := s.repo.UpdateICCard(cardID, map[string]interface{}{
		"balance":        amount + card.Balance,
		"total_recharge": amount + card.TotalRecharge,
	}); err != nil {
		return errs.ErrInternalServer.Wrap(err)
	}

	// 记录充值
	record := &model.ICCardRecharge{
		ID:     uuid.New().String(),
		CardID: cardID,
		Amount: amount,
		CreatedBy: createdBy,
	}
	return s.repo.CreateICCardRecharge(record)
}

func (s *CardService) BindICCardToUser(cardID string, userID uint) error {
	if _, err := s.repo.GetICCard(cardID); err != nil {
		return errs.NotFound("IC卡", cardID)
	}
	return s.repo.UpdateICCard(cardID, map[string]interface{}{"user_id": userID})
}

func (s *CardService) ReportLostICCard(cardID string) error {
	if _, err := s.repo.GetICCard(cardID); err != nil {
		return errs.NotFound("IC卡", cardID)
	}
	return s.repo.UpdateICCard(cardID, map[string]interface{}{"status": model.CardStatusLost})
}

func (s *CardService) DeleteICCard(id string) error {
	if _, err := s.repo.GetICCard(id); err != nil {
		return errs.NotFound("IC卡", id)
	}
	return s.repo.DeleteICCard(id)
}

func (s *CardService) BatchImportICCards(cardNos []string) (int, error) {
	cards := make([]*model.ICCard, 0, len(cardNos))
	for _, no := range cardNos {
		if _, err := s.repo.GetICCardByNo(no); err == nil {
			continue // 跳过已存在
		}
		cards = append(cards, &model.ICCard{
			ID:     uuid.New().String(),
			CardNo: no,
			Status: model.CardStatusActive,
		})
	}
	if len(cards) == 0 {
		return 0, nil
	}
	if err := s.repo.BatchCreateICCards(cards); err != nil {
		return 0, errs.ErrInternalServer.Wrap(err)
	}
	return len(cards), nil
}

// ========== Traffic Card ==========

func (s *CardService) CreateTrafficCard(iccid, imsi, carrier string) (*model.TrafficCard, error) {
	card := &model.TrafficCard{
		ID:      uuid.New().String(),
		ICCID:   iccid,
		IMSI:    imsi,
		Carrier: carrier,
		Status:  model.CardStatusActive,
	}
	if err := s.repo.CreateTrafficCard(card); err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}
	return card, nil
}

func (s *CardService) ListTrafficCards(page, pageSize int) ([]model.TrafficCard, int64, error) {
	return s.repo.ListTrafficCards(page, pageSize)
}

func (s *CardService) BindTrafficCardToDevice(cardID, deviceID string) error {
	if _, err := s.repo.GetTrafficCard(cardID); err != nil {
		return errs.NotFound("流量卡", cardID)
	}
	return s.repo.UpdateTrafficCard(cardID, map[string]interface{}{"device_id": deviceID})
}

// ========== Monthly Card ==========

func (s *CardService) IssueMonthlyCard(userID uint, schemeID, schemeName, deviceType string, durationDays int) (*model.MonthlyCard, error) {
	now := time.Now()
	endDate := now.AddDate(0, 0, durationDays)

	card := &model.MonthlyCard{
		ID:         uuid.New().String(),
		UserID:     userID,
		SchemeID:   schemeID,
		SchemeName: schemeName,
		DeviceType: deviceType,
		StartDate:  now,
		EndDate:    endDate,
		Status:     "active",
	}
	if err := s.repo.CreateMonthlyCard(card); err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}
	return card, nil
}

func (s *CardService) ListMonthlyCards(userID *uint, page, pageSize int) ([]model.MonthlyCard, int64, error) {
	return s.repo.ListMonthlyCards(userID, page, pageSize)
}

func (s *CardService) LogMonthlyOperation(cardID, operation, remark, operatorBy string) error {
	record := &model.MonthlyCardRecord{
		CardID:    cardID,
		Operation: operation,
		Remark:    remark,
		OperatorBy: operatorBy,
	}
	return s.repo.CreateMonthlyRecord(record)
}
