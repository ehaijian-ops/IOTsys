package ap3000

import (
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"iot-platform/internal/protocol/engine"
	"iot-platform/internal/protocol/model"
)

// AP3000 电单车充电桩协议适配器
// 通信方式: TCP 长连接
// 帧格式: 帧头(DNY:0x44 0x4E 0x59) + 长度(2B) + 设备ID(4B) + 消息ID(2B) + 命令(1B) + 数据(nB) + 校验(2B)
// 长度 = 命令 + 数据 + 设备ID(4) + 消息ID(2) 的总字节数，最大256
// 校验 = 除帧头外的所有字节累加和取低16位
// 字节序: 默认小端模式
// 重发: 每包默认发1次，超时15秒重发

const (
	ProtocolName    = "AP3000_v2"
	ProtocolVersion = "8.6"
	DeviceType      = "ebike_charger"

	// 帧头 DNY
	FrameHeader0 = 0x44 // 'D'
	FrameHeader1 = 0x4E // 'N'
	FrameHeader2 = 0x59 // 'Y'
	HeaderLen    = 3
	LenLen       = 2 // 长度字段
	DevIDLen     = 4 // 设备ID长度
	MsgIDLen     = 2 // 消息ID长度
	CmdLen       = 1 // 命令字长度
	ChecksumLen  = 2 // 校验和长度

	// 帧固定开销: 帧头(3) + 长度(2) + 设备ID(4) + 消息ID(2) + 命令(1) + 校验(2) = 14
	FrameOverhead = HeaderLen + LenLen + DevIDLen + MsgIDLen + CmdLen + ChecksumLen

	// ---- 设备上报命令 ----
	CmdLogin          = 0x01 // 设备登录(旧版)
	CmdSwipeCard      = 0x02 // 刷卡/余额查询
	CmdChargeRecord   = 0x03 // 充电记录上传
	CmdPortConfirm    = 0x04 // 端口充电确认
	CmdDeviceInfo     = 0x05 // 设备信息上报
	CmdChargeProgress = 0x06 // 充电中实时数据
	CmdPhoneMode      = 0x09 // 手机模式专用
	CmdPhoneModule    = 0x0A // 手机模块连接地址
	CmdRegister       = 0x20 // 设备注册(新版)
	CmdStatusReport   = 0x21 // 设备状态上报
	CmdGetTime        = 0x22 // 获取服务器时间
	CmdChargeRecordV2 = 0x23 // 充电记录上传(分时计费版)
	CmdChargeProgressV2 = 0x26 // 充电中实时数据(分时计费版)
	CmdPortLock       = 0x41 // 端口锁专用
	CmdLockCmd        = 0x42 // 锁指令
	CmdLockNotify     = 0x43 // 锁充满通知
	CmdPortAlarm      = 0x44 // 端口报警

	// ---- 服务器下发命令 ----
	CmdQueryStatus      = 0x81 // 查询设备状态
	CmdStartStopCharge  = 0x82 // 启动/停止充电
	CmdSetRunParam11    = 0x83 // 设置运行参数1.1
	CmdSetRunParam12    = 0x84 // 设置运行参数1.2
	CmdSetChargeParam   = 0x85 // 设置充电时长/最大功率
	CmdSetUserPassword  = 0x86 // 设置用户密码
	CmdReboot           = 0x87 // 远程重启
	CmdClearStorage     = 0x88 // 清除存储满
	CmdVoiceBroadcast   = 0x89 // 语音播报
	CmdModifyChargeMode = 0x8A // 修改充电模式/时长/金额
	CmdReadEEPROM       = 0x8B // 读EEPROM
	CmdWriteEEPROM      = 0x8C // 写EEPROM
	CmdSetWorkMode      = 0x8D // 设置工作模式(刷卡/投币)
	CmdSetQRCode        = 0x8E // 修改二维码地址
	CmdSetTCMode        = 0x8F // 设置TC刷卡模式
	CmdQueryParam90     = 0x90 // 查询运行参数1.1
	CmdQueryParam91     = 0x91 // 查询运行参数1.2
	CmdQueryParam92     = 0x92 // 查询运行参数2
	CmdQueryParam93     = 0x93 // 查询用户密码
	CmdQueryParam94     = 0x94 // 查询充电参数
	CmdSetTempQR        = 0x95 // 临时二维码
	CmdFindDevice       = 0x96 // 查找设备(蜂鸣)
	CmdSetWorkMode2     = 0x97 // 设置工作模式(新)
	CmdForceOpen        = 0x98 // 强制开锁
	CmdOTA              = 0xE0 // OTA固件升级(手机)
	CmdOTA2             = 0xE4 // OTA固件升级(服务器)
	CmdOTAPower         = 0xE1 // OTA电源固件
	CmdOTAUnified       = 0xE2 // OTA统一升级
	CmdOTAOld           = 0xF8 // OTA旧版固件升级
	CmdStartStopV2      = 0xA2 // 启动/停止充电(分时计费版)
	CmdStopCharge72     = 0x72 // 停止充电(带订单号)
	CmdReserved         = 0xFD // 预留指令
	CmdReserved2        = 0xFE // 预留指令2

	// 设备标识码
	DevIDCode03 = 0x03 // 单路
	DevIDCode04 = 0x04 // 双路
	DevIDCode05 = 0x05 // 10路
	DevIDCode06 = 0x06 // 16路
	DevIDCode07 = 0x07 // 12路
	DevIDCode09 = 0x09 // 投币器
	DevIDCode0A = 0x0A // 漏电检测
)

type AP3000Adapter struct{}

func init() {
	engine.Register(&AP3000Adapter{})
}

func (a *AP3000Adapter) Name() string       { return ProtocolName }
func (a *AP3000Adapter) Version() string    { return ProtocolVersion }
func (a *AP3000Adapter) DeviceType() string { return DeviceType }

// Validate 校验报文是否为 AP3000 协议
func (a *AP3000Adapter) Validate(raw []byte) bool {
	if len(raw) < FrameOverhead {
		return false
	}
	// 校验帧头 DNY
	if raw[0] != FrameHeader0 || raw[1] != FrameHeader1 || raw[2] != FrameHeader2 {
		return false
	}
	// 校验长度: 长度字段 = 命令 + 数据 + 设备ID(4) + 消息ID(2) 的总字节数
	dataLen := int(binary.LittleEndian.Uint16(raw[HeaderLen : HeaderLen+LenLen]))
	totalExpected := HeaderLen + LenLen + dataLen + ChecksumLen
	if totalExpected > len(raw) {
		return false
	}
	// 校验累加和
	return a.verifyChecksum(raw[:totalExpected])
}

// calcChecksum 计算累加和校验 (除帧头外的所有字节累加，取低16位)
func (a *AP3000Adapter) calcChecksum(data []byte) uint16 {
	var sum uint32
	for _, b := range data {
		sum += uint32(b)
	}
	return uint16(sum & 0xFFFF)
}

// verifyChecksum 验证累加和
func (a *AP3000Adapter) verifyChecksum(frame []byte) bool {
	if len(frame) < HeaderLen+ChecksumLen {
		return false
	}
	// 校验范围: 从长度字段到数据结束
	payloadEnd := len(frame) - ChecksumLen
	expected := binary.LittleEndian.Uint16(frame[payloadEnd:])
	actual := a.calcChecksum(frame[HeaderLen:payloadEnd])
	return expected == actual
}

// Decode 解析 AP3000 数据上报帧 → 标准数据模型
func (a *AP3000Adapter) Decode(raw []byte) (*model.StandardData, error) {
	if !a.Validate(raw) {
		return nil, fmt.Errorf("invalid AP3000 frame")
	}

	// 长度字段
	dataLen := int(binary.LittleEndian.Uint16(raw[HeaderLen : HeaderLen+LenLen]))
	// 设备ID (4字节, 小端)
	devID := binary.LittleEndian.Uint32(raw[HeaderLen+LenLen : HeaderLen+LenLen+DevIDLen])
	// 消息ID (2字节, 小端)
	msgID := int(binary.LittleEndian.Uint16(raw[HeaderLen+LenLen+DevIDLen : HeaderLen+LenLen+DevIDLen+MsgIDLen]))
	// 命令字
	cmdOffset := HeaderLen + LenLen + DevIDLen + MsgIDLen
	cmd := raw[cmdOffset]
	// 数据部分
	dataOffset := cmdOffset + CmdLen
	dataEnd := HeaderLen + LenLen + dataLen
	payload := raw[dataOffset:dataEnd]

	std := &model.StandardData{
		DeviceID:  fmt.Sprintf("%d", devID),
		Protocol:  ProtocolName,
		Timestamp: time.Now(),
		MsgID:     msgID,
		Extra:     make(map[string]interface{}),
	}

	var err error
	switch cmd {
	case CmdLogin: // 0x01 设备登录(旧版)
		err = a.parseLogin(payload, std)
	case CmdSwipeCard: // 0x02 刷卡
		err = a.parseSwipeCard(payload, std)
	case CmdChargeRecord: // 0x03 充电记录
		err = a.parseChargeRecord(payload, std)
	case CmdPortConfirm: // 0x04 端口充电确认
		err = a.parsePortConfirm(payload, std)
	case CmdChargeProgress: // 0x06 充电中实时数据
		err = a.parseChargeProgress(payload, std)
	case CmdRegister: // 0x20 设备注册
		err = a.parseRegister(payload, std)
	case CmdStatusReport: // 0x21 设备状态上报
		err = a.parseStatusReport(payload, std)
	case CmdGetTime: // 0x22 获取时间
		std.Extra["cmd"] = "get_time"
	case CmdChargeRecordV2: // 0x23 充电记录(分时计费)
		err = a.parseChargeRecordV2(payload, std)
	case CmdChargeProgressV2: // 0x26 充电中(分时计费)
		err = a.parseChargeProgressV2(payload, std)
	case CmdLockNotify: // 0x43 锁充满通知
		err = a.parseChargeRecord(payload, std) // 数据格式同03
	case CmdPortAlarm: // 0x44 端口报警
		err = a.parsePortAlarm(payload, std)
	default:
		std.Extra["raw_cmd"] = fmt.Sprintf("0x%02X", cmd)
		std.Extra["raw_data"] = fmt.Sprintf("%X", payload)
	}

	return std, err
}

// parseLogin 解析0x01登录包
func (a *AP3000Adapter) parseLogin(data []byte, std *model.StandardData) error {
	if len(data) < 1 {
		return fmt.Errorf("login data too short")
	}
	offset := 0

	// 设备类型 (1字节)
	std.DeviceTypeCode = int(data[offset])
	offset++

	// 固件版本 (2字节, 小端, 100=V1.00)
	if offset+2 <= len(data) {
		fwVer := binary.LittleEndian.Uint16(data[offset : offset+2])
		std.FirmwareVer = fmt.Sprintf("V%d.%02d", fwVer/100, fwVer%100)
		offset += 2
	}

	// 电压 (2字节, 0.1V)
	if offset+2 <= len(data) {
		std.Voltage = float64(binary.LittleEndian.Uint16(data[offset:offset+2])) * 0.1
		offset += 2
	}

	// 端口数量 (1字节)
	if offset < len(data) {
		std.PortCount = int(data[offset])
		offset++
	}

	// 各端口状态 (n字节)
	if offset+std.PortCount <= len(data) {
		std.PortData = make([]model.PortData, std.PortCount)
		for i := 0; i < std.PortCount; i++ {
			std.PortData[i] = model.PortData{
				PortIndex:  i + 1,
				PortStatus: a.parsePortStatus(data[offset+i]),
			}
		}
		offset += std.PortCount
	}

	// 各端口当前功率 (n*2字节, 0.1W)
	if offset+std.PortCount*2 <= len(data) {
		for i := 0; i < std.PortCount; i++ {
			std.PortData[i].Power = float64(binary.LittleEndian.Uint16(data[offset:offset+2])) * 0.1
			offset += 2
		}
	}

	// 各端口峰值功率 (n*2字节, 0.1W)
	if offset+std.PortCount*2 <= len(data) {
		for i := 0; i < std.PortCount; i++ {
			std.PortData[i].PeakPower = float64(binary.LittleEndian.Uint16(data[offset:offset+2])) * 0.1
			offset += 2
		}
	}

	// 节点ID (1字节)
	if offset < len(data) {
		std.Extra["node_id"] = int(data[offset])
		offset++
	}

	// 信号强度 (1字节, 0-31)
	if offset < len(data) {
		std.SignalStrength = int(data[offset])
		offset++
	}

	// 设备类型(扩展, 1字节)
	if offset < len(data) {
		if std.DeviceTypeCode == 0 {
			std.DeviceTypeCode = int(data[offset])
		}
		offset++
	}

	// 当前环境温度 (1字节, 减65)
	if offset < len(data) && data[offset] != 0 {
		std.Temperature = float64(data[offset]) - 65
		offset++
	}

	// 工作模式 (1字节)
	if offset < len(data) {
		std.Extra["work_mode"] = int(data[offset])
	}

	return nil
}

// parseRegister 解析0x20注册包
func (a *AP3000Adapter) parseRegister(data []byte, std *model.StandardData) error {
	if len(data) < 1 {
		return fmt.Errorf("register data too short")
	}
	offset := 0

	// 设备类型 (1字节)
	std.DeviceTypeCode = int(data[offset])
	offset++

	// 固件版本 (2字节)
	if offset+2 <= len(data) {
		fwVer := binary.LittleEndian.Uint16(data[offset : offset+2])
		std.FirmwareVer = fmt.Sprintf("V%d.%02d", fwVer/100, fwVer%100)
		offset += 2
	}

	// 端口数量 (1字节)
	if offset < len(data) {
		std.PortCount = int(data[offset])
		offset++
	}

	// 节点ID (1字节)
	if offset < len(data) {
		std.Extra["node_id"] = int(data[offset])
		offset++
	}

	// 设备类型(扩展, 1字节)
	if offset < len(data) {
		if std.DeviceTypeCode == 0 {
			std.DeviceTypeCode = int(data[offset])
		}
		offset++
	}

	// 工作模式 (1字节)
	if offset < len(data) {
		std.Extra["work_mode"] = int(data[offset])
		offset++
	}

	// 电源固件版本 (2字节)
	if offset+2 <= len(data) {
		pwVer := binary.LittleEndian.Uint16(data[offset : offset+2])
		std.Extra["power_fw_ver"] = fmt.Sprintf("V%d.%02d", pwVer/100, pwVer%100)
		offset += 2
	}

	// 设备计时计费功能 (1字节)
	if offset < len(data) {
		std.Extra["timing_charge"] = int(data[offset])
		offset++
	}

	// TC模式 (1字节)
	if offset < len(data) {
		std.Extra["tc_mode"] = int(data[offset])
	}

	return nil
}

// parseStatusReport 解析0x21状态上报
func (a *AP3000Adapter) parseStatusReport(data []byte, std *model.StandardData) error {
	if len(data) < 4 {
		return fmt.Errorf("status data too short")
	}
	offset := 0

	// 电压 (2字节, 0.1V)
	std.Voltage = float64(binary.LittleEndian.Uint16(data[offset:offset+2])) * 0.1
	offset += 2

	// 端口数量 (1字节)
	std.PortCount = int(data[offset])
	offset++

	// 各端口状态 (n字节)
	if offset+std.PortCount <= len(data) {
		std.PortData = make([]model.PortData, std.PortCount)
		for i := 0; i < std.PortCount; i++ {
			std.PortData[i] = model.PortData{
				PortIndex:  i + 1,
				PortStatus: a.parsePortStatus(data[offset+i]),
			}
		}
		offset += std.PortCount
	}

	// 信号强度 (1字节)
	if offset < len(data) {
		std.SignalStrength = int(data[offset])
		offset++
	}

	// 当前环境温度 (1字节, 减65)
	if offset < len(data) && data[offset] != 0 {
		std.Temperature = float64(data[offset]) - 65
	}

	return nil
}

// parseSwipeCard 解析0x02刷卡数据
func (a *AP3000Adapter) parseSwipeCard(data []byte, std *model.StandardData) error {
	if len(data) < 7 {
		return fmt.Errorf("swipe card data too short")
	}
	offset := 0

	// 卡ID (4字节, 大端)
	cardID := binary.BigEndian.Uint32(data[offset : offset+4])
	std.CardID = fmt.Sprintf("%d", cardID)
	offset += 4

	// 卡类型 (1字节) 0=旧卡 1=新卡 2=金额卡
	std.CardType = int(data[offset])
	offset++

	// 端口号 (1字节)
	std.PortNum = int(data[offset])
	offset++

	// 金额卡余额 (2字节, 分)
	if offset+2 <= len(data) {
		std.CardMoney = float64(binary.LittleEndian.Uint16(data[offset:offset+2])) / 100.0
	}

	return nil
}

// parseChargeRecord 解析0x03充电记录
func (a *AP3000Adapter) parseChargeRecord(data []byte, std *model.StandardData) error {
	if len(data) < 10 {
		return fmt.Errorf("charge record too short")
	}
	offset := 0

	// 充电时长 (2字节, 秒)
	std.ChargeTime = int(binary.LittleEndian.Uint16(data[offset : offset+2]))
	offset += 2

	// 峰值功率 (2字节, 0.1W)
	std.PeakPower = float64(binary.LittleEndian.Uint16(data[offset:offset+2])) * 0.1
	offset += 2

	// 耗电量 (2字节, 0.01度)
	std.ChargeEnergy = float64(binary.LittleEndian.Uint16(data[offset:offset+2])) * 0.01
	offset += 2

	// 端口号 (1字节)
	std.PortNum = int(data[offset])
	offset++

	// 启动/验证方式 (1字节) 0=刷卡启动 1=远程启动 3=验证码
	std.StartType = int(data[offset])
	offset++

	// 卡号/验证码 (4字节)
	std.VerifyCode = fmt.Sprintf("%d", binary.LittleEndian.Uint32(data[offset:offset+4]))
	offset += 4

	// 停止原因 (1字节)
	if offset < len(data) {
		std.StopReason = int(data[offset])
		offset++
	}

	// 订单号 (16字节)
	if offset+16 <= len(data) {
		std.OrderNo = strings.TrimRight(string(data[offset:offset+16]), "\x00")
		offset += 16
	}

	// 第二峰值功率 (2字节)
	if offset+2 <= len(data) {
		std.Extra["second_peak_power"] = float64(binary.LittleEndian.Uint16(data[offset:offset+2])) * 0.1
		offset += 2
	}

	// 时间戳 (4字节, Unix)
	if offset+4 <= len(data) {
		ts := int64(binary.LittleEndian.Uint32(data[offset : offset+4]))
		std.Timestamp = time.Unix(ts, 0)
	}

	return nil
}

// parsePortConfirm 解析0x04端口充电确认
func (a *AP3000Adapter) parsePortConfirm(data []byte, std *model.StandardData) error {
	if len(data) < 8 {
		return fmt.Errorf("port confirm data too short")
	}
	offset := 0

	// 端口号 (1字节)
	std.PortNum = int(data[offset])
	offset++

	// 启动方式 (1字节)
	std.StartType = int(data[offset])
	offset++

	// 卡ID (4字节)
	cardID := binary.BigEndian.Uint32(data[offset : offset+4])
	std.CardID = fmt.Sprintf("%d", cardID)
	offset += 4

	// 已充电时长 (2字节, 秒)
	std.ChargeTime = int(binary.LittleEndian.Uint16(data[offset : offset+2]))

	return nil
}

// parseChargeProgress 解析0x06充电中实时数据
func (a *AP3000Adapter) parseChargeProgress(data []byte, std *model.StandardData) error {
	if len(data) < 12 {
		return fmt.Errorf("charge progress data too short")
	}
	offset := 0

	// 端口号 (1字节)
	std.PortNum = int(data[offset])
	offset++

	// 端口状态 (1字节)
	portStatus := a.parsePortStatus(data[offset])
	offset++

	// 充电时长 (2字节, 秒)
	std.ChargeTime = int(binary.LittleEndian.Uint16(data[offset : offset+2]))
	offset += 2

	// 累计电量 (2字节, 0.01度)
	std.ChargeEnergy = float64(binary.LittleEndian.Uint16(data[offset:offset+2])) * 0.01
	offset += 2

	// 启动方式 (1字节)
	std.StartType = int(data[offset])
	offset++

	// 实时功率 (2字节, 0.1W)
	std.Power = float64(binary.LittleEndian.Uint16(data[offset:offset+2])) * 0.1
	offset += 2

	// 周期最大功率 (2字节)
	offset += 2

	// 周期最小功率 (2字节)
	offset += 2

	// 周期平均功率 (2字节)
	offset += 2

	// 订单号 (16字节)
	if offset+16 <= len(data) {
		std.OrderNo = strings.TrimRight(string(data[offset:offset+16]), "\x00")
		offset += 16
	}

	// 定时充电耗电量 (2字节)
	offset += 2

	// 峰值功率 (2字节, 0.1W)
	if offset+2 <= len(data) {
		std.PeakPower = float64(binary.LittleEndian.Uint16(data[offset:offset+2])) * 0.1
		offset += 2
	}

	// 电压 (2字节, 0.1V)
	if offset+2 <= len(data) {
		std.Voltage = float64(binary.LittleEndian.Uint16(data[offset:offset+2])) * 0.1
		offset += 2
	}

	// 电流 (2字节, 0.001A)
	if offset+2 <= len(data) {
		std.Current = float64(binary.LittleEndian.Uint16(data[offset:offset+2])) * 0.001
		offset += 2
	}

	// 环境温度 (1字节, 减65)
	if offset < len(data) && data[offset] != 0 {
		std.Temperature = float64(data[offset]) - 65
		offset++
	}

	// 端口温度 (1字节, 减65)
	if offset < len(data) {
		std.Extra["port_temp"] = float64(data[offset]) - 65
	}

	// 构建端口数据
	std.PortData = []model.PortData{{
		PortIndex:  std.PortNum,
		PortStatus: portStatus,
		Power:      std.Power,
	}}

	return nil
}

// parseChargeRecordV2 解析0x23充电记录(分时计费版)
func (a *AP3000Adapter) parseChargeRecordV2(data []byte, std *model.StandardData) error {
	if err := a.parseChargeRecord(data, std); err != nil {
		return err
	}
	// 附加字段: 占用时间、总费用、电费、服务费等
	// 这里解析基本框架，具体字段根据协议版本调整
	return nil
}

// parseChargeProgressV2 解析0x26充电中(分时计费版)
func (a *AP3000Adapter) parseChargeProgressV2(data []byte, std *model.StandardData) error {
	if err := a.parseChargeProgress(data, std); err != nil {
		return err
	}
	return nil
}

// parsePortAlarm 解析0x44端口报警
func (a *AP3000Adapter) parsePortAlarm(data []byte, std *model.StandardData) error {
	if len(data) < 3 {
		return fmt.Errorf("port alarm data too short")
	}
	std.PortNum = int(data[0])
	std.Extra["alarm_type"] = int(data[1])
	std.FaultCode = fmt.Sprintf("ALARM_%d", data[1])
	std.FaultMessage = a.getAlarmMessage(int(data[1]))
	return nil
}

// parsePortStatus 解析端口状态
func (a *AP3000Adapter) parsePortStatus(status byte) string {
	switch status {
	case 0x00:
		return "idle" // 空闲
	case 0x01:
		return "charging" // 充电中
	case 0x02:
		return "scanned" // 已扫码未充电
	case 0x03:
		return "scanned_charged" // 已扫码已充电
	case 0x04:
		return "short_circuit" // 短路
	case 0x05:
		return "floating" // 浮充
	case 0x06:
		return "storage_full" // 存储满坏
	case 0x07:
		return "chips_stuck" // 计量芯片卡住
	case 0x08:
		return "contactor_open" // 接触器/空开端
	case 0x09:
		return "relay_stuck" // 继电器粘连
	case 0x0A:
		return "meter_broken" // 计量器坏
	case 0x0B:
		return "relay_no_close" // 继电器无法闭合
	case 0x0D:
		return "short_circuit_pre" // 短路预警
	case 0x0E:
		return "relay_stuck_pre" // 继电器粘连预警
	case 0x0F:
		return "swipe_chip_broken" // 刷卡芯片坏
	case 0x10:
		return "line_short" // 线路短路
	default:
		return fmt.Sprintf("unknown_0x%02X", status)
	}
}

// getAlarmMessage 获取告警信息
func (a *AP3000Adapter) getAlarmMessage(alarmType int) string {
	messages := map[int]string{
		1: "门未关好",
	}
	if msg, ok := messages[alarmType]; ok {
		return msg
	}
	return fmt.Sprintf("未知告警_%d", alarmType)
}

// getStopReasonMessage 获取停止原因描述
func (a *AP3000Adapter) getStopReasonMessage(reason int) string {
	messages := map[int]string{
		0:  "正常充满",
		1:  "手动停止",
		2:  "达到最大充电时间",
		3:  "达到预设时间",
		4:  "达到预设电量",
		5:  "用户拔枪",
		6:  "过载保护",
		7:  "远程停止",
		8:  "动态负载",
		9:  "功率过小",
		0x0A: "环境温度过高",
		0x0B: "端口温度过高",
		0x0C: "未检测到电池",
		0x0D: "用户拔枪-1",
		0x0E: "无功率停止",
		0x0F: "继电器无法闭合预警",
		0x10: "水浸断电",
		0x11: "过载(本端口)",
		0x12: "过载(非本端口)",
		0x13: "用户刷卡断电",
		0x14: "未关好门",
		0x15: "外部强制停止",
		0x16: "刷卡强制停止",
		0x17: "紧急强制停止",
		0x18: "系统错误停止",
		0x19: "存储故障",
		0x1A: "过压",
		0x1B: "欠压",
		0x1C: "低功率断电",
	}
	if msg, ok := messages[reason]; ok {
		return msg
	}
	return fmt.Sprintf("未知原因_%d", reason)
}

// Encode 编码标准指令 → AP3000 协议帧
func (a *AP3000Adapter) Encode(cmd *model.StandardCommand) ([]byte, error) {
	// 解析设备ID
	devID := uint32(0)
	fmt.Sscanf(cmd.DeviceID, "%d", &devID)

	// 生成消息ID (简单递增)
	msgID := uint16(time.Now().UnixNano() % 65535)

	var cmdByte byte
	var payload []byte
	var err error

	switch cmd.CmdType {
	case "query_status":
		cmdByte = CmdQueryStatus
		payload = []byte{}
	case "start_charge":
		cmdByte = CmdStartStopCharge
		payload, err = a.encodeStartCharge(cmd)
	case "stop_charge":
		cmdByte = CmdStartStopCharge
		payload, err = a.encodeStopCharge(cmd)
	case "modify_charge":
		cmdByte = CmdModifyChargeMode
		payload, err = a.encodeModifyCharge(cmd)
	case "config":
		cmdByte = CmdSetRunParam11
		payload, err = a.encodeConfig(cmd)
	case "reboot":
		cmdByte = CmdReboot
		payload = []byte{}
	case "clear_storage":
		cmdByte = CmdClearStorage
		payload = []byte{}
	case "set_work_mode":
		cmdByte = CmdSetWorkMode
		payload, err = a.encodeWorkMode(cmd)
	case "set_tc_mode":
		cmdByte = CmdSetTCMode
		payload, err = a.encodeTCMode(cmd)
	case "voice":
		cmdByte = CmdVoiceBroadcast
		payload, err = a.encodeVoice(cmd)
	default:
		return nil, fmt.Errorf("unsupported command type: %s", cmd.CmdType)
	}

	if err != nil {
		return nil, err
	}

	return a.buildFrame(devID, msgID, cmdByte, payload), nil
}

// buildFrame 构建完整的AP3000协议帧
func (a *AP3000Adapter) buildFrame(devID uint32, msgID uint16, cmd byte, data []byte) []byte {
	// 数据长度 = 设备ID(4) + 消息ID(2) + 命令(1) + 数据(n)
	dataLen := DevIDLen + MsgIDLen + CmdLen + len(data)

	// 帧结构: 帧头(3) + 长度(2) + 设备ID(4) + 消息ID(2) + 命令(1) + 数据(n) + 校验(2)
	frame := make([]byte, HeaderLen+LenLen+dataLen+ChecksumLen)

	// 帧头 DNY
	frame[0] = FrameHeader0
	frame[1] = FrameHeader1
	frame[2] = FrameHeader2

	// 长度 (小端)
	binary.LittleEndian.PutUint16(frame[HeaderLen:], uint16(dataLen))

	// 设备ID (小端)
	binary.LittleEndian.PutUint32(frame[HeaderLen+LenLen:], devID)

	// 消息ID (小端)
	binary.LittleEndian.PutUint16(frame[HeaderLen+LenLen+DevIDLen:], msgID)

	// 命令
	frame[HeaderLen+LenLen+DevIDLen+MsgIDLen] = cmd

	// 数据
	copy(frame[HeaderLen+LenLen+DevIDLen+MsgIDLen+CmdLen:], data)

	// 校验 (累加和, 从长度字段开始)
	payloadEnd := HeaderLen + LenLen + dataLen
	checksum := a.calcChecksum(frame[HeaderLen:payloadEnd])
	binary.LittleEndian.PutUint16(frame[payloadEnd:], checksum)

	return frame
}

// encodeStartCharge 编码启动充电指令
func (a *AP3000Adapter) encodeStartCharge(cmd *model.StandardCommand) ([]byte, error) {
	data := make([]byte, 26) // 最小82指令长度

	// 充电模式 (1字节): 0=按时, 1=按量, 2=按金额, 3=按次
	mode := byte(0)
	if v, ok := cmd.Params["charge_mode"].(float64); ok {
		mode = byte(v)
	}
	data[0] = mode

	// 金额/有效期 (4字节, 小端)
	if v, ok := cmd.Params["amount"].(float64); ok {
		binary.LittleEndian.PutUint32(data[1:5], uint32(v))
	}

	// 端口号 (1字节), 0xFF=全部
	port := byte(0xFF)
	if v, ok := cmd.Params["port"].(float64); ok {
		port = byte(v - 1) // 端口从0开始
	}
	data[5] = port

	// 操作类型 (1字节): 0=停止, 1=开始
	action := byte(1)
	if v, ok := cmd.Params["action"].(float64); ok {
		action = byte(v)
	}
	data[6] = action

	// 充电时长/电量 (2字节, 小端)
	if v, ok := cmd.Params["charge_time"].(float64); ok {
		binary.LittleEndian.PutUint16(data[7:9], uint16(v))
	}

	// 订单号 (16字节)
	if orderNo, ok := cmd.Params["order_no"].(string); ok {
		copy(data[9:25], []byte(orderNo))
	}

	return data, nil
}

// encodeStopCharge 编码停止充电指令
func (a *AP3000Adapter) encodeStopCharge(cmd *model.StandardCommand) ([]byte, error) {
	data := make([]byte, 26)

	data[0] = 0   // 模式
	data[5] = 0xFF // 端口
	if v, ok := cmd.Params["port"].(float64); ok {
		data[5] = byte(v - 1)
	}
	data[6] = 0 // 停止

	// 订单号
	if orderNo, ok := cmd.Params["order_no"].(string); ok {
		copy(data[9:25], []byte(orderNo))
	}

	return data, nil
}

// encodeModifyCharge 编码修改充电参数
func (a *AP3000Adapter) encodeModifyCharge(cmd *model.StandardCommand) ([]byte, error) {
	data := make([]byte, 5)

	// 模式 (1字节): 0=按时间(不停止), 1=按时间(停止), 2=按电量(不停止)
	if v, ok := cmd.Params["mode"].(float64); ok {
		data[0] = byte(v)
	}

	// 端口号 (1字节)
	if v, ok := cmd.Params["port"].(float64); ok {
		data[1] = byte(v - 1)
	}

	// 时间/电量/金额 (2字节, 小端)
	if v, ok := cmd.Params["value"].(float64); ok {
		binary.LittleEndian.PutUint16(data[2:4], uint16(v))
	}

	return data, nil
}

// encodeConfig 编码参数配置
func (a *AP3000Adapter) encodeConfig(cmd *model.StandardCommand) ([]byte, error) {
	data := make([]byte, 12)

	// 拔枪识别功率 (2字节, 0.1W)
	if v, ok := cmd.Params["unplug_power"].(float64); ok {
		binary.LittleEndian.PutUint16(data[0:2], uint16(v/0.1))
	}

	// 拔枪识别时间 (2字节, 秒)
	if v, ok := cmd.Params["unplug_time"].(float64); ok {
		binary.LittleEndian.PutUint16(data[2:4], uint16(v))
	}

	return data, nil
}

// encodeWorkMode 编码工作模式
func (a *AP3000Adapter) encodeWorkMode(cmd *model.StandardCommand) ([]byte, error) {
	data := make([]byte, 1)
	if v, ok := cmd.Params["mode"].(float64); ok {
		data[0] = byte(v) // 0=投币, 1=刷卡
	}
	return data, nil
}

// encodeTCMode 编码TC刷卡模式
func (a *AP3000Adapter) encodeTCMode(cmd *model.StandardCommand) ([]byte, error) {
	data := make([]byte, 1)
	if v, ok := cmd.Params["mode"].(float64); ok {
		data[0] = byte(v) // 0=按时, 1=计次, 3=计数, 4=一元
	}
	return data, nil
}

// encodeVoice 编码语音播报
func (a *AP3000Adapter) encodeVoice(cmd *model.StandardCommand) ([]byte, error) {
	data := make([]byte, 3)
	// 是否续播 (1字节)
	if v, ok := cmd.Params["continuous"].(float64); ok {
		data[0] = byte(v)
	}
	// 语音数量 (1字节)
	if v, ok := cmd.Params["count"].(float64); ok {
		data[1] = byte(v)
	}
	// 语音内容 (n字节)
	if content, ok := cmd.Params["content"].(string); ok {
		copy(data[2:], []byte(content))
	}
	return data, nil
}

// DecodeResponse 解析指令响应
func (a *AP3000Adapter) DecodeResponse(raw []byte) (*model.StandardCommandResponse, error) {
	if !a.Validate(raw) {
		return nil, fmt.Errorf("invalid response frame")
	}

	dataLen := int(binary.LittleEndian.Uint16(raw[HeaderLen : HeaderLen+LenLen]))
	cmdOffset := HeaderLen + LenLen + DevIDLen + MsgIDLen
	cmd := raw[cmdOffset]
	dataOffset := cmdOffset + CmdLen
	payloadEnd := HeaderLen + LenLen + dataLen

	resp := &model.StandardCommandResponse{
		Timestamp: time.Now(),
		Success:   false,
	}

	// 检查响应码: 多数命令的响应只有1字节(0=成功)
	if dataOffset < payloadEnd {
		resultCode := raw[dataOffset]
		resp.Success = resultCode == 0x00
		if !resp.Success {
			resp.Error = fmt.Sprintf("error code: 0x%02X", resultCode)
		}

		// 82指令响应特殊处理
		if cmd == CmdStartStopCharge && payloadEnd-dataOffset >= 20 {
			resp.Result = map[string]interface{}{
				"order_no": strings.TrimRight(string(raw[dataOffset+1:dataOffset+17]), "\x00"),
				"port":     int(raw[dataOffset+17]),
			}
		}
	}

	return resp, nil
}
