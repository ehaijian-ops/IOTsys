package repository

import (
	"iot-platform/internal/maintenance/model"

	"gorm.io/gorm"
)

// MaintenanceRepository 运维数据访问层
type MaintenanceRepository struct {
	db *gorm.DB
}

func NewMaintenanceRepository(db *gorm.DB) *MaintenanceRepository {
	return &MaintenanceRepository{db: db}
}

// ========== FaultReport ==========

func (r *MaintenanceRepository) CreateFault(fault *model.FaultReport) error {
	return r.db.Create(fault).Error
}

func (r *MaintenanceRepository) GetFault(id string) (*model.FaultReport, error) {
	var f model.FaultReport
	err := r.db.First(&f, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *MaintenanceRepository) ListFaults(deviceID, status string, page, pageSize int) ([]model.FaultReport, int64, error) {
	var faults []model.FaultReport
	var total int64
	q := r.db.Model(&model.FaultReport{})
	if deviceID != "" {
		q = q.Where("device_id = ?", deviceID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&faults).Error
	return faults, total, err
}

func (r *MaintenanceRepository) UpdateFault(id string, updates map[string]interface{}) error {
	return r.db.Model(&model.FaultReport{}).Where("id = ?", id).Updates(updates).Error
}

// ========== ScheduledTask ==========

func (r *MaintenanceRepository) CreateTask(task *model.ScheduledTask) error {
	return r.db.Create(task).Error
}

func (r *MaintenanceRepository) GetTask(id string) (*model.ScheduledTask, error) {
	var t model.ScheduledTask
	err := r.db.First(&t, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *MaintenanceRepository) ListTasks(page, pageSize int) ([]model.ScheduledTask, int64, error) {
	var tasks []model.ScheduledTask
	var total int64
	if err := r.db.Model(&model.ScheduledTask{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := r.db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&tasks).Error
	return tasks, total, err
}

func (r *MaintenanceRepository) UpdateTask(id string, updates map[string]interface{}) error {
	return r.db.Model(&model.ScheduledTask{}).Where("id = ?", id).Updates(updates).Error
}

func (r *MaintenanceRepository) DeleteTask(id string) error {
	return r.db.Delete(&model.ScheduledTask{}, "id = ?", id).Error
}

// ========== TaskLog ==========

func (r *MaintenanceRepository) CreateTaskLog(log *model.TaskLog) error {
	return r.db.Create(log).Error
}

func (r *MaintenanceRepository) ListTaskLogs(taskID string, page, pageSize int) ([]model.TaskLog, int64, error) {
	var logs []model.TaskLog
	var total int64
	q := r.db.Model(&model.TaskLog{})
	if taskID != "" {
		q = q.Where("task_id = ?", taskID)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error
	return logs, total, err
}

// ========== DownloadTask ==========

func (r *MaintenanceRepository) CreateDownload(task *model.DownloadTask) error {
	return r.db.Create(task).Error
}

func (r *MaintenanceRepository) GetDownload(id string) (*model.DownloadTask, error) {
	var t model.DownloadTask
	err := r.db.First(&t, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *MaintenanceRepository) ListDownloads(createdBy string, page, pageSize int) ([]model.DownloadTask, int64, error) {
	var tasks []model.DownloadTask
	var total int64
	q := r.db.Model(&model.DownloadTask{})
	if createdBy != "" {
		q = q.Where("created_by = ?", createdBy)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&tasks).Error
	return tasks, total, err
}

func (r *MaintenanceRepository) UpdateDownload(id string, updates map[string]interface{}) error {
	return r.db.Model(&model.DownloadTask{}).Where("id = ?", id).Updates(updates).Error
}
