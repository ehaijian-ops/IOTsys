package model

import "time"

// StandardData 平台标准设备数据模型
// 所有厂商协议解析后都转换为此标准格式
type StandardData struct {
	// 设备标识
	DeviceID string `json:"device_id"`
	Protocol string `json:"protocol"` // AP3000_v2 / TF100_v1
	MsgID    int    `json:"msg_id"`   // 消息ID (防重放)

	// 时间
	Timestamp time.Time `json:"timestamp"`

	// 电气参数 (设备级别汇总)
	Voltage float64 `json:"voltage"` // 电压 (V)
	Current float64 `json:"current"` // 电流 (A)
	Power   float64 `json:"power"`   // 功率 (W)

	// 电能
	EnergyTotal float64 `json:"energy_total"` // 累计电量 (kWh)
	EnergyToday float64 `json:"energy_today"` // 今日电量 (kWh)

	// 充电状态
	ChargingStatus string `json:"charging_status"` // idle/charging/finished/fault
	FaultCode      string `json:"fault_code"`      // 故障码
	FaultMessage   string `json:"fault_message"`   // 故障描述

	// 环境参数
	Temperature float64 `json:"temperature"` // 温度 (℃)
	Humidity    float64 `json:"humidity"`    // 湿度 (%)

	// 设备状态
	SignalStrength int    `json:"signal_strength"` // 信号强度
	FirmwareVer    string `json:"firmware_ver"`    // 固件版本
	DeviceTypeCode int    `json:"device_type_code"` // 设备类型码

	// 电单车充电桩特有
	PortCount int        `json:"port_count,omitempty"` // 充电端口数
	PortData  []PortData `json:"port_data,omitempty"`  // 各端口数据
	PortPower int        `json:"port_power,omitempty"` // 每端口峰值功率 (W)

	// 汽车充电桩特有
	GunCount      int       `json:"gun_count,omitempty"`      // 充电枪数
	GunData       []GunData `json:"gun_data,omitempty"`       // 各枪数据
	ConnectorType string    `json:"connector_type,omitempty"` // 接口类型

	// 电单车 — 刷卡/充电记录相关
	CardID    string  `json:"card_id,omitempty"`    // 卡号
	CardType  int     `json:"card_type,omitempty"`  // 卡类型
	CardMoney float64 `json:"card_money,omitempty"` // 卡余额
	PortNum   int     `json:"port_num,omitempty"`   // 当前操作端口号

	// 充电会话详情
	OrderNo       string  `json:"order_no,omitempty"`       // 订单号
	ChargeTime    int     `json:"charge_time,omitempty"`    // 已充电时长(秒)
	ChargeEnergy  float64 `json:"charge_energy,omitempty"`  // 已充电量(kWh)
	ChargeMoney   float64 `json:"charge_money,omitempty"`   // 已充金额(元)
	PeakPower     float64 `json:"peak_power,omitempty"`     // 峰值功率(W)
	StopReason    int     `json:"stop_reason,omitempty"`    // 停止原因
	StartType     int     `json:"start_type,omitempty"`     // 启动方式
	ChargeMode    int     `json:"charge_mode,omitempty"`    // 充电模式
	VerifyCode    string  `json:"verify_code,omitempty"`    // 验证码

	// 汽车桩特有 — 费率相关
	EleCost       float64 `json:"ele_cost,omitempty"`        // 电费
	ServCost      float64 `json:"serv_cost,omitempty"`       // 服务费
	EleLoss       float64 `json:"ele_loss,omitempty"`        // 电损比例
	RateVersion   int     `json:"rate_version,omitempty"`    // 费率版本
	IsHighPrecise bool    `json:"is_high_precise,omitempty"` // 是否高精度费率

	// 汽车桩特有 — SIM/通信信息
	SIM     string `json:"sim,omitempty"`     // SIM卡号
	IMEI    string `json:"imei,omitempty"`    // IMEI
	NetType int    `json:"net_type,omitempty"` // 网络类型
	ModVer  string `json:"mod_ver,omitempty"` // 模块版本

	// 汽车桩特有 — 地锁相关
	LockAddr  int    `json:"lock_addr,omitempty"`  // 地锁地址
	LockState int    `json:"lock_state,omitempty"` // 地锁状态
	LockCover int    `json:"lock_cover,omitempty"` // 地锁是否在盖

	// 扩展字段
	Extra map[string]interface{} `json:"extra,omitempty"`
}

// PortData 电单车充电端口数据
type PortData struct {
	PortIndex   int     `json:"port_index"`
	PortStatus  string  `json:"port_status"`  // idle/charging/scanned/finished/fault/reserved
	Voltage     float64 `json:"voltage"`
	Current     float64 `json:"current"`
	Power       float64 `json:"power"`
	EnergyTotal float64 `json:"energy_total"`
	PeakPower   float64 `json:"peak_power"`
	Temperature float64 `json:"temperature"`
}

// GunData 汽车充电枪数据
type GunData struct {
	GunIndex         int     `json:"gun_index"`
	GunStatus        string  `json:"gun_status"` // idle/connected/charging/finished/fault/reserved
	Voltage          float64 `json:"voltage"`
	Current          float64 `json:"current"`
	Power            float64 `json:"power"`
	EnergyTotal      float64 `json:"energy_total"`
	ChargingProgress int     `json:"charging_progress"`
	ChargingDuration int     `json:"charging_duration"`
	RemainingTime    int     `json:"remaining_time"`
	Temperature      float64 `json:"temperature"`
	BatterySOC       int     `json:"battery_soc,omitempty"`
	MaxPower         float64 `json:"max_power,omitempty"`
	CPVoltage        float64 `json:"cp_voltage,omitempty"` // CP电压
}

// StandardCommand 平台标准指令模型
type StandardCommand struct {
	CommandID string                 `json:"command_id"`
	DeviceID  string                 `json:"device_id"`
	Protocol  string                 `json:"protocol"`
	CmdType   string                 `json:"cmd_type"` // start_charge/stop_charge/config/reboot/ota/query_status/clear_storage/change_mode
	Params    map[string]interface{} `json:"params"`
	Timeout   int                    `json:"timeout"`
	CreatedAt time.Time              `json:"created_at"`
}

// StandardCommandResponse 平台标准指令响应
type StandardCommandResponse struct {
	CommandID string                 `json:"command_id"`
	DeviceID  string                 `json:"device_id"`
	Success   bool                   `json:"success"`
	Result    map[string]interface{} `json:"result,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}
