package wsd

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"iot-platform/internal/protocol/engine"
	"iot-platform/internal/protocol/model"
)

// 微小电 (WeiShaoDian) 12路电单车充电桩协议适配器
// 通信方式: TCP 长连接
// 帧格式: SOP(1) | LEN(1) | CMD(1) | SESSION_ID(6) | DATA(N) | SUM(1)
// LEN = CMD(1) + SESSION_ID(6) + DATA(N) + SUM(1)
// SUM = LEN ^ CMD ^ SESSION_ID[0:6] ^ DATA[0:N] 异或值
// 字节序: 全部小端模式
// 设备登录时 SESSION_ID = "000000"，平台回复带回实际 SESSION_ID

const (
	ProtocolName    = "WSD_v1"
	ProtocolVersion = "1.0"
	DeviceType      = "ebike_charger"

	// 帧结构
	SOP          = 0xEE
	CmdLen       = 1
	SessionIDLen = 6
	SumLen       = 1
	// HeaderOverhead: SOP(1) + LEN(1)
	HeaderOverhead = 2
	// MinFrameLen: SOP(1) + LEN(1) + CMD(1) + SESSION_ID(6) + SUM(1) = 10 (无数据时)
	MinFrameLen = HeaderOverhead + CmdLen + SessionIDLen + SumLen // 10

	// ==================== 设备 → 平台 ====================
	CmdLogin          = 0xA0 // 设备登录
	CmdRemoteCtrlAck  = 0xA3 // 远程控制应答
	CmdHeartbeat      = 0xA4 // 心跳上传
	CmdHBIntervalAck  = 0xA7 // 心跳间隔设置应答
	CmdTimeRequest    = 0xA8 // 对时请求
	CmdAllPortsReply  = 0xB1 // 所有端口状态回复
	CmdOnePortReply   = 0xB3 // 某个端口状态回复
	CmdLocalCharge    = 0xB4 // 本地充电上报
	CmdLocalChargeRsp = 0xB5 // 本地充电上报应答(实际平台发，但设备也可能确认)
	CmdSwipeCard      = 0xB6 // 在线卡刷卡上报
	CmdRemoteStartAck = 0xB8 // 远程启动应答
	CmdRemoteStopAck  = 0xBA // 远程停止应答
	CmdSettlement     = 0xBB // 结算订单上传
	CmdFaultReport    = 0xC0 // 故障上报
	CmdGearReport     = 0xC2 // 充电分档上报
	CmdParamSetAck    = 0xC4 // 参数配置应答
	CmdConfigUpload   = 0xC6 // 上传配置表
	CmdPlatformParam  = 0xC7 // 获取平台参数表
	CmdQueryCard      = 0xD0 // 查询在线卡

	// ==================== 平台 → 设备 ====================
	CmdLoginAck      = 0xA1 // 登录应答(含对时)
	CmdRemoteCtrl    = 0xA2 // 远程控制(复位/版本更新)
	CmdHeartbeatAck  = 0xA5 // 心跳应答
	CmdSetHBInterval = 0xA6 // 心跳间隔设置
	CmdTimeAck       = 0xA9 // 对时应答
	CmdQueryAllPorts = 0xB0 // 读取所有端口状态
	CmdQueryOnePort  = 0xB2 // 读取某个端口状态
	CmdChargeAck     = 0xB5 // 本地充电上报应答
	CmdRemoteStart   = 0xB7 // 远程启动
	CmdRemoteStop    = 0xB9 // 远程停止
	CmdSettlementAck = 0xBC // 结算订单应答
	CmdCardInvalid   = 0xBD // 卡无效/余额不足
	CmdFaultAck      = 0xC1 // 故障上报应答
	CmdSetParam      = 0xC3 // 设置参数表
	CmdReadParam     = 0xC5 // 读取参数
	CmdCardReply     = 0xD1 // 返回卡余额与状态

	// 端口状态码
	PortIdle            = 0x00 // 空闲
	PortCharging        = 0x01 // 充电中
	PortScanned         = 0x02 // 已扫码未充电
	PortScannedCharged  = 0x03 // 已扫码已充电
	PortShortCircuit    = 0x04 // 短路
	PortFloating        = 0x05 // 浮充
	PortFullStop        = 0x06 // 充满自停
	PortChipStuck       = 0x07 // 计量芯片卡住
	PortRelayStuck      = 0x09 // 继电器粘连
	PortRelayNoClose    = 0x0B // 继电器无法闭合
	PortOverTemp        = 0x0E // 过温保护
	PortReserved        = 0xFF // 预留/未使用
)

// WSDAdapter 微小电协议适配器
type WSDAdapter struct{}

func init() {
	engine.Register(&WSDAdapter{})
}

// ========== 接口实现 ==========

func (a *WSDAdapter) Name() string       { return ProtocolName }
func (a *WSDAdapter) Version() string    { return ProtocolVersion }
func (a *WSDAdapter) DeviceType() string { return DeviceType }

// Validate 校验报文是否为微小电协议帧
func (a *WSDAdapter) Validate(raw []byte) bool {
	if len(raw) < MinFrameLen {
		return false
	}
	if raw[0] != SOP {
		return false
	}
	frameLen := int(raw[1])
	if frameLen < CmdLen+SessionIDLen+SumLen { // at least CMD+SID+SUM = 8
		return false
	}
	totalExpected := HeaderOverhead + frameLen
	if totalExpected > len(raw) {
		return false
	}
	return a.verifyXOR(raw[:totalExpected])
}

// calcXOR 计算异或校验值：LEN ^ CMD ^ SESSION_ID ^ DATA
func (a *WSDAdapter) calcXOR(data []byte) byte {
	var sum byte
	for _, b := range data {
		sum ^= b
	}
	return sum
}

// verifyXOR 验证异或校验
// 帧结构: [0]=SOP [1]=LEN [2]=CMD [3:9]=SESSION_ID ... [1+LEN]=SUM
// XOR 覆盖 frame[1 : 1+LEN] (LEN,CMD,SESSION_ID,DATA), 校验字在 frame[1+LEN]
func (a *WSDAdapter) verifyXOR(frame []byte) bool {
	frameLen := int(frame[1])
	xorEnd := 1 + frameLen // position of SUM byte (exclusive for slice)
	if xorEnd < 9 || xorEnd >= len(frame) {
		return false
	}
	payload := frame[1:xorEnd] // indices 1 .. LEN inclusive
	expected := frame[xorEnd]  // SUM at index 1+LEN
	actual := a.calcXOR(payload)
	return expected == actual
}

// getDataPayload 从帧中提取 DATA 段和 SESSION_ID
func (a *WSDAdapter) getDataPayload(raw []byte) (sessionID []byte, data []byte, cmd byte) {
	frameLen := int(raw[1])
	cmd = raw[2]
	sessionID = raw[3:9] // 6 bytes
	// DATA at indices [9, 1+LEN), SUM at index 1+LEN
	dataEnd := 1 + frameLen // position of SUM byte
	if dataEnd > 9 {
		data = raw[9:dataEnd]
	}
	return
}

// ========== Decode 主入口 ==========

// Decode 解析微小电数据上报帧 → 标准数据模型
func (a *WSDAdapter) Decode(raw []byte) (*model.StandardData, error) {
	if !a.Validate(raw) {
		return nil, fmt.Errorf("invalid WSD frame")
	}

	_, data, cmd := a.getDataPayload(raw)

	std := &model.StandardData{
		Protocol:  ProtocolName,
		Timestamp: time.Now(),
		Extra:     make(map[string]interface{}),
	}

	var err error
	switch cmd {
	case CmdLogin: // 0xA0
		err = a.parseLogin(data, std)
	case CmdHeartbeat: // 0xA4
		err = a.parseHeartbeat(data, std)
	case CmdRemoteCtrlAck: // 0xA3
		err = a.parseRemoteCtrlAck(data, std)
	case CmdHBIntervalAck: // 0xA7
		err = a.parseHBIntervalAck(data, std)
	case CmdTimeRequest: // 0xA8
		err = a.parseTimeRequest(data, std)
	case CmdAllPortsReply: // 0xB1
		err = a.parseAllPortsReply(data, std)
	case CmdOnePortReply: // 0xB3
		err = a.parseOnePortReply(data, std)
	case CmdLocalCharge: // 0xB4
		err = a.parseLocalCharge(data, std)
	case CmdSwipeCard: // 0xB6
		err = a.parseSwipeCard(data, std)
	case CmdRemoteStartAck: // 0xB8
		err = a.parseRemoteStartAck(data, std)
	case CmdRemoteStopAck: // 0xBA
		err = a.parseRemoteStopAck(data, std)
	case CmdSettlement: // 0xBB
		err = a.parseSettlement(data, std)
	case CmdFaultReport: // 0xC0
		err = a.parseFaultReport(data, std)
	case CmdGearReport: // 0xC2
		err = a.parseGearReport(data, std)
	case CmdParamSetAck: // 0xC4
		err = a.parseParamSetAck(data, std)
	case CmdConfigUpload: // 0xC6
		err = a.parseConfigUpload(data, std)
	case CmdPlatformParam: // 0xC7
		err = a.parsePlatformParam(data, std)
	case CmdQueryCard: // 0xD0
		err = a.parseQueryCard(data, std)
	default:
		std.Extra["raw_cmd"] = fmt.Sprintf("0x%02X", cmd)
		std.Extra["raw_data"] = fmt.Sprintf("%X", data)
	}

	return std, err
}

// ========== 各命令解析函数 ==========

// parseLogin 解析 0xA0 设备登录帧
// DATA: SN(8) + HW_Major(1) + HW_Minor(1) + DeviceID(2->4) + FW_Major(1) + FW_Minor(1) + ModuleNo(4) + SIM(11) + Signal(1)
func (a *WSDAdapter) parseLogin(data []byte, std *model.StandardData) error {
	std.Extra["cmd"] = "login"
	if len(data) < 20 {
		return fmt.Errorf("login data too short: %d", len(data))
	}
	offset := 0

	// SN (8字节)
	std.Extra["sn"] = strings.TrimRight(string(data[offset:offset+8]), "\x00")
	offset += 8

	// 硬件版本 (2字节)
	if offset+2 <= len(data) {
		std.Extra["hw_ver"] = fmt.Sprintf("V%d.%d", data[offset], data[offset+1])
		offset += 2
	}

	// 设备ID (4字节 LE)
	if offset+4 <= len(data) {
		devID := binary.LittleEndian.Uint32(data[offset : offset+4])
		std.DeviceID = fmt.Sprintf("%d", devID)
		std.Extra["raw_device_id"] = devID
		offset += 4
	}

	// 固件版本 (2字节)
	if offset+2 <= len(data) {
		std.FirmwareVer = fmt.Sprintf("V%d.%d", data[offset], data[offset+1])
		offset += 2
	}

	// 模块号 (4字节)
	if offset+4 <= len(data) {
		std.Extra["module_no"] = fmt.Sprintf("%X", data[offset:offset+4])
		offset += 4
	}

	// SIM卡号 (11字节 ASCII)
	if offset+11 <= len(data) {
		std.SIM = strings.TrimRight(string(data[offset:offset+11]), "\x00")
		offset += 11
	}

	// 信号强度 (1字节)
	if offset < len(data) {
		std.SignalStrength = int(data[offset])
	}

	return nil
}

// parseHeartbeat 解析 0xA4 心跳上传帧
// DATA: PortCount(1) + PortStates[N] + FaultFlag(1) + Voltage(2) + Temperature(1) + [各端口充电动态...]
func (a *WSDAdapter) parseHeartbeat(data []byte, std *model.StandardData) error {
	std.Extra["cmd"] = "heartbeat"
	if len(data) < 4 {
		return fmt.Errorf("heartbeat data too short: %d", len(data))
	}
	offset := 0

	// 端口数量 (1字节)
	portCount := int(data[offset])
	std.PortCount = portCount
	offset++

	// 各端口状态 (portCount 字节)
	if offset+portCount <= len(data) {
		std.PortData = make([]model.PortData, portCount)
		for i := 0; i < portCount; i++ {
			std.PortData[i] = model.PortData{
				PortIndex:  i + 1,
				PortStatus: a.parsePortStatus(data[offset+i]),
			}
		}
		offset += portCount
	}

	// 故障标志 (1字节)
	if offset < len(data) {
		faultFlag := data[offset]
		std.Extra["fault_flag"] = int(faultFlag)
		if faultFlag != 0 {
			std.ChargingStatus = "fault"
			std.FaultCode = fmt.Sprintf("FAULT_%d", faultFlag)
		}
		offset++
	}

	// 电压 (2字节, 0.1V)
	if offset+2 <= len(data) {
		std.Voltage = float64(binary.LittleEndian.Uint16(data[offset:offset+2])) * 0.1
		offset += 2
	}

	// 温度 (1字节, 减40)
	if offset < len(data) && data[offset] != 0 {
		std.Temperature = float64(data[offset]) - 40
		offset++
	}

	// 如果有更多数据，尝试解析各端口动态数据
	// TODO: parse per-port dynamic charging data if present
	if offset < len(data) {
		std.Extra["heartbeat_extra"] = fmt.Sprintf("%X", data[offset:])
	}

	return nil
}

// parseRemoteCtrlAck 解析 0xA3 远程控制应答
func (a *WSDAdapter) parseRemoteCtrlAck(data []byte, std *model.StandardData) error {
	std.Extra["cmd"] = "remote_ctrl_ack"
	if len(data) >= 1 {
		std.Extra["result"] = int(data[0])
	}
	return nil
}

// parseHBIntervalAck 解析 0xA7 心跳间隔设置应答
func (a *WSDAdapter) parseHBIntervalAck(data []byte, std *model.StandardData) error {
	std.Extra["cmd"] = "hb_interval_ack"
	if len(data) >= 1 {
		std.Extra["result"] = int(data[0])
	}
	return nil
}

// parseTimeRequest 解析 0xA8 对时请求
func (a *WSDAdapter) parseTimeRequest(data []byte, std *model.StandardData) error {
	std.Extra["cmd"] = "time_request"
	if len(data) >= 6 {
		// BCD 6字节: 年(2)+月(1)+日(1)+时(1)+分(1)+秒(1)
		std.Extra["device_time"] = fmt.Sprintf("20%02d-%02d-%02d %02d:%02d:%02d",
			data[0], data[1], data[2], data[3], data[4], data[5])
	}
	return nil
}

// parseAllPortsReply 解析 0xB1 所有端口状态回复
func (a *WSDAdapter) parseAllPortsReply(data []byte, std *model.StandardData) error {
	std.Extra["cmd"] = "all_ports_reply"
	if len(data) < 1 {
		return fmt.Errorf("all ports reply too short")
	}
	// 每端口固定长度：Status(1)+OrderNo(16)+Power(2)+Gear(1)+ChargeTime(2)+Energy(2)+StartType(1)+... = 约30字节
	// 简化解析：首字节为端口数
	portCount := int(data[0])
	std.PortCount = portCount
	offset := 1
	portRecordLen := 25 // approximate minimum per-port record size

	if portCount > 0 && len(data) >= offset+portCount*portRecordLen {
		std.PortData = make([]model.PortData, portCount)
		for i := 0; i < portCount; i++ {
			po := offset + i*portRecordLen
			if po+portRecordLen > len(data) {
				break
			}
			pd := model.PortData{PortIndex: i + 1}
			pd.PortStatus = a.parsePortStatus(data[po])
			pd.Power = float64(binary.LittleEndian.Uint16(data[po+17 : po+19])) * 0.1
			std.PortData[i] = pd
		}
	}

	if len(data) > 0 {
		std.Extra["raw"] = fmt.Sprintf("%X", data)
	}
	return nil
}

// parseOnePortReply 解析 0xB3 某个端口状态回复
func (a *WSDAdapter) parseOnePortReply(data []byte, std *model.StandardData) error {
	std.Extra["cmd"] = "one_port_reply"
	if len(data) < 3 {
		return fmt.Errorf("one port reply too short")
	}
	offset := 0

	std.PortNum = int(data[offset]) + 1 // 内码0-based转1-based
	offset++

	portStatus := a.parsePortStatus(data[offset])
	offset++

	if offset+4 <= len(data) {
		std.Extra["charge_mode"] = int(data[offset])
		offset++
	}

	std.PortData = []model.PortData{{
		PortIndex:  std.PortNum,
		PortStatus: portStatus,
	}}

	if offset+16 <= len(data) {
		std.OrderNo = strings.TrimRight(string(data[offset:offset+16]), "\x00")
		offset += 16
	}
	if offset+2 <= len(data) {
		std.Power = float64(binary.LittleEndian.Uint16(data[offset:offset+2])) * 0.1
		offset += 2
	}
	if offset+2 <= len(data) {
		std.ChargeTime = int(binary.LittleEndian.Uint16(data[offset : offset+2]))
		offset += 2
	}
	if offset+2 <= len(data) {
		std.ChargeEnergy = float64(binary.LittleEndian.Uint16(data[offset:offset+2])) * 0.01
	}

	return nil
}

// parseLocalCharge 解析 0xB4 本地充电上报
func (a *WSDAdapter) parseLocalCharge(data []byte, std *model.StandardData) error {
	std.Extra["cmd"] = "local_charge"
	if len(data) < 4 {
		return fmt.Errorf("local charge data too short")
	}
	offset := 0

	std.PortNum = int(data[offset]) + 1
	offset++

	std.StartType = int(data[offset]) // 0=投币 1=刷卡 2=即插即充
	offset++

	// 卡号 (4字节, BCD) 或 投币次数
	if offset+4 <= len(data) {
		std.CardID = fmt.Sprintf("%08X", binary.LittleEndian.Uint32(data[offset:offset+4]))
		offset += 4
	}

	// 充电时间 (2字节, 分钟) 或 金额
	if offset+2 <= len(data) {
		std.ChargeTime = int(binary.LittleEndian.Uint16(data[offset:offset+2])) * 60
		offset += 2
	}

	std.ChargingStatus = "charging"
	return nil
}

// parseSwipeCard 解析 0xB6 在线卡刷卡上报
func (a *WSDAdapter) parseSwipeCard(data []byte, std *model.StandardData) error {
	std.Extra["cmd"] = "swipe_card"
	if len(data) < 7 {
		return fmt.Errorf("swipe card data too short")
	}
	offset := 0

	// 卡号 (4字节, BCD大端)
	cardID := binary.BigEndian.Uint32(data[offset : offset+4])
	std.CardID = fmt.Sprintf("%d", cardID)
	offset += 4

	// 端口号
	std.PortNum = int(data[offset]) + 1
	offset++

	// 扣费金额 (2字节, 分)
	if offset+2 <= len(data) {
		std.CardMoney = float64(binary.LittleEndian.Uint16(data[offset:offset+2])) / 100.0
	}

	return nil
}

// parseRemoteStartAck 解析 0xB8 远程启动应答
func (a *WSDAdapter) parseRemoteStartAck(data []byte, std *model.StandardData) error {
	std.Extra["cmd"] = "remote_start_ack"
	if len(data) >= 1 {
		result := data[0]
		std.Extra["result"] = int(result)
		if result == 0 {
			std.ChargingStatus = "charging"
		} else {
			std.FaultCode = fmt.Sprintf("START_FAIL_%d", result)
		}
	}
	if len(data) >= 3 {
		std.PortNum = int(data[1]) + 1
	}
	return nil
}

// parseRemoteStopAck 解析 0xBA 远程停止应答
func (a *WSDAdapter) parseRemoteStopAck(data []byte, std *model.StandardData) error {
	std.Extra["cmd"] = "remote_stop_ack"
	if len(data) >= 1 {
		std.Extra["result"] = int(data[0])
	}
	return nil
}

// parseSettlement 解析 0xBB 结算订单上传 (18字段)
// DATA: Port(1) + OrderNo(16) + StartType(1) + StopReason(1) + ChargeMode(1) +
//
//	StartTime(6,BCD) + EndTime(6,BCD) + ChargeTime(4,秒) + Energy(4,0.001kWh) +
//	Money(4,分) + PeakPower(2,0.1W) + AvgPower(2,0.1W) + CardNo(4) +
//	CoinCount(1) + GearStart(1) + GearEnd(1) + MaxTemp(1) + PortVoltage(2,0.1V)
func (a *WSDAdapter) parseSettlement(data []byte, std *model.StandardData) error {
	std.Extra["cmd"] = "settlement"
	if len(data) < 45 {
		return fmt.Errorf("settlement data too short: %d", len(data))
	}
	offset := 0

	// 端口号 (1字节)
	std.PortNum = int(data[offset]) + 1
	offset++

	// 订单号 (16字节 ASCII)
	std.OrderNo = strings.TrimRight(string(data[offset:offset+16]), "\x00")
	offset += 16

	// 启动方式 (1字节)
	std.StartType = int(data[offset])
	offset++

	// 停止原因 (1字节)
	std.StopReason = int(data[offset])
	std.Extra["stop_reason_msg"] = a.getStopReasonMsg(std.StopReason)
	offset++

	// 充电模式 (1字节)
	std.ChargeMode = int(data[offset])
	offset++

	// 充电起止时间 (6字节 BCD each)
	if offset+6 <= len(data) {
		std.Extra["start_time"] = fmt.Sprintf("20%02d-%02d-%02d %02d:%02d:%02d",
			data[offset], data[offset+1], data[offset+2],
			data[offset+3], data[offset+4], data[offset+5])
		offset += 6
	}
	if offset+6 <= len(data) {
		std.Extra["end_time"] = fmt.Sprintf("20%02d-%02d-%02d %02d:%02d:%02d",
			data[offset], data[offset+1], data[offset+2],
			data[offset+3], data[offset+4], data[offset+5])
		offset += 6
	}

	// 充电总时长 (4字节, 秒)
	if offset+4 <= len(data) {
		std.ChargeTime = int(binary.LittleEndian.Uint32(data[offset : offset+4]))
		offset += 4
	}

	// 充电电量 (4字节, 0.001kWh)
	if offset+4 <= len(data) {
		std.ChargeEnergy = float64(binary.LittleEndian.Uint32(data[offset:offset+4])) * 0.001
		offset += 4
	}

	// 充电金额 (4字节, 分)
	if offset+4 <= len(data) {
		std.ChargeMoney = float64(binary.LittleEndian.Uint32(data[offset:offset+4])) / 100.0
		offset += 4
	}

	// 峰值功率 (2字节, 0.1W)
	if offset+2 <= len(data) {
		std.PeakPower = float64(binary.LittleEndian.Uint16(data[offset:offset+2])) * 0.1
		offset += 2
	}

	// 平均功率 (2字节, 0.1W)
	if offset+2 <= len(data) {
		std.Extra["avg_power"] = float64(binary.LittleEndian.Uint16(data[offset:offset+2])) * 0.1
		offset += 2
	}

	// 卡号 (4字节)
	if offset+4 <= len(data) {
		std.CardID = fmt.Sprintf("%d", binary.LittleEndian.Uint32(data[offset:offset+4]))
		offset += 4
	}

	// 投币次数 (1字节)
	if offset < len(data) {
		std.Extra["coin_count"] = int(data[offset])
		offset++
	}

	// 起始/结束分档 (各1字节)
	if offset < len(data) {
		std.Extra["gear_start"] = int(data[offset])
		offset++
	}
	if offset < len(data) {
		std.Extra["gear_end"] = int(data[offset])
		offset++
	}

	// 最高温度 (1字节, 减40)
	if offset < len(data) && data[offset] != 0 {
		std.Extra["max_temp"] = float64(data[offset]) - 40
		offset++
	}

	// 端口电压 (2字节, 0.1V)
	if offset+2 <= len(data) {
		std.Voltage = float64(binary.LittleEndian.Uint16(data[offset:offset+2])) * 0.1
	}

	std.ChargingStatus = "finished"
	return nil
}

// parseFaultReport 解析 0xC0 故障上报
func (a *WSDAdapter) parseFaultReport(data []byte, std *model.StandardData) error {
	std.Extra["cmd"] = "fault_report"
	if len(data) < 3 {
		return fmt.Errorf("fault report too short")
	}
	offset := 0

	std.PortNum = int(data[offset]) + 1
	offset++

	faultType := data[offset]
	std.FaultCode = fmt.Sprintf("FAULT_%d", faultType)
	std.FaultMessage = a.getFaultMsg(int(faultType))
	offset++

	if offset < len(data) {
		std.Temperature = float64(data[offset]) - 40
		offset++
	}
	if offset+2 <= len(data) {
		std.Voltage = float64(binary.LittleEndian.Uint16(data[offset:offset+2])) * 0.1
	}

	std.ChargingStatus = "fault"
	return nil
}

// parseGearReport 解析 0xC2 充电分档上报
func (a *WSDAdapter) parseGearReport(data []byte, std *model.StandardData) error {
	std.Extra["cmd"] = "gear_report"
	if len(data) < 3 {
		return fmt.Errorf("gear report too short")
	}
	std.PortNum = int(data[0]) + 1
	std.Extra["gear_level"] = int(data[1])
	if len(data) >= 4 {
		std.Power = float64(binary.LittleEndian.Uint16(data[2:4])) * 0.1
	}
	return nil
}

// parseParamSetAck 解析 0xC4 参数配置应答
func (a *WSDAdapter) parseParamSetAck(data []byte, std *model.StandardData) error {
	std.Extra["cmd"] = "param_set_ack"
	if len(data) >= 1 {
		std.Extra["result"] = int(data[0])
	}
	return nil
}

// parseConfigUpload 解析 0xC6 上传配置表
func (a *WSDAdapter) parseConfigUpload(data []byte, std *model.StandardData) error {
	std.Extra["cmd"] = "config_upload"
	std.Extra["raw_config"] = fmt.Sprintf("%X", data)
	return nil
}

// parsePlatformParam 解析 0xC7 获取平台参数表
func (a *WSDAdapter) parsePlatformParam(data []byte, std *model.StandardData) error {
	std.Extra["cmd"] = "platform_param"
	return nil
}

// parseQueryCard 解析 0xD0 查询在线卡
func (a *WSDAdapter) parseQueryCard(data []byte, std *model.StandardData) error {
	std.Extra["cmd"] = "query_card"
	if len(data) >= 4 {
		cardID := binary.BigEndian.Uint32(data[0:4])
		std.CardID = fmt.Sprintf("%d", cardID)
	}
	return nil
}

// ========== 公共辅助函数 ==========

// parsePortStatus 解析端口状态码 → 字符串
func (a *WSDAdapter) parsePortStatus(status byte) string {
	switch status {
	case PortIdle:
		return "idle"
	case PortCharging:
		return "charging"
	case PortScanned:
		return "scanned"
	case PortScannedCharged:
		return "scanned_charged"
	case PortShortCircuit:
		return "short_circuit"
	case PortFloating:
		return "floating"
	case PortFullStop:
		return "finished"
	case PortChipStuck:
		return "chip_stuck"
	case PortRelayStuck:
		return "relay_stuck"
	case PortRelayNoClose:
		return "relay_no_close"
	case PortOverTemp:
		return "over_temp"
	case PortReserved:
		return "reserved"
	default:
		return fmt.Sprintf("unknown_0x%02X", status)
	}
}

// getStopReasonMsg 获取停止原因描述
func (a *WSDAdapter) getStopReasonMsg(reason int) string {
	messages := map[int]string{
		0x00: "充满自停",
		0x01: "远程停止",
		0x02: "刷卡停止",
		0x03: "投币时间到",
		0x04: "过载保护",
		0x05: "过温保护",
		0x06: "过压保护",
		0x07: "欠压保护",
		0x08: "功率过小",
		0x09: "漏电保护",
		0x0A: "紧急停止",
		0x0B: "未检测到电池",
		0x0C: "继电器故障",
		0x0D: "短路保护",
		0x0E: "插座移除",
		0x0F: "断网强制停止",
		0x10: "系统错误",
		0x11: "余额不足",
	}
	if msg, ok := messages[reason]; ok {
		return msg
	}
	return fmt.Sprintf("未知原因_%d", reason)
}

// getFaultMsg 获取故障描述
func (a *WSDAdapter) getFaultMsg(faultType int) string {
	messages := map[int]string{
		1:  "过载",
		2:  "过温",
		3:  "过压",
		4:  "欠压",
		5:  "漏电",
		6:  "继电器粘连",
		7:  "继电器不闭合",
		8:  "短路",
		9:  "计量芯片故障",
		10: "刷卡芯片故障",
		11: "存储器故障",
		12: "通信故障",
		13: "插座移除",
	}
	if msg, ok := messages[faultType]; ok {
		return msg
	}
	return fmt.Sprintf("未知故障_%d", faultType)
}

// ========== Encode: 平台指令编码 ==========

// Encode 将标准指令编码为微小电协议帧
func (a *WSDAdapter) Encode(cmd *model.StandardCommand) ([]byte, error) {
	sessionID := a.generateSessionID()
	var cmdByte byte
	var payload []byte
	var err error

	switch cmd.CmdType {
	case "query_all_ports":
		cmdByte = CmdQueryAllPorts
		payload = []byte{}
	case "query_one_port":
		cmdByte = CmdQueryOnePort
		payload, err = a.buildQueryOnePort(cmd)
	case "start_charge":
		cmdByte = CmdRemoteStart
		payload, err = a.buildRemoteStart(cmd)
	case "stop_charge":
		cmdByte = CmdRemoteStop
		payload, err = a.buildRemoteStop(cmd)
	case "set_params":
		cmdByte = CmdSetParam
		payload, err = a.buildSetParams(cmd)
	case "read_params":
		cmdByte = CmdReadParam
		payload = []byte{}
	case "set_hb_interval":
		cmdByte = CmdSetHBInterval
		payload, err = a.buildSetHBInterval(cmd)
	case "remote_ctrl":
		cmdByte = CmdRemoteCtrl
		payload, err = a.buildRemoteCtrl(cmd)
	case "card_invalid":
		cmdByte = CmdCardInvalid
		payload = []byte{}
	default:
		return nil, fmt.Errorf("unsupported command type: %s", cmd.CmdType)
	}

	if err != nil {
		return nil, err
	}

	return a.buildFrame(sessionID, cmdByte, payload), nil
}

// buildFrame 构建完整的微小电协议帧
func (a *WSDAdapter) buildFrame(sessionID []byte, cmd byte, data []byte) []byte {
	dataLen := len(data)
	frameLen := CmdLen + SessionIDLen + dataLen + SumLen // LEN = 1 + 6 + N + 1
	totalLen := HeaderOverhead + frameLen                // SOP(1) + LEN value

	frame := make([]byte, totalLen)
	frame[0] = SOP
	frame[1] = byte(frameLen)
	frame[2] = cmd
	copy(frame[3:9], sessionID)
	copy(frame[9:9+dataLen], data)

	// 计算异或校验: LEN ^ CMD ^ SESSION_ID ^ DATA (at position 1+LEN)
	frame[9+dataLen] = a.calcXOR(frame[1 : 9+dataLen]) // 9+dataLen == 1+frameLen

	return frame
}

// generateSessionID 生成 6 字节随机 SessionID
func (a *WSDAdapter) generateSessionID() []byte {
	sid := make([]byte, 6)
	rand.Read(sid)
	return sid
}

// ========== Encode 辅助 ==========

func (a *WSDAdapter) buildQueryOnePort(cmd *model.StandardCommand) ([]byte, error) {
	port := byte(0)
	if v, ok := cmd.Params["port"].(float64); ok {
		port = byte(v - 1) // 转为0-based
	}
	return []byte{port}, nil
}

func (a *WSDAdapter) buildRemoteStart(cmd *model.StandardCommand) ([]byte, error) {
	// DATA: Port(1) + OrderNo(16) + ChargeMode(1) + Amount(4,分/秒) + MaxPower(2,0.1W)
	data := make([]byte, 24)

	// 端口号 (0-based)
	port := byte(0)
	if v, ok := cmd.Params["port"].(float64); ok {
		port = byte(v - 1)
	}
	data[0] = port

	// 订单号 (16字节)
	if orderNo, ok := cmd.Params["order_no"].(string); ok {
		copy(data[1:17], []byte(orderNo))
	}

	// 充电模式: 0=按时 1=按量 2=按金额
	mode := byte(0)
	if v, ok := cmd.Params["charge_mode"].(float64); ok {
		mode = byte(v)
	}
	data[17] = mode

	// 金额/时间/电量 (4字节)
	if v, ok := cmd.Params["amount"].(float64); ok {
		binary.LittleEndian.PutUint32(data[18:22], uint32(v))
	}

	// 最大功率 (2字节, 0.1W)
	if v, ok := cmd.Params["max_power"].(float64); ok {
		binary.LittleEndian.PutUint16(data[22:24], uint16(v/0.1))
	}

	return data, nil
}

func (a *WSDAdapter) buildRemoteStop(cmd *model.StandardCommand) ([]byte, error) {
	// DATA: Port(1) + OrderNo(16)
	data := make([]byte, 17)
	port := byte(0)
	if v, ok := cmd.Params["port"].(float64); ok {
		port = byte(v - 1)
	}
	data[0] = port
	if orderNo, ok := cmd.Params["order_no"].(string); ok {
		copy(data[1:17], []byte(orderNo))
	}
	return data, nil
}

func (a *WSDAdapter) buildSetParams(cmd *model.StandardCommand) ([]byte, error) {
	// 28项参数配置表，按协议固定顺序编码
	data := make([]byte, 64) // 28项大约需要64字节
	// 通过 cmd.Params 按协议格式编码
	// 格式: index(1) + value(N)
	// 简化：直接透传 params 的 raw 数据
	if raw, ok := cmd.Params["raw"].(string); ok {
		return []byte(raw), nil
	}
	// 否则构建空的默认帧
	return data, nil
}

func (a *WSDAdapter) buildSetHBInterval(cmd *model.StandardCommand) ([]byte, error) {
	// DATA: Interval(2,秒,LE) + IncludePortStatus(1)
	data := make([]byte, 3)
	interval := uint16(30) // 默认30秒
	if v, ok := cmd.Params["interval"].(float64); ok {
		interval = uint16(v)
	}
	binary.LittleEndian.PutUint16(data[0:2], interval)
	includePort := byte(1) // 默认包含端口状态
	if v, ok := cmd.Params["include_port_status"].(float64); ok {
		includePort = byte(v)
	}
	data[2] = includePort
	return data, nil
}

func (a *WSDAdapter) buildRemoteCtrl(cmd *model.StandardCommand) ([]byte, error) {
	// DATA: CtrlType(1): 0=复位 1=版本更新
	ctrlType := byte(0)
	if v, ok := cmd.Params["ctrl_type"].(float64); ok {
		ctrlType = byte(v)
	}
	return []byte{ctrlType}, nil
}

// ========== AutoReply: 自动回复 ==========

// AutoReply 构建自动回复帧 (登录应答/心跳应答/对时应答)
func (a *WSDAdapter) AutoReply(raw []byte, std *model.StandardData) []byte {
	cmdStr, _ := std.Extra["cmd"].(string)

	// 提取原始请求帧的 SESSION_ID
	var rawSID []byte
	if len(raw) >= 9 {
		rawSID = raw[3:9]
	}

	switch cmdStr {
	case "login":
		// 0xA1 登录应答: 含对时功能 (BCD 6字节: 年月日时分秒)
		// 设备登录时 SESSION_ID 通常为 "000000"，需替换为有效ID
		sessionID := rawSID
		if len(sessionID) == 6 {
			allZero := true
			for _, b := range sessionID {
				if b != 0 {
					allZero = false
					break
				}
			}
			if allZero {
				sessionID = a.generateSessionID()
			}
		}
		return a.buildTimeResponse(CmdLoginAck, sessionID)

	case "heartbeat":
		// 0xA5 心跳应答: 回显登录时分配的 SESSION_ID（不是设备心跳帧中的全零SID）
		// 设备需要收到与登录ACK一致的SESSION_ID才会确认心跳有效
		sessionID := rawSID
		// 优先使用登录时分配的 SESSION_ID
		if stored, ok := std.Extra["wsd_session_id"]; ok {
			if sidBytes, ok := stored.([]byte); ok && len(sidBytes) == 6 {
				sessionID = sidBytes
			}
		}
		if len(sessionID) < 6 {
			sessionID = a.generateSessionID()
		}
		return a.buildFrame(sessionID, CmdHeartbeatAck, []byte{})

	case "time_request":
		// 0xA9 对时应答: BCD 6字节时间，回显设备 SESSION_ID
		sessionID := rawSID
		if len(sessionID) < 6 {
			sessionID = a.generateSessionID()
		}
		return a.buildTimeResponse(CmdTimeAck, sessionID)
	}

	return nil
}

// buildTimeResponse 构建带 BCD 时间回复帧
func (a *WSDAdapter) buildTimeResponse(cmd byte, sessionID []byte) []byte {
	now := time.Now()
	data := make([]byte, 6)
	data[0] = byte(now.Year() % 100) // 年(后2位, BCD式)
	data[1] = byte(now.Month())      // 月
	data[2] = byte(now.Day())        // 日
	data[3] = byte(now.Hour())       // 时
	data[4] = byte(now.Minute())     // 分
	data[5] = byte(now.Second())     // 秒
	return a.buildFrame(sessionID, cmd, data)
}

// ========== DecodeResponse: 解析设备对指令的响应 ==========

// DecodeResponse 解析设备对平台指令的响应帧
func (a *WSDAdapter) DecodeResponse(raw []byte) (*model.StandardCommandResponse, error) {
	if !a.Validate(raw) {
		return nil, fmt.Errorf("invalid WSD response frame")
	}

	_, data, cmd := a.getDataPayload(raw)

	resp := &model.StandardCommandResponse{
		Timestamp: time.Now(),
		Success:   false,
	}

	switch cmd {
	case CmdRemoteStartAck, CmdRemoteStopAck, CmdRemoteCtrlAck, CmdParamSetAck, CmdHBIntervalAck:
		if len(data) >= 1 {
			result := data[0]
			resp.Success = result == 0x00
			if !resp.Success {
				resp.Error = fmt.Sprintf("error code: 0x%02X", result)
			}
		}

	case CmdAllPortsReply:
		resp.Success = true
		resp.Result = map[string]interface{}{
			"port_count": int(data[0]),
			"raw":        fmt.Sprintf("%X", data),
		}

	case CmdOnePortReply:
		resp.Success = true
		result := map[string]interface{}{"port": int(data[0]) + 1}
		if len(data) >= 18 {
			result["order_no"] = strings.TrimRight(string(data[1:17]), "\x00")
		}
		resp.Result = result

	case CmdConfigUpload:
		resp.Success = true
		resp.Result = map[string]interface{}{"config": fmt.Sprintf("%X", data)}

	case CmdSettlement:
		resp.Success = true

	default:
		resp.Success = true
		resp.Result = map[string]interface{}{
			"cmd": fmt.Sprintf("0x%02X", cmd),
			"raw": fmt.Sprintf("%X", data),
		}
	}

	return resp, nil
}
