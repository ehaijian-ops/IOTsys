package migrations

import (
	adModel "iot-platform/internal/advertisement/model"
	agentModel "iot-platform/internal/agent/model"
	billingModel "iot-platform/internal/billing/model"
	cardModel "iot-platform/internal/card/model"
	cmdService "iot-platform/internal/command/service"
	deviceModel "iot-platform/internal/device/model"
	financeModel "iot-platform/internal/finance/model"
	icModel "iot-platform/internal/interconnect/model"
	mtModel "iot-platform/internal/maintenance/model"
	orderModel "iot-platform/internal/order/model"
	sysModel "iot-platform/internal/system/model"
	userModel "iot-platform/internal/user/model"

	"gorm.io/gorm"
)

// AutoMigrate 自动迁移数据库表
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		// ---- 设备管理 ----
		&deviceModel.Device{},
		&deviceModel.Site{},
		&deviceModel.DeviceType{},
		&deviceModel.OTAFirmware{},
		&deviceModel.DeviceQRCode{},
		&deviceModel.DeviceOnlineLog{},
		&deviceModel.VirtualDevice{},

		// ---- 指令管理 ----
		&cmdService.DeviceCommand{},

		// ---- 用户管理 ----
		&userModel.User{},
		&adModel.WechatUser{},

		// ---- 订单与充电 ----
		&orderModel.ChargeOrder{},
		&orderModel.ChargeRecord{},
		&orderModel.OrderRefund{},

		// ---- 收费方案 ----
		&billingModel.BillingScheme{},
		&billingModel.BillingPeriod{},
		&billingModel.MonthlyCardScheme{},
		&billingModel.RechargeScheme{},
		&billingModel.BusinessConfig{},

		// ---- 财务管理 ----
		&financeModel.UserWallet{},
		&financeModel.RechargeRecord{},
		&financeModel.WithdrawRecord{},
		&financeModel.RevenueSplit{},

		// ---- 卡片管理 ----
		&cardModel.ICCard{},
		&cardModel.ICCardRecharge{},
		&cardModel.TrafficCard{},
		&cardModel.MonthlyCard{},
		&cardModel.MonthlyCardRecord{},

		// ---- 代理商与运营商 ----
		&agentModel.Agent{},
		&agentModel.Operator{},

		// ---- 互联互通 ----
		&icModel.InterconnectOrg{},
		&icModel.InterconnectKey{},

		// ---- 运营管理 ----
		&adModel.Advertisement{},
		&adModel.FranchiseApplication{},

		// ---- 运维管理 ----
		&mtModel.FaultReport{},
		&mtModel.ScheduledTask{},
		&mtModel.TaskLog{},
		&mtModel.DownloadTask{},

		// ---- 系统管理 ----
		&sysModel.Role{},
		&sysModel.Menu{},
		&sysModel.LoginLog{},
		&sysModel.SystemLog{},
	)
}
