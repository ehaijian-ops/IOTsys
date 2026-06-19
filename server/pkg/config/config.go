package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config 全局配置
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	MySQL     MySQLConfig     `mapstructure:"mysql"`
	MongoDB   MongoDBConfig   `mapstructure:"mongodb"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Kafka     KafkaConfig     `mapstructure:"kafka"`
	TCP       TCPConfig       `mapstructure:"tcp"`
	MQTT      MQTTConfig      `mapstructure:"mqtt"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	WebSocket WebSocketConfig `mapstructure:"websocket"`
	Log       LogConfig       `mapstructure:"log"`
	Alert     AlertConfig     `mapstructure:"alert"`
	Services  ServicesConfig  `mapstructure:"services"`
}

type ServerConfig struct {
	Name string `mapstructure:"name"`
	Env  string `mapstructure:"env"`
	Port int    `mapstructure:"port"`
}

type MySQLConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Database        string        `mapstructure:"database"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	LogLevel        string        `mapstructure:"log_level"`
}

func (m MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		m.User, m.Password, m.Host, m.Port, m.Database)
}

type MongoDBConfig struct {
	URI         string        `mapstructure:"uri"`
	Database    string        `mapstructure:"database"`
	MaxPoolSize uint64        `mapstructure:"max_pool_size"`
	MinPoolSize uint64        `mapstructure:"min_pool_size"`
	Timeout     time.Duration `mapstructure:"timeout"`
}

type RedisConfig struct {
	Addr         string `mapstructure:"addr"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
}

type KafkaConfig struct {
	Brokers       []string         `mapstructure:"brokers"`
	ConsumerGroup string           `mapstructure:"consumer_group"`
	Topics        KafkaTopicsConfig `mapstructure:"topics"`
}

type KafkaTopicsConfig struct {
	DeviceDataReport    string `mapstructure:"device_data_report"`
	DeviceEventOnline   string `mapstructure:"device_event_online"`
	DeviceEventOffline  string `mapstructure:"device_event_offline"`
	DeviceEventFault    string `mapstructure:"device_event_fault"`
	DeviceCommandSend   string `mapstructure:"device_command_send"`
	DeviceCommandResult string `mapstructure:"device_command_result"`
	AlertTriggered      string `mapstructure:"alert_triggered"`
	AlertResolved       string `mapstructure:"alert_resolved"`
	DeviceOTAUpgrade    string `mapstructure:"device_ota_upgrade"`
	DeviceOTAProgress   string `mapstructure:"device_ota_progress"`
	DeviceOTAResult     string `mapstructure:"device_ota_result"`
}

type TCPConfig struct {
	Enabled           bool          `mapstructure:"enabled"`
	Port              int           `mapstructure:"port"`
	MaxConnections    int           `mapstructure:"max_connections"`
	ReadTimeout       time.Duration `mapstructure:"read_timeout"`
	WriteTimeout      time.Duration `mapstructure:"write_timeout"`
	HeartbeatInterval time.Duration `mapstructure:"heartbeat_interval"`
}

type MQTTConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Broker   string `mapstructure:"broker"`
	ClientID string `mapstructure:"client_id"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type JWTConfig struct {
	Secret          string        `mapstructure:"secret"`
	AccessTokenTTL  time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl"`
}

type WebSocketConfig struct {
	HeartbeatInterval time.Duration `mapstructure:"heartbeat_interval"`
	WriteTimeout      time.Duration `mapstructure:"write_timeout"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
	FilePath string `mapstructure:"file_path"`
}

type AlertConfig struct {
	DeviceOfflineTimeout time.Duration `mapstructure:"device_offline_timeout"`
	CheckInterval        time.Duration `mapstructure:"check_interval"`
}

type ServicesConfig struct {
	DeviceService   int `mapstructure:"device_service"`
	CommandService  int `mapstructure:"command_service"`
	AlertService    int `mapstructure:"alert_service"`
	UserService     int `mapstructure:"user_service"`
	ReportService   int `mapstructure:"report_service"`
	ProtocolService int `mapstructure:"protocol_service"`
}

// Load 加载配置
func Load(configPath string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
