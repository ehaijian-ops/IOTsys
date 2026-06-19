package engine

import (
	"fmt"
	"sync"

	"iot-platform/internal/protocol/model"
	"iot-platform/pkg/logger"

	"go.uber.org/zap"
)

// ProtocolAdapter 协议适配器接口
type ProtocolAdapter interface {
	// Name 协议标识 (如 AP3000_v2)
	Name() string
	// Version 协议版本
	Version() string
	// DeviceType 设备类型 (ebike_charger / ev_charger)
	DeviceType() string
	// Decode 解析设备上报的原始数据 → 标准数据模型
	Decode(raw []byte) (*model.StandardData, error)
	// Encode 将标准指令编码为设备协议帧
	Encode(cmd *model.StandardCommand) ([]byte, error)
	// DecodeResponse 解析设备对指令的响应
	DecodeResponse(raw []byte) (*model.StandardCommandResponse, error)
	// Validate 校验报文是否为该协议
	Validate(raw []byte) bool
}

// AdapterRegistry 适配器注册中心
type AdapterRegistry struct {
	mu       sync.RWMutex
	adapters map[string]ProtocolAdapter // key: 协议名
}

var registry = &AdapterRegistry{
	adapters: make(map[string]ProtocolAdapter),
}

// Register 注册协议适配器
func Register(adapter ProtocolAdapter) {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	name := adapter.Name()
	if _, exists := registry.adapters[name]; exists {
		logger.Warn("Protocol adapter already registered, overwriting",
			zap.String("name", name),
		)
	}
	registry.adapters[name] = adapter
	logger.Info("Protocol adapter registered",
		zap.String("name", name),
		zap.String("version", adapter.Version()),
		zap.String("device_type", adapter.DeviceType()),
	)
}

// GetAdapter 获取协议适配器
func GetAdapter(name string) (ProtocolAdapter, error) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	adapter, exists := registry.adapters[name]
	if !exists {
		return nil, fmt.Errorf("protocol adapter not found: %s", name)
	}
	return adapter, nil
}

// ListAdapters 列出所有已注册适配器
func ListAdapters() []ProtocolAdapter {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	adapters := make([]ProtocolAdapter, 0, len(registry.adapters))
	for _, a := range registry.adapters {
		adapters = append(adapters, a)
	}
	return adapters
}

// DetectProtocol 自动检测报文协议
func DetectProtocol(raw []byte) (ProtocolAdapter, error) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	for _, adapter := range registry.adapters {
		if adapter.Validate(raw) {
			return adapter, nil
		}
	}
	return nil, fmt.Errorf("unknown protocol for data: %x", raw[:min(len(raw), 20)])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Decode 使用指定协议解析数据
func Decode(protocolName string, raw []byte) (*model.StandardData, error) {
	adapter, err := GetAdapter(protocolName)
	if err != nil {
		return nil, err
	}
	return adapter.Decode(raw)
}

// Encode 使用指定协议编码指令
func Encode(cmd *model.StandardCommand) ([]byte, error) {
	adapter, err := GetAdapter(cmd.Protocol)
	if err != nil {
		return nil, err
	}
	return adapter.Encode(cmd)
}
