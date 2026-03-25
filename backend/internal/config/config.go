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
	Port int `mapstructure:"port"`
}

type DataConfig struct {
	Dir string `mapstructure:"dir"`
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
	v.SetDefault("server.port", 10000)
	v.SetDefault("data.dir", "/var/opt/edge-setting")
	v.SetDefault("log.level", "info")
	v.SetDefault("vsoa.ms_address", "localhost:3001")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	return v.Unmarshal(&Global)
}
