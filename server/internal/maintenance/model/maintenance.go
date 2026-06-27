package model

import "time"

// 故障状态常量
const (
	FaultPending   = "pending"   // 待处理
	FaultProcessing = "processing" // 处理中
	FaultResolved   = "resolved"  // 已解决
	FaultClosed     = "closed"    // 已关闭
)

// FaultReport 故障报修/反馈
type FaultReport struct {
	ID           string     `json:"id" gorm:"primaryKey;size:36"`
	DeviceID     string     `json:"device_id" gorm:"index;size:36;not null"`
	UserID       *uint      `json:"user_id" gorm:"index"`                     // 报修用户（小程序用户）
	FaultType    string     `json:"fault_type" gorm:"size:50;not null"`        // 故障类型
	Description  string     `json:"description" gorm:"type:text"`
	Images       string     `json:"images" gorm:"type:text"`                   // 故障图片（JSON数组）
	Status       string     `json:"status" gorm:"size:20;default:pending;index"`
	HandleResult string     `json:"handle_result" gorm:"type:text"`            // 处理结果
	HandlerBy    string     `json:"handler_by" gorm:"size:64"`                // 处理人
	HandledAt    *time.Time `json:"handled_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (FaultReport) TableName() string {
	return "fault_reports"
}

// ScheduledTask 定时任务配置
type ScheduledTask struct {
	ID          string     `json:"id" gorm:"primaryKey;size:36"`
	Name        string     `json:"name" gorm:"size:100;not null"`
	TaskType    string     `json:"task_type" gorm:"size:50;not null"`          // 任务类型: ota/reboot/config/report
	CronExpr    string     `json:"cron_expr" gorm:"size:100;not null"`         // Cron 表达式
	Params      string     `json:"params" gorm:"type:text"`                    // 任务参数（JSON）
	Status      string     `json:"status" gorm:"size:20;default:active;index"` // active/paused/expired
	LastRunAt   *time.Time `json:"last_run_at"`
	NextRunAt   *time.Time `json:"next_run_at"`
	Description string     `json:"description" gorm:"size:500"`
	CreatedBy   string     `json:"created_by" gorm:"size:64"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (ScheduledTask) TableName() string {
	return "scheduled_tasks"
}

// TaskLog 任务执行日志
type TaskLog struct {
	ID         uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	TaskID     string    `json:"task_id" gorm:"index;size:36;not null"`
	Status     string    `json:"status" gorm:"size:20;index"`          // success/failed/running
	Result     string    `json:"result" gorm:"type:text"`             // 执行结果
	StartAt    time.Time `json:"start_at"`
	EndAt      *time.Time `json:"end_at"`
	DurationMs int64     `json:"duration_ms" gorm:"default:0"`
	CreatedAt  time.Time `json:"created_at"`
}

func (TaskLog) TableName() string {
	return "task_logs"
}

// DownloadTask 下载任务
type DownloadTask struct {
	ID         string     `json:"id" gorm:"primaryKey;size:36"`
	TaskType   string     `json:"task_type" gorm:"size:50;not null"`           // 导出类型: orders/devices/statistics
	Params     string     `json:"params" gorm:"type:text"`                     // 导出参数（JSON）
	FileName   string     `json:"file_name" gorm:"size:255"`
	FileURL    string     `json:"file_url" gorm:"size:500"`                    // 下载地址
	FileSize   int64      `json:"file_size" gorm:"default:0"`                  // 文件大小（字节）
	Status     string     `json:"status" gorm:"size:20;default:pending;index"` // pending/processing/completed/failed
	ErrorMsg   string     `json:"error_msg" gorm:"size:500"`
	CreatedBy  string     `json:"created_by" gorm:"size:64;index"`
	CreatedAt  time.Time  `json:"created_at"`
	FinishedAt *time.Time `json:"finished_at"`
}

func (DownloadTask) TableName() string {
	return "download_tasks"
}
