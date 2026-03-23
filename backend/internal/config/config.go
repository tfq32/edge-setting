package config

import (
	"strings"

	"github.com/spf13/viper"
)

// Config 全局配置
type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Data   DataConfig   `mapstructure:"data"`
	Log    LogConfig    `mapstructure:"log"`
	VSOA   VSOAConfig   `mapstructure:"vsoa"`
}

type ServerConfig struct {
	Port     int    `mapstructure:"port"`
	HTTPS    bool   `mapstructure:"https"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
}

type DataConfig struct {
	Dir           string `mapstructure:"dir"`
	MetricsRetain int    `mapstructure:"metrics_retain_days"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
}

type VSOAConfig struct {
	MSAddress string `mapstructure:"ms_address"`
}

var Global Config

// Load 从文件或环境变量加载配置
func Load(path string) error {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 默认值
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.https", false)
	v.SetDefault("data.dir", "/var/lib/edge-setting")
	v.SetDefault("data.metrics_retain_days", 7)
	v.SetDefault("log.level", "info")
	v.SetDefault("vsoa.ms_address", "vsoa://localhost:3000")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	return v.Unmarshal(&Global)
}
