package service

import (
	"time"

	"iot-platform/internal/finance/model"
	"iot-platform/internal/finance/repository"
	errs "iot-platform/pkg/errors"

	"github.com/google/uuid"
)

// FinanceService 财务业务逻辑层
type FinanceService struct {
	repo *repository.FinanceRepository
}

func NewFinanceService(repo *repository.FinanceRepository) *FinanceService {
	return &FinanceService{repo: repo}
}

// ========== Wallet ==========

func (s *FinanceService) GetOrCreateWallet(userID uint) (*model.UserWallet, error) {
	wallet, err := s.repo.GetWallet(userID)
	if err == nil {
		return wallet, nil
	}
	wallet = &model.UserWallet{
		UserID: userID,
	}
	if err := s.repo.CreateWallet(wallet); err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}
	return wallet, nil
}

func (s *FinanceService) GetWallet(userID uint) (*model.UserWallet, error) {
	wallet, err := s.repo.GetWallet(userID)
	if err != nil {
		return nil, errs.NotFound("钱包", "")
	}
	return wallet, nil
}

// ========== Recharge ==========

func (s *FinanceService) Recharge(userID uint, amount, bonusAmount float64, payMethod, tradeNo string) (*model.RechargeRecord, error) {
	// 创建充值记录
	record := &model.RechargeRecord{
		ID:          uuid.New().String(),
		UserID:      userID,
		Amount:      amount,
		BonusAmount: bonusAmount,
		PayMethod:   payMethod,
		TradeNo:     tradeNo,
		Status:      model.RechargeSuccess,
	}

	if err := s.repo.CreateRecharge(record); err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}

	// 确保钱包存在
	s.GetOrCreateWallet(userID)

	// 更新钱包余额（本金+赠送）
	totalAmount := amount + bonusAmount
	if err := s.repo.UpdateWalletBalance(userID, totalAmount); err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}

	return record, nil
}

func (s *FinanceService) AdminRecharge(userID uint, amount float64, remark string) (*model.RechargeRecord, error) {
	return s.Recharge(userID, amount, 0, "admin", remark)
}

func (s *FinanceService) ListRecharges(userID *uint, status string, page, pageSize int) ([]model.RechargeRecord, int64, error) {
	return s.repo.ListRecharges(userID, status, page, pageSize)
}

// ========== Withdraw ==========

func (s *FinanceService) ApplyWithdraw(userID uint, amount float64, bankName, bankCardNo, bankAccount string) (*model.WithdrawRecord, error) {
	wallet, err := s.repo.GetWallet(userID)
	if err != nil {
		return nil, errs.NotFound("钱包", "")
	}
	if wallet.Balance < amount {
		return nil, errs.New("BALANCE_INSUFFICIENT", "余额不足", 400)
	}

	// 冻结提现金额
	if err := s.repo.FreezeBalance(userID, amount); err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}

	record := &model.WithdrawRecord{
		ID:          uuid.New().String(),
		Amount:      amount,
		BankName:    bankName,
		BankCardNo:  bankCardNo,
		BankAccount: bankAccount,
		Status:      model.WithdrawPending,
	}
	record.UserID = &userID

	if err := s.repo.CreateWithdraw(record); err != nil {
		_ = s.repo.UnfreezeBalance(userID, amount)
		return nil, errs.ErrInternalServer.Wrap(err)
	}

	return record, nil
}

func (s *FinanceService) ProcessWithdraw(id, status string, actualAmount float64, remark, processedBy string) error {
	record, err := s.repo.GetWithdraw(id)
	if err != nil {
		return errs.NotFound("提现记录", id)
	}
	if record.Status != model.WithdrawPending {
		return errs.New("WITHDRAW_STATUS_ERROR", "提现状态不正确", 400)
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":        status,
		"processed_at":  now,
		"processed_by":  processedBy,
	}
	if actualAmount > 0 {
		updates["actual_amount"] = actualAmount
	}
	if remark != "" {
		updates["remark"] = remark
	}

	if err := s.repo.UpdateWithdraw(id, updates); err != nil {
		return errs.ErrInternalServer.Wrap(err)
	}

	// 如果驳回，退回冻结金额
	if status == model.WithdrawRejected && record.UserID != nil {
		_ = s.repo.UnfreezeBalance(*record.UserID, record.Amount)
	}

	return nil
}

func (s *FinanceService) ListWithdraws(status string, page, pageSize int) ([]model.WithdrawRecord, int64, error) {
	return s.repo.ListWithdraws(status, page, pageSize)
}

// ========== RevenueSplit ==========

func (s *FinanceService) CreateSplit(orderID string, totalAmount, agentRate, operatorRate float64, agentID, operatorID *string) (*model.RevenueSplit, error) {
	agentAmount := totalAmount * agentRate
	operatorAmount := totalAmount * operatorRate
	platformAmount := totalAmount - agentAmount - operatorAmount

	split := &model.RevenueSplit{
		ID:             uuid.New().String(),
		OrderID:        orderID,
		TotalAmount:    totalAmount,
		AgentID:        agentID,
		AgentAmount:    agentAmount,
		AgentRate:      agentRate,
		OperatorID:     operatorID,
		OperatorAmount: operatorAmount,
		OperatorRate:   operatorRate,
		PlatformAmount: platformAmount,
		Status:         model.SplitPending,
	}

	if err := s.repo.CreateSplit(split); err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}
	return split, nil
}

func (s *FinanceService) ListSplits(page, pageSize int) ([]model.RevenueSplit, int64, error) {
	return s.repo.ListSplits(page, pageSize)
}
