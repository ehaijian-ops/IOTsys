package tf100

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"iot-platform/internal/protocol/engine"
	"iot-platform/internal/protocol/model"
)

// TF100 电动汽车充电桩协议适配器
// 通信方式: TCP 长连接
// 帧格式: CCMD:长度{JSON} (设备上报) / SCMD:长度{JSON} (服务器下发)
// 长度字段固定占3字节, 不足前补0; 长度达到4位时使用4位
// JSON中msgType字段标识指令类型
// 设备ID: devId字段, 数字型
// 消息ID: msgId字段, 防重放

const (
	ProtocolName    = "TF100_v1"
	ProtocolVersion = "2.9"
	DeviceType      = "ev_charger"

	// 帧前缀
	DevicePrefix = "CCMD:" // 设备上报前缀
	ServerPrefix = "SCMD:" // 服务器下发前缀

	// ---- 设备上报 msgType ----
	MsgRegister        = 20 // 设备注册
	MsgStatusReport    = 21 // 设备状态上报
	MsgSwipeCard       = 2  // 刷卡/余额查询
	MsgChargeRecord    = 3  // 充电记录上传
	MsgDeviceAlarm     = 4  // 设备告警
	MsgChargeProgress  = 6  // 充电中实时数据
	MsgCommInfo        = 11 // 通信模块信息
	MsgMeterEnergy     = 17 // 电表累计电量
	MsgGetTime         = 22 // 获取服务器时间
	MsgPushInfo        = 23 // 推送信息(枪状态)
	MsgLockHeartbeat   = 24 // 地锁心跳
	MsgLockOccupied    = 25 // 地锁占用记录
	MsgLockStatus      = 26 // 地锁状态上报
	MsgLockAddrQuery   = 27 // 地锁地址查询
	MsgGunTimeline     = 28 // 枪时间线

	// ---- 服务器下发 msgType ----
	MsgQueryStatus     = 81 // 查询设备状态
	MsgStartCharge     = 82 // 启动充电
	MsgStopCharge      = 83 // 停止充电
	MsgSetRate         = 84 // 设置费率
	MsgQueryRate       = 85 // 查询费率
	MsgSetDeviceParam  = 86 // 设置设备参数
	MsgReboot          = 87 // 远程重启
	MsgClearStorage    = 88 // 清除存储
	MsgVoice           = 89 // 语音播报
	MsgChargeAdjust    = 90 // 充电调整
	MsgSetCardMode     = 91 // 设置刷卡模式
	MsgQueryCardMode   = 92 // 查询刷卡模式
	MsgQueryDevParam   = 93 // 查询设备参数
	MsgModifyCurrent   = 94 // 修改充电电流
	MsgSetQRCode       = 95 // 设置二维码
	MsgSetCommAddr     = 96 // 设置通信地址
	MsgSetCardInfo     = 97 // 设置IC卡信息
	MsgOTAControl      = 98 // OTA控制板
	MsgOTAPower        = 99 // OTA电源板
	MsgLockUp          = 161 // 地锁升起
	MsgLockDown        = 162 // 地锁降下
	MsgLockQuery       = 163 // 地锁状态查询
	MsgLockCoverCheck  = 164 // 地锁遮挡检测
	MsgLockAddrSet     = 165 // 地锁地址修改
	MsgLockAddrFind    = 166 // 地锁地址查找
	MsgReserved167     = 167 // 预留
	MsgReserved168     = 168 // 预留
	MsgTest16Protocol  = 999 // 测试16进制协议
)

// TF100Message 通用TF100 JSON消息结构
type TF100Message struct {
	MsgType    int             `json:"msgType"`
	DevID      int64           `json:"devId"`
	MsgID      int             `json:"msgId"`
	// 注册
	MainVer  int   `json:"mainVer,omitempty"`
	PowVer   []int `json:"powVer,omitempty"`
	Ports    int   `json:"ports,omitempty"`
	VsID     int   `json:"vsId,omitempty"`
	DevType  int   `json:"devType,omitempty"`
	PeType   int   `json:"peType,omitempty"`
	EnUpd    int   `json:"enUpd,omitempty"`
	RateVer  int   `json:"rateVer,omitempty"`
	RateX    int   `json:"rateX,omitempty"`
	TcMode   int   `json:"tcMode,omitempty"`
	UseOldSetCard int `json:"useOldSetCard,omitempty"`
	// 状态
	DevTem    int     `json:"devTem,omitempty"`
	Status    []int   `json:"status,omitempty"`
	Voltage   []int   `json:"voltage,omitempty"`
	CPVol     []int   `json:"cpVol,omitempty"`
	// 刷卡
	CardID       string `json:"cardId,omitempty"`
	CardType     int    `json:"cardType,omitempty"`
	Port         int    `json:"port,omitempty"`
	Remain       int    `json:"remain,omitempty"`
	CustomCardID string `json:"customCardId,omitempty"`
	CardStatus   int    `json:"cardStatus,omitempty"`
	VoiceMode    int    `json:"voiceMode,omitempty"`
	// 充电记录
	Time      int    `json:"time,omitempty"`
	MaxCurrent int   `json:"maxCurrent,omitempty"`
	Energy    int    `json:"energy,omitempty"`
	Money     int    `json:"money,omitempty"`
	MoneyX    float64 `json:"moneyX,omitempty"`
	StartType int    `json:"startType,omitempty"`
	Stops     int    `json:"stops,omitempty"`
	Orders    string `json:"orders,omitempty"`
	EleCost   int    `json:"eleCost,omitempty"`
	ServCost  int    `json:"servCost,omitempty"`
	EleXCost  float64 `json:"eleXCost,omitempty"`
	ServXCost float64 `json:"servXCost,omitempty"`
	EleLoss   int    `json:"eleLoss,omitempty"`
	// 充电中实时
	NumTem      int    `json:"numTem,omitempty"`
	Current     int    `json:"current,omitempty"`
	Power       int    `json:"power,omitempty"`
	TotalEnergy int    `json:"totalEnergy,omitempty"`
	TickEnergy  int    `json:"tickEnergy,omitempty"`
	TotalMoney  int    `json:"totalMoney,omitempty"`
	TickMoney   int    `json:"tickMoney,omitempty"`
	RTC         int64  `json:"rtc,omitempty"`
	TerminalL   int    `json:"terminalL,omitempty"`
	TerminalN   int    `json:"terminalN,omitempty"`
	// 告警
	ErrType []int `json:"errType,omitempty"`
	// 通信信息
	RTCType int    `json:"rtcType,omitempty"`
	Sigv    int    `json:"sigv,omitempty"`
	Netv    int    `json:"netv,omitempty"`
	SIM     string `json:"sim,omitempty"`
	IMEI    string `json:"imei,omitempty"`
	ModVer  string `json:"modVer,omitempty"`
	// 电表电量
	MeEnergy []int `json:"meEnergy,omitempty"`
	// 推送/地锁
	PushType  int    `json:"pushType,omitempty"`
	Count     int    `json:"count,omitempty"`
	Heartbeat []LockHB `json:"heartbeat,omitempty"`
	Addr      int    `json:"addr,omitempty"`
	State     int    `json:"state,omitempty"`
	Cover     int    `json:"cover,omitempty"`
	// 枪时间线
	HolderOut        int64 `json:"holderOut,omitempty"`
	CarIn            int64 `json:"carIn,omitempty"`
	AppointmentStart int64 `json:"appointmentStart,omitempty"`
	ChargeStart      int64 `json:"chargeStart,omitempty"`
	ChargeStop       int64 `json:"chargeStop,omitempty"`
	CarOut           int64 `json:"carOut,omitempty"`
	HolderIn         int64 `json:"holderIn,omitempty"`
	// 指令响应
	Ack int `json:"ack,omitempty"`
	// 费率
	TotalRate int        `json:"totalRate,omitempty"`
	Server    int        `json:"server,omitempty"`
	Rates     []RateSlot `json:"rates,omitempty"`
	// 设备参数
	MaxTime      int `json:"maxTime,omitempty"`
	MinSwitch    int `json:"minSwitch,omitempty"`
	MinCurrent   int `json:"minCurrent,omitempty"`
	MinTime      int `json:"minTime,omitempty"`
	MaxDevTem    int `json:"maxDevTem,omitempty"`
	PortTickTime int `json:"portTickTime,omitempty"`
	ChTickTime   int `json:"chTickTime,omitempty"`
	BeepOff      int `json:"beepOff,omitempty"`
	Language     int `json:"language,omitempty"`
	NetMode      int `json:"netMode,omitempty"`
	// 充电调整
	Mode     int    `json:"mode,omitempty"`
	Value    int    `json:"value,omitempty"`
	NewValue int    `json:"newValue,omitempty"`
	NewAck   int    `json:"newAck,omitempty"`
	// OTA
	TotalPack int    `json:"totalPack,omitempty"`
	CurPack   int    `json:"curPack,omitempty"`
	Data      string `json:"data,omitempty"`
	// 刷卡模式
	Pwd      string `json:"pwd,omitempty"`
	Sect     int    `json:"sect,omitempty"`
	Model    int    `json:"model,omitempty"`
	BindSw   int    `json:"bindSw,omitempty"`
	LockSw   int    `json:"lockSw,omitempty"`
	LockKey  string `json:"lockKey,omitempty"`
	DecMoney int    `json:"decMoney,omitempty"`
	MaxMoney int    `json:"maxMoney,omitempty"`
	RefundSw int    `json:"refundSw,omitempty"`
	// 二维码
	QRStr  string `json:"qrStr,omitempty"`
	QRType int    `json:"qrType,omitempty"`
	// 通信地址
	IP    string `json:"ip,omitempty"`
	SockB int    `json:"sockb,omitempty"`
	// IC卡
	Block     int `json:"block,omitempty"`
	ByteStart int `json:"byteStart,omitempty"`
	ByteRead  int `json:"byteRead,omitempty"`
	Radix     int `json:"radix,omitempty"`
	Ends      int `json:"ends,omitempty"`
	IDLen     int `json:"idLen,omitempty"`
	InZero    int `json:"inZero,omitempty"`
	// 地锁
	SrcAddr int    `json:"srcAddr,omitempty"`
	DstAddr int    `json:"dstAddr,omitempty"`
	Result  int    `json:"result,omitempty"`
	Fun     int    `json:"fun,omitempty"`
	// 预留
	WaitTime   int `json:"waitTime,omitempty"`
	ResTime    int64 `json:"resTime,omitempty"`
}

// LockHB 地锁心跳数据
type LockHB struct {
	Addr  int `json:"addr"`
	State int `json:"state"`
	Cover int `json:"cover"`
}

// RateSlot 费率时段
type RateSlot struct {
	StartTime int     `json:"startTime"`
	EndTime   int     `json:"endTime"`
	Ele       int     `json:"ele"`
	EleX      float64 `json:"eleX,omitempty"`
	Fee       int     `json:"fee"`
	FeeX      float64 `json:"feeX,omitempty"`
}

type TF100Adapter struct{}

func init() {
	engine.Register(&TF100Adapter{})
}

func (a *TF100Adapter) Name() string       { return ProtocolName }
func (a *TF100Adapter) Version() string    { return ProtocolVersion }
func (a *TF100Adapter) DeviceType() string { return DeviceType }

// Validate 校验是否为 TF100 协议帧 (CCMD:/SCMD: + 长度 + JSON)
func (a *TF100Adapter) Validate(raw []byte) bool {
	s := string(raw)
	if !strings.HasPrefix(s, "CCMD:") && !strings.HasPrefix(s, "SCMD:") {
		return false
	}
	// 找到 { 开始位置
	idx := strings.IndexByte(s, '{')
	if idx < 0 {
		return false
	}
	// 校验长度前缀
	prefixEnd := 5 // "CCMD:" or "SCMD:"
	if len(s) > 5 && s[5] >= '0' && s[5] <= '9' {
		// 至少3位长度
		if idx-prefixEnd < 3 {
			return false
		}
	}
	// 验证JSON格式
	msg := TF100Message{}
	if err := json.Unmarshal([]byte(s[idx:]), &msg); err != nil {
		return false
	}
	return msg.DevID > 0
}

// Decode 解析 TF100 数据上报帧 → 标准数据模型
func (a *TF100Adapter) Decode(raw []byte) (*model.StandardData, error) {
	s := string(raw)

	// 只处理设备上报 CCMD
	if !strings.HasPrefix(s, "CCMD:") {
		return nil, fmt.Errorf("not a device upload frame")
	}

	idx := strings.IndexByte(s, '{')
	if idx < 0 {
		return nil, fmt.Errorf("no JSON found")
	}

	msg := TF100Message{}
	if err := json.Unmarshal([]byte(s[idx:]), &msg); err != nil {
		return nil, fmt.Errorf("JSON parse error: %w", err)
	}

	std := &model.StandardData{
		DeviceID:  fmt.Sprintf("%d", msg.DevID),
		Protocol:  ProtocolName,
		Timestamp: time.Now(),
		MsgID:     msg.MsgID,
		Extra:     make(map[string]interface{}),
	}

	switch msg.MsgType {
	case MsgRegister: // 20
		a.parseRegister(&msg, std)
	case MsgStatusReport: // 21
		a.parseStatusReport(&msg, std)
	case MsgSwipeCard: // 2
		a.parseSwipeCard(&msg, std)
	case MsgChargeRecord: // 3
		a.parseChargeRecord(&msg, std)
	case MsgDeviceAlarm: // 4
		a.parseDeviceAlarm(&msg, std)
	case MsgChargeProgress: // 6
		a.parseChargeProgress(&msg, std)
	case MsgCommInfo: // 11
		a.parseCommInfo(&msg, std)
	case MsgMeterEnergy: // 17
		a.parseMeterEnergy(&msg, std)
	case MsgGetTime: // 22
		std.Extra["cmd"] = "get_time"
	case MsgPushInfo: // 23
		a.parsePushInfo(&msg, std)
	case MsgLockHeartbeat: // 24
		a.parseLockHeartbeat(&msg, std)
	case MsgLockOccupied: // 25
		a.parseLockOccupied(&msg, std)
	case MsgLockStatus: // 26
		a.parseLockStatus(&msg, std)
	case MsgLockAddrQuery: // 27
		a.parseLockAddrQuery(&msg, std)
	case MsgGunTimeline: // 28
		a.parseGunTimeline(&msg, std)
	default:
		std.Extra["raw_msg_type"] = msg.MsgType
	}

	return std, nil
}

// parseRegister 解析20注册
func (a *TF100Adapter) parseRegister(msg *TF100Message, std *model.StandardData) {
	std.FirmwareVer = fmt.Sprintf("V%d.%02d", msg.MainVer/100, msg.MainVer%100)
	std.GunCount = msg.Ports
	std.DeviceTypeCode = msg.DevType
	std.Extra["pe_type"] = msg.PeType
	std.Extra["enable_update"] = msg.EnUpd
	std.Extra["rate_version"] = msg.RateVer
	std.IsHighPrecise = msg.RateX == 1
	std.Extra["tc_mode"] = msg.TcMode
	std.Extra["use_old_set_card"] = msg.UseOldSetCard

	if len(msg.PowVer) >= 2 {
		std.Extra["power_fw_ver"] = fmt.Sprintf("V%d.%02d/V%d.%02d",
			msg.PowVer[0]/100, msg.PowVer[0]%100,
			msg.PowVer[1]/100, msg.PowVer[1]%100)
	}
}

// parseStatusReport 解析21状态
func (a *TF100Adapter) parseStatusReport(msg *TF100Message, std *model.StandardData) {
	if msg.DevTem != 0 {
		std.Temperature = float64(msg.DevTem) - 65
	}
	std.GunCount = msg.Ports

	// 各枪状态
	std.GunData = make([]model.GunData, msg.Ports)
	for i := 0; i < msg.Ports && i < len(msg.Status); i++ {
		std.GunData[i] = model.GunData{
			GunIndex:  i + 1,
			GunStatus: a.parseGunStatus(msg.Status[i]),
		}
	}

	// 各枪电压
	for i := 0; i < msg.Ports && i < len(msg.Voltage); i++ {
		std.GunData[i].Voltage = float64(msg.Voltage[i]) * 0.1
	}

	// CP电压
	for i := 0; i < msg.Ports && i < len(msg.CPVol); i++ {
		std.GunData[i].CPVoltage = float64(msg.CPVol[i]) * 0.01
	}
}

// parseSwipeCard 解析刷卡
func (a *TF100Adapter) parseSwipeCard(msg *TF100Message, std *model.StandardData) {
	std.CardID = msg.CardID
	std.CardType = msg.CardType
	std.PortNum = msg.Port
	std.CardMoney = float64(msg.Remain) / 100.0 // 分转元
	std.Extra["card_status"] = msg.CardStatus
	std.Extra["voice_mode"] = msg.VoiceMode
	std.Extra["custom_card_id"] = msg.CustomCardID
}

// parseChargeRecord 解析充电记录
func (a *TF100Adapter) parseChargeRecord(msg *TF100Message, std *model.StandardData) {
	std.ChargeTime = msg.Time
	std.ChargeEnergy = float64(msg.Energy) * 0.001 // 0.001度精度
	std.PortNum = msg.Port
	std.StartType = msg.StartType
	std.CardID = msg.CardID
	std.StopReason = msg.Stops
	std.OrderNo = msg.Orders

	// 金额 (高精度优先)
	if msg.IsHighPrecise() {
		std.ChargeMoney = msg.MoneyX
		std.EleCost = msg.EleXCost
		std.ServCost = msg.ServXCost
	} else {
		std.ChargeMoney = float64(msg.Money) * 0.01
		std.EleCost = float64(msg.EleCost) * 0.01
		std.ServCost = float64(msg.ServCost) * 0.01
	}

	std.EleLoss = float64(msg.EleLoss) / 100.0
}

// parseDeviceAlarm 解析设备告警
func (a *TF100Adapter) parseDeviceAlarm(msg *TF100Message, std *model.StandardData) {
	std.GunCount = msg.Ports

	std.GunData = make([]model.GunData, msg.Ports)
	for i := 0; i < msg.Ports && i < len(msg.ErrType); i++ {
		std.GunData[i] = model.GunData{
			GunIndex:  i + 1,
			GunStatus: a.getAlarmGunStatus(msg.ErrType[i]),
		}
		if msg.ErrType[i] != 0 {
			std.FaultCode = fmt.Sprintf("ERR_%d", msg.ErrType[i])
			std.FaultMessage = a.getFaultMessage(msg.ErrType[i])
		}
	}
}

// parseChargeProgress 解析充电中实时数据
func (a *TF100Adapter) parseChargeProgress(msg *TF100Message, std *model.StandardData) {
	std.PortNum = msg.Port
	std.ChargeTime = msg.Time
	std.StartType = msg.StartType

	if msg.NumTem != 0 {
		std.Extra["port_temp"] = float64(msg.NumTem) - 65
	}
	if msg.DevTem != 0 {
		std.Temperature = float64(msg.DevTem) - 65
	}

	std.Voltage = float64(msg.Voltage) * 0.1
	std.Current = float64(msg.Current) * 0.01
	std.Power = float64(msg.Power) * 0.1
	std.ChargeEnergy = float64(msg.TotalEnergy) * 0.001
	std.ChargeMoney = float64(msg.TotalMoney) * 0.01
	std.OrderNo = msg.Orders
	std.EleCost = float64(msg.EleCost) * 0.01
	std.ServCost = float64(msg.ServCost) * 0.01
	std.EleLoss = float64(msg.EleLoss) / 100.0

	if msg.RTC > 0 {
		std.Timestamp = time.Unix(msg.RTC, 0)
	}

	// 接线端子温度
	if msg.TerminalL != 0 {
		std.Extra["terminal_l_temp"] = float64(msg.TerminalL) - 65
	}
	if msg.TerminalN != 0 {
		std.Extra["terminal_n_temp"] = float64(msg.TerminalN) - 65
	}

	std.GunData = []model.GunData{{
		GunIndex:    msg.Port,
		GunStatus:   "charging",
		Voltage:     std.Voltage,
		Current:     std.Current,
		Power:       std.Power,
		EnergyTotal: std.ChargeEnergy,
	}}
}

// parseCommInfo 解析通信模块信息
func (a *TF100Adapter) parseCommInfo(msg *TF100Message, std *model.StandardData) {
	std.SignalStrength = msg.Sigv
	std.NetType = msg.Netv
	std.SIM = msg.SIM
	std.IMEI = msg.IMEI
	std.ModVer = msg.ModVer
	std.DeviceTypeCode = msg.DevType

	if msg.RTC > 0 {
		std.Timestamp = time.Unix(msg.RTC, 0)
	}
	std.Extra["rtc_type"] = msg.RTCType
}

// parseMeterEnergy 解析电表累计电量
func (a *TF100Adapter) parseMeterEnergy(msg *TF100Message, std *model.StandardData) {
	std.GunCount = msg.Ports
	std.GunData = make([]model.GunData, msg.Ports)
	for i := 0; i < msg.Ports && i < len(msg.MeEnergy); i++ {
		std.GunData[i] = model.GunData{
			GunIndex:    i + 1,
			EnergyTotal: float64(msg.MeEnergy[i]) * 0.01,
		}
	}
}

// parsePushInfo 解析推送信息
func (a *TF100Adapter) parsePushInfo(msg *TF100Message, std *model.StandardData) {
	std.PortNum = msg.Port
	std.OrderNo = msg.Orders
	std.Extra["push_type"] = msg.PushType
	std.Extra["push_count"] = msg.Count
}

// parseLockHeartbeat 解析地锁心跳
func (a *TF100Adapter) parseLockHeartbeat(msg *TF100Message, std *model.StandardData) {
	std.Extra["lock_heartbeat"] = msg.Heartbeat
	if len(msg.Heartbeat) > 0 {
		std.LockAddr = msg.Heartbeat[0].Addr
		std.LockState = msg.Heartbeat[0].State
		std.LockCover = msg.Heartbeat[0].Cover
	}
}

// parseLockOccupied 解析地锁占用
func (a *TF100Adapter) parseLockOccupied(msg *TF100Message, std *model.StandardData) {
	std.LockAddr = msg.Addr
	std.ChargeTime = msg.Time
	std.OrderNo = msg.Orders
}

// parseLockStatus 解析地锁状态
func (a *TF100Adapter) parseLockStatus(msg *TF100Message, std *model.StandardData) {
	std.PortNum = msg.Port
	std.LockAddr = msg.Addr
	std.LockState = msg.State
	std.LockCover = msg.Cover
	std.ChargeTime = msg.Time
	std.OrderNo = msg.Orders
}

// parseLockAddrQuery 解析地锁地址查询
func (a *TF100Adapter) parseLockAddrQuery(msg *TF100Message, std *model.StandardData) {
	std.Extra["lock_addr_query"] = true
}

// parseGunTimeline 解析枪时间线
func (a *TF100Adapter) parseGunTimeline(msg *TF100Message, std *model.StandardData) {
	std.PortNum = msg.Port
	std.OrderNo = msg.Orders
	std.Extra["gun_timeline"] = map[string]interface{}{
		"holder_out":        msg.HolderOut,
		"car_in":            msg.CarIn,
		"appointment_start": msg.AppointmentStart,
		"charge_start":      msg.ChargeStart,
		"charge_stop":       msg.ChargeStop,
		"car_out":           msg.CarOut,
		"holder_in":         msg.HolderIn,
	}
}

// IsHighPrecise 判断是否高精度费率模式
func (msg *TF100Message) IsHighPrecise() bool {
	return msg.MoneyX != 0 || msg.EleXCost != 0 || msg.ServXCost != 0
}

// parseGunStatus 解析枪状态
func (a *TF100Adapter) parseGunStatus(status int) string {
	switch status {
	case 0:
		return "idle"
	case 1:
		return "charging"
	case 2:
		return "waiting" // 等待充电确认
	case 3:
		return "finished"
	case 4:
		return "reserved" // 定时预约
	case 5:
		return "plugged" // 已插枪未充电
	default:
		return fmt.Sprintf("unknown_%d", status)
	}
}

// getAlarmGunStatus 根据告警类型获取枪状态
func (a *TF100Adapter) getAlarmGunStatus(errType int) string {
	if errType == 0 {
		return "normal"
	}
	return "fault"
}

// getFaultMessage 获取故障描述
func (a *TF100Adapter) getFaultMessage(errType int) string {
	messages := map[int]string{
		1:  "存储损坏",
		2:  "接地故障",
		3:  "继电器粘连",
		4:  "漏电故障",
		5:  "CP电压异常",
		6:  "急停",
		7:  "漏电故障",
		8:  "短路故障",
		9:  "门锁故障",
		10: "输入电压异常",
		11: "继电器脱扣",
		12: "桩停机",
	}
	if msg, ok := messages[errType]; ok {
		return msg
	}
	return fmt.Sprintf("未知故障_%d", errType)
}

// getStopReason 获取停止原因描述
func (a *TF100Adapter) getStopReason(reason int) string {
	messages := map[int]string{
		0:  "手动停止",
		1:  "达到最大充电时间",
		2:  "达到预设时间",
		3:  "达到预设电量",
		4:  "达到预设金额",
		5:  "用户拔枪",
		6:  "继电器故障停止",
		7:  "环境温度过高",
		8:  "EEPROM错误",
		9:  "接地错误",
		10: "漏电停止",
		11: "CP电压异常",
		12: "漏电异常",
		13: "短路异常",
		14: "急停",
		15: "电流过小停止",
		16: "继电器粘连停止",
		17: "门锁停止",
		18: "电压过小停止",
		19: "车辆充电超时",
		20: "电流为0停止",
		21: "人工手动停止",
		22: "继电器脱扣停止",
		23: "绝缘检测停止",
		24: "急停停止",
		25: "温度过高停止",
		26: "枪温过高停止",
	}
	if msg, ok := messages[reason]; ok {
		return msg
	}
	return fmt.Sprintf("未知原因_%d", reason)
}

// Encode 编码标准指令 → TF100 协议帧
func (a *TF100Adapter) Encode(cmd *model.StandardCommand) ([]byte, error) {
	devID, _ := strconv.ParseInt(cmd.DeviceID, 10, 64)
	msgID := int(time.Now().UnixNano() % 65535)

	msg := TF100Message{
		DevID: devID,
		MsgID: msgID,
	}

	switch cmd.CmdType {
	case "query_status":
		msg.MsgType = MsgQueryStatus
	case "start_charge":
		msg.MsgType = MsgStartCharge
		a.fillStartCharge(cmd, &msg)
	case "stop_charge":
		msg.MsgType = MsgStopCharge
		a.fillStopCharge(cmd, &msg)
	case "set_rate":
		msg.MsgType = MsgSetRate
		a.fillSetRate(cmd, &msg)
	case "query_rate":
		msg.MsgType = MsgQueryRate
	case "config":
		msg.MsgType = MsgSetDeviceParam
		a.fillSetDeviceParam(cmd, &msg)
	case "query_config":
		msg.MsgType = MsgQueryDevParam
	case "reboot":
		msg.MsgType = MsgReboot
	case "clear_storage":
		msg.MsgType = MsgClearStorage
	case "voice":
		msg.MsgType = MsgVoice
		a.fillVoice(cmd, &msg)
	case "charge_adjust":
		msg.MsgType = MsgChargeAdjust
		a.fillChargeAdjust(cmd, &msg)
	case "set_card_mode":
		msg.MsgType = MsgSetCardMode
		a.fillCardMode(cmd, &msg)
	case "query_card_mode":
		msg.MsgType = MsgQueryCardMode
	case "modify_current":
		msg.MsgType = MsgModifyCurrent
		a.fillModifyCurrent(cmd, &msg)
	case "set_qrcode":
		msg.MsgType = MsgSetQRCode
		a.fillQRCode(cmd, &msg)
	case "set_comm_addr":
		msg.MsgType = MsgSetCommAddr
		a.fillCommAddr(cmd, &msg)
	case "lock_up":
		msg.MsgType = MsgLockUp
		a.fillLockCmd(cmd, &msg)
	case "lock_down":
		msg.MsgType = MsgLockDown
		a.fillLockCmd(cmd, &msg)
	case "lock_query":
		msg.MsgType = MsgLockQuery
		a.fillLockQuery(cmd, &msg)
	case "lock_addr_set":
		msg.MsgType = MsgLockAddrSet
		a.fillLockAddrSet(cmd, &msg)
	default:
		return nil, fmt.Errorf("unsupported command type: %s", cmd.CmdType)
	}

	return a.buildFrame(msg)
}

// buildFrame 构建 TF100 JSON 帧
func (a *TF100Adapter) buildFrame(msg TF100Message) ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("JSON marshal error: %w", err)
	}

	jsonLen := len(jsonBytes)
	// 长度字段3位或4位
	lenStr := fmt.Sprintf("%03d", jsonLen)
	if jsonLen >= 1000 {
		lenStr = fmt.Sprintf("%04d", jsonLen)
	}

	frame := fmt.Sprintf("SCMD:%s%s", lenStr, string(jsonBytes))
	return []byte(frame), nil
}

// fillStartCharge 填充启动充电参数
func (a *TF100Adapter) fillStartCharge(cmd *model.StandardCommand, msg *TF100Message) {
	if v, ok := cmd.Params["port"].(float64); ok {
		msg.Port = int(v)
	} else {
		msg.Port = 1
	}
	if v, ok := cmd.Params["mode"].(float64); ok {
		msg.Mode = int(v) // 0=自动, 1=按时, 2=按量, 3=按金额
	}
	if v, ok := cmd.Params["voice_mode"].(float64); ok {
		msg.VoiceMode = int(v)
	}
	if v, ok := cmd.Params["remain"].(float64); ok {
		msg.Remain = int(v) // 金额(分)/有效期
	}
	if v, ok := cmd.Params["energy"].(float64); ok {
		msg.Energy = int(v) // 0.001度精度
	}
	if v, ok := cmd.Params["time"].(float64); ok {
		msg.Time = int(v) // 秒
	}
	if v, ok := cmd.Params["money"].(float64); ok {
		msg.Money = int(v) // 分
	}
	if v, ok := cmd.Params["res_time"].(float64); ok {
		msg.ResTime = int64(v) // 预约时间
	}
	if v, ok := cmd.Params["wait_time"].(float64); ok {
		msg.WaitTime = int(v)
	}
	if v, ok := cmd.Params["current"].(float64); ok {
		msg.Current = int(v)
	}
	if v, ok := cmd.Params["order_no"].(string); ok {
		msg.Orders = v
	}
}

// fillStopCharge 填充停止充电参数
func (a *TF100Adapter) fillStopCharge(cmd *model.StandardCommand, msg *TF100Message) {
	if v, ok := cmd.Params["port"].(float64); ok {
		msg.Port = int(v)
	}
}

// fillSetRate 填充设置费率
func (a *TF100Adapter) fillSetRate(cmd *model.StandardCommand, msg *TF100Message) {
	if v, ok := cmd.Params["ele_loss"].(float64); ok {
		msg.EleLoss = int(v * 100)
	}
	if v, ok := cmd.Params["server_fee"].(float64); ok {
		msg.Server = int(v * 100)
	}
	if v, ok := cmd.Params["rate_ver"].(float64); ok {
		msg.RateVer = int(v)
	}
	if rates, ok := cmd.Params["rates"].([]interface{}); ok {
		msg.TotalRate = len(rates)
		for _, r := range rates {
			if rm, ok2 := r.(map[string]interface{}); ok2 {
				slot := RateSlot{}
				if v, ok3 := rm["start_time"].(float64); ok3 {
					slot.StartTime = int(v)
				}
				if v, ok3 := rm["end_time"].(float64); ok3 {
					slot.EndTime = int(v)
				}
				if v, ok3 := rm["ele"].(float64); ok3 {
					slot.Ele = int(v)
				}
				if v, ok3 := rm["fee"].(float64); ok3 {
					slot.Fee = int(v)
				}
				msg.Rates = append(msg.Rates, slot)
			}
		}
	}
}

// fillSetDeviceParam 填充设置设备参数
func (a *TF100Adapter) fillSetDeviceParam(cmd *model.StandardCommand, msg *TF100Message) {
	if v, ok := cmd.Params["max_time"].(float64); ok {
		msg.MaxTime = int(v)
	}
	if v, ok := cmd.Params["max_current"].(float64); ok {
		msg.MaxCurrent = int(v)
	}
	if v, ok := cmd.Params["min_switch"].(float64); ok {
		msg.MinSwitch = int(v)
	}
	if v, ok := cmd.Params["min_current"].(float64); ok {
		msg.MinCurrent = int(v)
	}
	if v, ok := cmd.Params["min_time"].(float64); ok {
		msg.MinTime = int(v)
	}
	if v, ok := cmd.Params["max_dev_tem"].(float64); ok {
		msg.MaxDevTem = int(v)
	}
	if v, ok := cmd.Params["port_tick_time"].(float64); ok {
		msg.PortTickTime = int(v)
	}
	if v, ok := cmd.Params["ch_tick_time"].(float64); ok {
		msg.ChTickTime = int(v)
	}
	if v, ok := cmd.Params["beep_off"].(float64); ok {
		msg.BeepOff = int(v)
	}
	if v, ok := cmd.Params["language"].(float64); ok {
		msg.Language = int(v)
	}
	if v, ok := cmd.Params["net_mode"].(float64); ok {
		msg.NetMode = int(v)
	}
}

// fillVoice 填充语音播报
func (a *TF100Adapter) fillVoice(cmd *model.StandardCommand, msg *TF100Message) {
	if v, ok := cmd.Params["break"].(float64); ok {
		msg.PushType = int(v)
	}
	if v, ok := cmd.Params["len"].(float64); ok {
		msg.Count = int(v)
	}
}

// fillChargeAdjust 填充充电调整
func (a *TF100Adapter) fillChargeAdjust(cmd *model.StandardCommand, msg *TF100Message) {
	if v, ok := cmd.Params["port"].(float64); ok {
		msg.Port = int(v)
	}
	if v, ok := cmd.Params["mode"].(float64); ok {
		msg.Mode = int(v) // 0=加金额, 1=加电量, 2=加时间
	}
	if v, ok := cmd.Params["value"].(float64); ok {
		msg.Value = int(v)
	}
	if v, ok := cmd.Params["new_value"].(float64); ok {
		msg.NewValue = int(v)
	}
	if v, ok := cmd.Params["order_no"].(string); ok {
		msg.Orders = v
	}
}

// fillCardMode 填充刷卡模式
func (a *TF100Adapter) fillCardMode(cmd *model.StandardCommand, msg *TF100Message) {
	if v, ok := cmd.Params["pwd"].(string); ok {
		msg.Pwd = v
	}
	if v, ok := cmd.Params["sect"].(float64); ok {
		msg.Sect = int(v)
	}
	if v, ok := cmd.Params["model"].(float64); ok {
		msg.Model = int(v)
	}
	if v, ok := cmd.Params["bind_sw"].(float64); ok {
		msg.BindSw = int(v)
	}
	if v, ok := cmd.Params["lock_sw"].(float64); ok {
		msg.LockSw = int(v)
	}
	if v, ok := cmd.Params["lock_key"].(string); ok {
		msg.LockKey = v
	}
	if v, ok := cmd.Params["dec_money"].(float64); ok {
		msg.DecMoney = int(v)
	}
	if v, ok := cmd.Params["max_money"].(float64); ok {
		msg.MaxMoney = int(v)
	}
	if v, ok := cmd.Params["refund_sw"].(float64); ok {
		msg.RefundSw = int(v)
	}
}

// fillModifyCurrent 填充修改电流
func (a *TF100Adapter) fillModifyCurrent(cmd *model.StandardCommand, msg *TF100Message) {
	if v, ok := cmd.Params["port"].(float64); ok {
		msg.Port = int(v)
	}
	if v, ok := cmd.Params["current"].(float64); ok {
		msg.Current = int(v)
	}
}

// fillQRCode 填充二维码
func (a *TF100Adapter) fillQRCode(cmd *model.StandardCommand, msg *TF100Message) {
	if v, ok := cmd.Params["qr_str"].(string); ok {
		msg.QRStr = v
	}
	if v, ok := cmd.Params["qr_type"].(float64); ok {
		msg.QRType = int(v)
	}
}

// fillCommAddr 填充通信地址
func (a *TF100Adapter) fillCommAddr(cmd *model.StandardCommand, msg *TF100Message) {
	if v, ok := cmd.Params["ip"].(string); ok {
		msg.IP = v
	}
	if v, ok := cmd.Params["port"].(float64); ok {
		msg.Port = int(v)
	}
	if v, ok := cmd.Params["sockb"].(float64); ok {
		msg.SockB = int(v)
	}
}

// fillLockCmd 填充地锁指令
func (a *TF100Adapter) fillLockCmd(cmd *model.StandardCommand, msg *TF100Message) {
	if v, ok := cmd.Params["port"].(float64); ok {
		msg.Port = int(v)
	}
	if v, ok := cmd.Params["addr"].(float64); ok {
		msg.Addr = int(v)
	}
	if v, ok := cmd.Params["order_no"].(string); ok {
		msg.Orders = v
	}
}

// fillLockQuery 填充地锁查询
func (a *TF100Adapter) fillLockQuery(cmd *model.StandardCommand, msg *TF100Message) {
	if v, ok := cmd.Params["port"].(float64); ok {
		msg.Port = int(v)
	}
	if v, ok := cmd.Params["addr"].(float64); ok {
		msg.Addr = int(v)
	}
}

// fillLockAddrSet 填充地锁地址修改
func (a *TF100Adapter) fillLockAddrSet(cmd *model.StandardCommand, msg *TF100Message) {
	if v, ok := cmd.Params["port"].(float64); ok {
		msg.Port = int(v)
	}
	if v, ok := cmd.Params["src_addr"].(float64); ok {
		msg.SrcAddr = int(v)
	}
	if v, ok := cmd.Params["dst_addr"].(float64); ok {
		msg.DstAddr = int(v)
	}
}

// DecodeResponse 解析指令响应
func (a *TF100Adapter) DecodeResponse(raw []byte) (*model.StandardCommandResponse, error) {
	s := string(raw)

	if !strings.HasPrefix(s, "CCMD:") {
		return nil, fmt.Errorf("not a device response frame")
	}

	idx := strings.IndexByte(s, '{')
	if idx < 0 {
		return nil, fmt.Errorf("no JSON found")
	}

	msg := TF100Message{}
	if err := json.Unmarshal([]byte(s[idx:]), &msg); err != nil {
		return nil, fmt.Errorf("JSON parse error: %w", err)
	}

	resp := &model.StandardCommandResponse{
		DeviceID:  fmt.Sprintf("%d", msg.DevID),
		Timestamp: time.Now(),
	}

	switch msg.MsgType {
	case MsgStartCharge, MsgStopCharge, MsgChargeAdjust:
		resp.Success = msg.Ack == 0
		if !resp.Success {
			resp.Error = fmt.Sprintf("error code: %d", msg.Ack)
		}
		resp.Result = map[string]interface{}{
			"port":     msg.Port,
			"order_no": msg.Orders,
		}
	case MsgSetRate, MsgSetDeviceParam, MsgReboot, MsgClearStorage,
		MsgSetCardMode, MsgModifyCurrent, MsgSetQRCode, MsgSetCommAddr:
		resp.Success = msg.Ack == 0
		if !resp.Success {
			resp.Error = fmt.Sprintf("error code: %d", msg.Ack)
		}
	case MsgLockUp, MsgLockDown:
		resp.Success = msg.Result == 0
		if !resp.Success {
			resp.Error = fmt.Sprintf("lock error code: %d", msg.Result)
		}
		resp.Result = map[string]interface{}{
			"addr":     msg.Addr,
			"order_no": msg.Orders,
		}
	case MsgLockAddrSet:
		resp.Success = msg.Result == 0
		resp.Result = map[string]interface{}{
			"addr": msg.Addr,
		}
	default:
		// 默认判断: ack=0为成功
		resp.Success = msg.Ack == 0
	}

	return resp, nil
}
