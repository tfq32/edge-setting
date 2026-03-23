package config

import (
	"strings"

	"github.com/spf13/viper"
)

// Config 全局配置
type Config struct {
	Server  ServerConfig  `mapstructure:"server"`
	Auth    AuthConfig    `mapstructure:"auth"`
	Data    DataConfig    `mapstructure:"data"`
	Log     LogConfig     `mapstructure:"log"`
	VSOA    VSOAConfig    `mapstructure:"vsoa"`
}

type ServerConfig struct {
	Port     int    `mapstructure:"port"`
	HTTPS    bool   `mapstructure:"https"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
}

type TokenEntry struct {
	Token      string `mapstructure:"token"`
	Permission string `mapstructure:"permission"` // readonly | readwrite
}

type AuthConfig struct {
	Tokens []TokenEntry `mapstructure:"tokens"`
}

type DataConfig struct {
	Dir             string `mapstructure:"dir"`
	MetricsRetain   int    `mapstructure:"metrics_retain_days"`
	LogMaxSizeMB    int    `mapstructure:"log_max_size_mb"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type VSOAConfig struct {
	MSAddress      string `mapstructure:"ms_address"`
	ConnectTimeout string `mapstructure:"connect_timeout"`
	RetryMaxInterval string `mapstructure:"retry_max_interval"`
}

var Global Config

// Load 从文件或环境变量加载配置
func Load(path string) error {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	// 环境变量覆盖
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 默认值
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.https", false)
	v.SetDefault("data.dir", "/var/lib/edge-setting")
	v.SetDefault("data.metrics_retain_days", 7)
	v.SetDefault("data.log_max_size_mb", 500)
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
	v.SetDefault("vsoa.ms_address", "vsoa://localhost:3000")
	v.SetDefault("vsoa.connect_timeout", "5s")
	v.SetDefault("vsoa.retry_max_interval", "30s")

	if err := v.ReadInConfig(); err != nil {
		// 配置文件不存在时使用默认值
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	return v.Unmarshal(&Global)
}
