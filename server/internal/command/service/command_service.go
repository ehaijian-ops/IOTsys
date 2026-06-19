package service

import (
	"context"
	"fmt"
	"time"

	"iot-platform/internal/connector/tcpserver"
	"iot-platform/internal/protocol/engine"
	"iot-platform/internal/protocol/model"
	"iot-platform/pkg/errors"
	"iot-platform/pkg/logger"
	"iot-platform/pkg/mq/kafka"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CommandService 指令服务
type CommandService struct {
	db       *gorm.DB
	producer *kafka.Producer
	tcpSrv   *tcpserver.Server
}

// NewCommandService 创建指令服务
func NewCommandService(db *gorm.DB, producer *kafka.Producer, tcpSrv *tcpserver.Server) *CommandService {
	return &CommandService{
		db:       db,
		producer: producer,
		tcpSrv:   tcpSrv,
	}
}

// CreateCommand 创建并下发指令
func (s *CommandService) CreateCommand(ctx context.Context, deviceID, cmdType string, params map[string]interface{}, userID string) (*DeviceCommand, error) {
	// 检查设备是否在线
	if _, ok := s.tcpSrv.GetSession(deviceID); !ok {
		return nil, errors.ErrDeviceOffline
	}

	// 获取设备协议信息
	protocol, err := s.getDeviceProtocol(ctx, deviceID)
	if err != nil {
		return nil, err
	}

	commandID := uuid.New().String()
	now := time.Now()

	// 创建指令记录
	cmd := &DeviceCommand{
		ID:        commandID,
		DeviceID:  deviceID,
		CmdType:   cmdType,
		Payload:   params,
		Status:    "pending",
		CreatedBy: userID,
		CreatedAt: now,
	}

	if err := s.db.WithContext(ctx).Create(cmd).Error; err != nil {
		return nil, fmt.Errorf("failed to create command: %w", err)
	}

	// 构建标准指令
	stdCmd := &model.StandardCommand{
		CommandID: commandID,
		DeviceID:  deviceID,
		Protocol:  protocol,
		CmdType:   cmdType,
		Params:    params,
		Timeout:   10,
		CreatedAt: now,
	}

	// 编码为设备协议帧
	frame, err := engine.Encode(stdCmd)
	if err != nil {
		cmd.Status = "failed"
		cmd.Result = map[string]interface{}{"error": err.Error()}
		s.db.Save(cmd)
		return nil, fmt.Errorf("failed to encode command: %w", err)
	}

	// 更新状态为已下发
	cmd.Status = "sent"
	s.db.Save(cmd)

	// 发送到设备
	if err := s.tcpSrv.SendCommand(deviceID, frame); err != nil {
		cmd.Status = "failed"
		cmd.Result = map[string]interface{}{"error": err.Error()}
		s.db.Save(cmd)
		return nil, fmt.Errorf("failed to send command: %w", err)
	}

	logger.Info("Command sent to device",
		zap.String("command_id", commandID),
		zap.String("device_id", deviceID),
		zap.String("cmd_type", cmdType),
	)

	return cmd, nil
}

// GetCommand 查询指令详情
func (s *CommandService) GetCommand(ctx context.Context, commandID string) (*DeviceCommand, error) {
	var cmd DeviceCommand
	if err := s.db.WithContext(ctx).Where("id = ?", commandID).First(&cmd).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("Command", commandID)
		}
		return nil, err
	}
	return &cmd, nil
}

// ListCommands 查询设备指令列表
func (s *CommandService) ListCommands(ctx context.Context, deviceID string, page, pageSize int) ([]DeviceCommand, int64, error) {
	var commands []DeviceCommand
	var total int64

	query := s.db.WithContext(ctx).Model(&DeviceCommand{}).Where("device_id = ?", deviceID)
	query.Count(&total)

	if err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&commands).Error; err != nil {
		return nil, 0, err
	}

	return commands, total, nil
}

// UpdateCommandStatus 更新指令状态
func (s *CommandService) UpdateCommandStatus(ctx context.Context, commandID, status string, result map[string]interface{}) error {
	updates := map[string]interface{}{
		"status":       status,
		"result":       result,
		"responded_at": time.Now(),
	}

	if err := s.db.WithContext(ctx).Model(&DeviceCommand{}).
		Where("id = ?", commandID).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update command status: %w", err)
	}

	return nil
}

// getDeviceProtocol 获取设备协议
func (s *CommandService) getDeviceProtocol(ctx context.Context, deviceID string) (string, error) {
	var device struct {
		Protocol string
	}
	if err := s.db.WithContext(ctx).
		Table("devices").
		Select("protocol").
		Where("id = ?", deviceID).
		First(&device).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", errors.NotFound("Device", deviceID)
		}
		return "", err
	}
	return device.Protocol, nil
}

// DeviceCommand 指令数据模型
type DeviceCommand struct {
	ID          string                 `json:"id" gorm:"primaryKey;size:36"`
	DeviceID    string                 `json:"device_id" gorm:"index;size:36;not null"`
	CmdType     string                 `json:"cmd_type" gorm:"size:50;not null"`
	Payload     map[string]interface{} `json:"payload" gorm:"serializer:json"`
	Status      string                 `json:"status" gorm:"size:20;default:pending"` // pending/sent/responded/success/failed/timeout
	Result      map[string]interface{} `json:"result" gorm:"serializer:json"`
	CreatedBy   string                 `json:"created_by" gorm:"size:36"`
	CreatedAt   time.Time              `json:"created_at"`
	RespondedAt *time.Time             `json:"responded_at"`
}

func (DeviceCommand) TableName() string {
	return "device_commands"
}
