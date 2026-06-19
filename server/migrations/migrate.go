package migrations

import (
	"iot-platform/internal/device/model"
	cmdService "iot-platform/internal/command/service"

	"gorm.io/gorm"
)

// AutoMigrate 自动迁移数据库表
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		// 设备管理
		&model.Device{},
		&model.Site{},

		// 指令管理
		&cmdService.DeviceCommand{},

		// TODO: 后续添加更多表
		// &alert.Alert{},
		// &user.User{},
		// &report.ChargeRecord{},
	)
}
