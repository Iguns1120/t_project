package configs

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 代表全域配置結構
type Config struct {
	Server      ServerConfig      `mapstructure:"server"`
	Logger      LoggerConfig      `mapstructure:"logger"`
	Persistence PersistenceConfig `mapstructure:"persistence"`
	Database    DatabaseConfig    `mapstructure:"database"`
	Redis       RedisConfig       `mapstructure:"redis"`
	RocketMQ    RocketMQConfig    `mapstructure:"rocketmq"`
	HealthCheck HealthCheckConfig `mapstructure:"health_check"`
}

// PersistenceConfig 代表持久化配置
type PersistenceConfig struct {
	Type string `mapstructure:"type"` // "memory" 或 "mysql"
}

type ServerConfig struct {
	Port          int    `mapstructure:"port"`
	Mode          string `mapstructure:"mode"`
	SlowThreshold int    `mapstructure:"slow_threshold"`
}

type LoggerConfig struct {
	Level    string `mapstructure:"level"`
	Encoding string `mapstructure:"encoding"`
}

type DatabaseConfig struct {
	DSN                    string `mapstructure:"dsn"`
	MaxOpenConns           int    `mapstructure:"max_open_conns"`
	MaxIdleConns           int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetimeMinutes int    `mapstructure:"conn_max_lifetime_minutes"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type RocketMQConfig struct {
	NameSrvAddr   string `mapstructure:"namesrv_addr"`
	ProducerGroup string `mapstructure:"producer_group"`
	ConsumerGroup string `mapstructure:"consumer_group"`
}

type HealthCheckConfig struct {
	LatencyThreshold int `mapstructure:"latency_threshold"`
}

// LoadConfig 從檔案載入配置
func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("讀取配置檔失敗: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置失敗: %w", err)
	}

	return &config, nil
}