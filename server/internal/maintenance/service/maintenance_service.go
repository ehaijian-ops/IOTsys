package service

import (
	"time"

	"iot-platform/internal/maintenance/model"
	"iot-platform/internal/maintenance/repository"
	errs "iot-platform/pkg/errors"

	"github.com/google/uuid"
)

// MaintenanceService 运维业务逻辑层
type MaintenanceService struct {
	repo *repository.MaintenanceRepository
}

func NewMaintenanceService(repo *repository.MaintenanceRepository) *MaintenanceService {
	return &MaintenanceService{repo: repo}
}

// ========== FaultReport ==========

func (s *MaintenanceService) CreateFault(deviceID, faultType, description, images string, userID *uint) (*model.FaultReport, error) {
	fault := &model.FaultReport{
		ID:          uuid.New().String(),
		DeviceID:    deviceID,
		UserID:      userID,
		FaultType:   faultType,
		Description: description,
		Images:      images,
		Status:      model.FaultPending,
	}
	if err := s.repo.CreateFault(fault); err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}
	return fault, nil
}

func (s *MaintenanceService) GetFault(id string) (*model.FaultReport, error) {
	fault, err := s.repo.GetFault(id)
	if err != nil {
		return nil, errs.NotFound("故障报告", id)
	}
	return fault, nil
}

func (s *MaintenanceService) ListFaults(deviceID, status string, page, pageSize int) ([]model.FaultReport, int64, error) {
	return s.repo.ListFaults(deviceID, status, page, pageSize)
}

func (s *MaintenanceService) HandleFault(id, result, handledBy string) error {
	fault, err := s.repo.GetFault(id)
	if err != nil {
		return errs.NotFound("故障报告", id)
	}
	if fault.Status == model.FaultResolved || fault.Status == model.FaultClosed {
		return errs.New("FAULT_ALREADY_HANDLED", "故障已处理", 400)
	}
	now := time.Now()
	return s.repo.UpdateFault(id, map[string]interface{}{
		"status":        model.FaultResolved,
		"handle_result":  result,
		"handler_by":     handledBy,
		"handled_at":     now,
	})
}

// ========== ScheduledTask ==========

func (s *MaintenanceService) CreateTask(name, taskType, cronExpr, params, description, createdBy string) (*model.ScheduledTask, error) {
	task := &model.ScheduledTask{
		ID:          uuid.New().String(),
		Name:        name,
		TaskType:    taskType,
		CronExpr:    cronExpr,
		Params:      params,
		Description: description,
		CreatedBy:   createdBy,
		Status:      "active",
	}
	if err := s.repo.CreateTask(task); err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}
	return task, nil
}

func (s *MaintenanceService) ListTasks(page, pageSize int) ([]model.ScheduledTask, int64, error) {
	return s.repo.ListTasks(page, pageSize)
}

func (s *MaintenanceService) UpdateTask(id string, updates map[string]interface{}) error {
	if _, err := s.repo.GetTask(id); err != nil {
		return errs.NotFound("定时任务", id)
	}
	return s.repo.UpdateTask(id, updates)
}

func (s *MaintenanceService) DeleteTask(id string) error {
	if _, err := s.repo.GetTask(id); err != nil {
		return errs.NotFound("定时任务", id)
	}
	return s.repo.DeleteTask(id)
}

func (s *MaintenanceService) GetTaskLogs(taskID string, page, pageSize int) ([]model.TaskLog, int64, error) {
	return s.repo.ListTaskLogs(taskID, page, pageSize)
}

// ========== DownloadTask ==========

func (s *MaintenanceService) CreateDownload(taskType, params, createdBy string) (*model.DownloadTask, error) {
	task := &model.DownloadTask{
		ID:        uuid.New().String(),
		TaskType:  taskType,
		Params:    params,
		CreatedBy: createdBy,
		Status:    "pending",
	}
	if err := s.repo.CreateDownload(task); err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}
	return task, nil
}

func (s *MaintenanceService) ListDownloads(createdBy string, page, pageSize int) ([]model.DownloadTask, int64, error) {
	return s.repo.ListDownloads(createdBy, page, pageSize)
}
