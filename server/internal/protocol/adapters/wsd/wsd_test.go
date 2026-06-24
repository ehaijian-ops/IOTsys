package wsd

import (
	"encoding/binary"
	"testing"
	"time"

	"iot-platform/internal/protocol/engine"
	"iot-platform/internal/protocol/model"
)

// TestValidate 验证帧校验逻辑
func TestValidate(t *testing.T) {
	adapter := &WSDAdapter{}

	// 构建一个标准 0xA5 心跳应答帧（无DATA）
	frame := buildTestFrame(adapter, []byte{1, 2, 3, 4, 5, 6}, CmdHeartbeatAck, nil)

	if !adapter.Validate(frame) {
		t.Error("Validate() should return true for valid heartbeat ack frame")
	}

	// 测试帧头错误
	badFrame := make([]byte, len(frame))
	copy(badFrame, frame)
	badFrame[0] = 0x44 // 改为DNY的帧头
	if adapter.Validate(badFrame) {
		t.Error("Validate() should return false for non-0xEE SOP")
	}

	// 测试帧过短
	if adapter.Validate(frame[:5]) {
		t.Error("Validate() should return false for short frame")
	}

	// 测试校验错误
	badChecksum := make([]byte, len(frame))
	copy(badChecksum, frame)
	badChecksum[len(badChecksum)-1] ^= 0xFF // 破坏校验
	if adapter.Validate(badChecksum) {
		t.Error("Validate() should return false for bad checksum")
	}
}

// TestDetectProtocolConflict 验证 SOP 检测与 AP3000 不冲突
func TestDetectProtocolConflict(t *testing.T) {
	wsdAdapter := &WSDAdapter{}

	// 构建微小电帧 (SOP=0xEE)
	wsdFrame := buildTestFrame(wsdAdapter, []byte{0, 0, 0, 0, 0, 0}, CmdLogin, make([]byte, 30))
	if t.Failed() {
		return
	}

	// 确认 WSD 帧 Validate 通过
	if !wsdAdapter.Validate(wsdFrame) {
		t.Error("WSD adapter should validate 0xEE frame")
	}

	// 构建 AP3000 帧 (SOP=0x44,0x4E,0x59)
	dnyFrame := buildDNYFrame()

	// 确保 WSD adapter 不误判 DNY 帧
	if wsdAdapter.Validate(dnyFrame) {
		t.Error("WSD adapter should NOT validate DNY (0x44) frame")
	}
}

// TestDecodeLogin 测试解析登录帧
func TestDecodeLogin(t *testing.T) {
	adapter := &WSDAdapter{}

	// 构造登录帧 DATA: SN(8) + HW(2) + DevID(4) + FW(2) + Module(4) + SIM(11) + Signal(1) = 32
	data := make([]byte, 32)
	copy(data[0:8], []byte("WSD12345"))   // SN
	data[8] = 1                            // HW Major
	data[9] = 0                            // HW Minor
	binary.LittleEndian.PutUint32(data[10:14], 1001) // DeviceID
	data[14] = 1                           // FW Major
	data[15] = 0                           // FW Minor
	binary.LittleEndian.PutUint32(data[16:20], 0x01020304) // Module No
	copy(data[20:31], []byte("13800138000")) // SIM
	data[31] = 25                          // Signal

	sessionID := []byte{0, 0, 0, 0, 0, 0}
	frame := buildTestFrame(adapter, sessionID, CmdLogin, data)

	result, err := adapter.Decode(frame)
	if err != nil {
		t.Fatalf("Decode login failed: %v", err)
	}

	if result.DeviceID != "1001" {
		t.Errorf("Expected DeviceID '1001', got '%s'", result.DeviceID)
	}
	if result.SIM != "13800138000" {
		t.Errorf("Expected SIM '13800138000', got '%s'", result.SIM)
	}
	if result.SignalStrength != 25 {
		t.Errorf("Expected SignalStrength 25, got %d", result.SignalStrength)
	}
	if result.FirmwareVer != "V1.0" {
		t.Errorf("Expected FirmwareVer 'V1.0', got '%s'", result.FirmwareVer)
	}
	if sn, _ := result.Extra["sn"].(string); sn != "WSD12345" {
		t.Errorf("Expected SN 'WSD12345', got '%s'", sn)
	}
}

// TestDecodeHeartbeat 测试解析心跳帧
func TestDecodeHeartbeat(t *testing.T) {
	adapter := &WSDAdapter{}

	// 心跳 DATA: PortCount(1) + 12xPortState(12) + FaultFlag(1) + Voltage(2) + Temperature(1) = 17
	portCount := 12
	data := make([]byte, 1+portCount+1+2+1)
	data[0] = byte(portCount)
	// 端口0-3空闲，4-5充电，6空闲，7空闲...
	data[1] = 0x00
	data[2] = 0x00
	data[3] = 0x00
	data[4] = 0x00
	data[5] = 0x01 // charging
	data[6] = 0x01 // charging
	for i := 7; i < 13; i++ {
		data[i] = 0x00
	}
	data[13] = 0x00 // fault flag
	binary.LittleEndian.PutUint16(data[14:16], 2205) // 220.5V
	data[16] = 75 // 35°C (75-40=35)

	frame := buildTestFrame(adapter, []byte{1, 2, 3, 4, 5, 6}, CmdHeartbeat, data)

	result, err := adapter.Decode(frame)
	if err != nil {
		t.Fatalf("Decode heartbeat failed: %v", err)
	}

	if result.PortCount != 12 {
		t.Errorf("Expected PortCount 12, got %d", result.PortCount)
	}
	if len(result.PortData) != 12 {
		t.Errorf("Expected 12 PortData entries, got %d", len(result.PortData))
	}
	if result.Voltage != 220.5 {
		t.Errorf("Expected Voltage 220.5, got %.1f", result.Voltage)
	}
	if result.Temperature != 35.0 {
		t.Errorf("Expected Temperature 35.0, got %.1f", result.Temperature)
	}
	if result.PortData[4].PortStatus != "charging" {
		t.Errorf("Expected port 5 status 'charging', got '%s'", result.PortData[4].PortStatus)
	}
	if result.PortData[0].PortStatus != "idle" {
		t.Errorf("Expected port 1 status 'idle', got '%s'", result.PortData[0].PortStatus)
	}
}

// TestDecodeSettlement 测试解析结算订单
func TestDecodeSettlement(t *testing.T) {
	adapter := &WSDAdapter{}

	// 结算 DATA approx 56 bytes
	data := make([]byte, 56)
	data[0] = 2 // port 2 (0-based → port 3)
	copy(data[1:17], []byte("OD20260619000123"))
	data[17] = 1 // start_type: remote
	data[18] = 0 // stop_reason: 充满自停
	data[19] = 0 // charge_mode: 按时
	// start time: 2026-06-19 12:30:45 (BCD)
	data[20] = 26 // year
	data[21] = 6  // month
	data[22] = 19 // day
	data[23] = 12 // hour
	data[24] = 30 // min
	data[25] = 45 // sec
	// end time: 2026-06-19 14:30:45
	data[26] = 26
	data[27] = 6
	data[28] = 19
	data[29] = 14
	data[30] = 30
	data[31] = 45
	binary.LittleEndian.PutUint32(data[32:36], 7200)    // 7200秒
	binary.LittleEndian.PutUint32(data[36:40], 1500)    // 1.5kWh (1500 * 0.001)
	binary.LittleEndian.PutUint32(data[40:44], 150)     // 1.50元 (150分)
	binary.LittleEndian.PutUint16(data[44:46], 6500)     // 650W peak
	binary.LittleEndian.PutUint16(data[46:48], 4500)     // 450W avg
	binary.LittleEndian.PutUint32(data[48:52], 12345678) // card no
	data[52] = 2 // coin count
	data[53] = 1 // gear start
	data[54] = 2 // gear end
	data[55] = 82 // max temp 42°C (82-40)

	frame := buildTestFrame(adapter, []byte{1, 2, 3, 4, 5, 6}, CmdSettlement, data)

	result, err := adapter.Decode(frame)
	if err != nil {
		t.Fatalf("Decode settlement failed: %v", err)
	}

	if result.PortNum != 3 {
		t.Errorf("Expected PortNum 3, got %d", result.PortNum)
	}
	if result.OrderNo != "OD20260619000123" {
		t.Errorf("Expected OrderNo 'OD20260619000123', got '%s'", result.OrderNo)
	}
	if result.ChargeTime != 7200 {
		t.Errorf("Expected ChargeTime 7200, got %d", result.ChargeTime)
	}
	if result.ChargeEnergy < 1.49 || result.ChargeEnergy > 1.51 {
		t.Errorf("Expected ChargeEnergy ~1.5, got %.3f", result.ChargeEnergy)
	}
	if result.ChargeMoney < 1.49 || result.ChargeMoney > 1.51 {
		t.Errorf("Expected ChargeMoney ~1.50, got %.3f", result.ChargeMoney)
	}
	if result.PeakPower < 649 || result.PeakPower > 651 {
		t.Errorf("Expected PeakPower ~650, got %.1f", result.PeakPower)
	}
	if result.Voltage != 0.0 {
		t.Errorf("Expected Voltage 0.0 (not set in this test), got %.1f", result.Voltage)
	}
}

// TestAutoReply 测试自动回复
func TestAutoReply(t *testing.T) {
	adapter := &WSDAdapter{}

	// 构造登录请求帧
	loginData := make([]byte, 32)
	copy(loginData[0:8], []byte("WSD12345"))
	// ... fill rest
	loginFrame := buildTestFrame(adapter, []byte{0, 0, 0, 0, 0, 0}, CmdLogin, loginData)

	// 先解析
	std, err := adapter.Decode(loginFrame)
	if err != nil {
		t.Fatalf("Decode login for auto-reply failed: %v", err)
	}

	// 测试自动回复
	reply := adapter.AutoReply(loginFrame, std)
	if len(reply) == 0 {
		t.Error("AutoReply should return non-empty reply for login")
	}
	if reply[0] != SOP {
		t.Error("AutoReply frame should start with SOP 0xEE")
	}
	if reply[2] != CmdLoginAck {
		t.Errorf("AutoReply cmd should be 0xA1, got 0x%02X", reply[2])
	}
	if !adapter.Validate(reply) {
		t.Error("AutoReply frame should be valid")
	}

	// 验证回复中包含时间信息 (6字节BCD)
	// SUM at frame[1+LEN], DATA = frame[9 : 1+LEN]
	frameLen := int(reply[1])
	dataEnd := 1 + frameLen // position of SUM byte
	dataLen := dataEnd - 9  // DATA length
	if dataLen != 6 {
		t.Errorf("Login ack data should be 6 bytes (BCD time), got %d (frameLen=%d)", dataLen, frameLen)
	}
}

// TestEncodeRemoteStart 测试编码远程启动指令
func TestEncodeRemoteStart(t *testing.T) {
	adapter := &WSDAdapter{}

	cmd := &model.StandardCommand{
		DeviceID:  "1001",
		Protocol:  ProtocolName,
		CmdType:   "start_charge",
		CreatedAt: time.Now(),
		Params: map[string]interface{}{
			"port":        float64(3),
			"order_no":    "OD20260619000123",
			"charge_mode": float64(0), // 按时
			"amount":      float64(120), // 120分钟
			"max_power":   float64(6500), // 650W
		},
	}

	frame, err := adapter.Encode(cmd)
	if err != nil {
		t.Fatalf("Encode remote start failed: %v", err)
	}

	if !adapter.Validate(frame) {
		t.Error("Encoded frame should be valid")
	}
	if frame[2] != CmdRemoteStart {
		t.Errorf("Expected cmd 0xB7, got 0x%02X", frame[2])
	}
	// Port 3 → index 2 (0-based)
	if frame[9] != 2 {
		t.Errorf("Expected port index 2, got %d", frame[9])
	}
}

// TestEncodeRemoteStop 测试编码远程停止指令
func TestEncodeRemoteStop(t *testing.T) {
	adapter := &WSDAdapter{}

	cmd := &model.StandardCommand{
		DeviceID:  "1001",
		Protocol:  ProtocolName,
		CmdType:   "stop_charge",
		CreatedAt: time.Now(),
		Params: map[string]interface{}{
			"port":     float64(3),
			"order_no": "OD20260619000123",
		},
	}

	frame, err := adapter.Encode(cmd)
	if err != nil {
		t.Fatalf("Encode remote stop failed: %v", err)
	}

	if frame[2] != CmdRemoteStop {
		t.Errorf("Expected cmd 0xB9, got 0x%02X", frame[2])
	}
}

// TestDecodeResponse 测试解析响应
func TestDecodeResponse(t *testing.T) {
	adapter := &WSDAdapter{}

	// 测试远程启动应答 (成功)
	data := []byte{0x00, 0x02} // result=success, port=2
	frame := buildTestFrame(adapter, []byte{1, 2, 3, 4, 5, 6}, CmdRemoteStartAck, data)

	resp, err := adapter.DecodeResponse(frame)
	if err != nil {
		t.Fatalf("DecodeResponse failed: %v", err)
	}
	if !resp.Success {
		t.Error("Remote start ack with result=0 should be success")
	}

	// 测试远程启动应答 (失败)
	failData := []byte{0x01, 0x02} // result=fail
	failFrame := buildTestFrame(adapter, []byte{1, 2, 3, 4, 5, 6}, CmdRemoteStartAck, failData)

	failResp, err := adapter.DecodeResponse(failFrame)
	if err != nil {
		t.Fatalf("DecodeResponse failed: %v", err)
	}
	if failResp.Success {
		t.Error("Remote start ack with result=1 should not be success")
	}
}

// TestEngineRegister 验证适配器正确注册到引擎
func TestEngineRegister(t *testing.T) {
	adapter, err := engine.GetAdapter(ProtocolName)
	if err != nil {
		t.Fatalf("WSD adapter not registered in engine: %v", err)
	}
	if adapter.Name() != ProtocolName {
		t.Errorf("Expected name %s, got %s", ProtocolName, adapter.Name())
	}
	if adapter.DeviceType() != DeviceType {
		t.Errorf("Expected device_type %s, got %s", DeviceType, adapter.DeviceType())
	}
}

// ========== 辅助函数 ==========

// buildTestFrame 构建测试用协议帧
func buildTestFrame(adapter *WSDAdapter, sessionID []byte, cmd byte, data []byte) []byte {
	return adapter.buildFrame(sessionID, cmd, data)
}

// buildDNYFrame 构建AP3000协议帧(DNY头)用于冲突测试
func buildDNYFrame() []byte {
	// AP3000 帧头 DNY (0x44 0x4E 0x59)
	frame := []byte{
		0x44, 0x4E, 0x59, // DNY header
		0x0B, 0x00, // len = 11 (devID+msgID+cmd+checksum)
		0x01, 0x00, 0x00, 0x00, // deviceID
		0x01, 0x00, // msgID
		0x01, // cmd
		0x00, 0x00, // dummy checksum
	}
	// 计算正确的 AP3000 checksum
	var sum uint32
	for _, b := range frame[:len(frame)-2] {
		sum += uint32(b)
	}
	binary.LittleEndian.PutUint16(frame[len(frame)-2:], uint16(sum&0xFFFF))
	return frame
}
