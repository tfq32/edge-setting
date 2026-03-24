package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config 全局配置（与 backend 保持一致）
type Config struct {
	Server ServerConfig `yaml:"server"`
	Data   DataConfig   `yaml:"data"`
	Log    LogConfig    `yaml:"log"`
	VSOA   VSOAConfig   `yaml:"vsoa"`
}

type ServerConfig struct {
	Port     int    `yaml:"port"`
	HTTPS    bool   `yaml:"https"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

type DataConfig struct {
	Dir           string `yaml:"dir"`
	MetricsRetain int    `yaml:"metrics_retain_days"`
}

type LogConfig struct {
	Level string `yaml:"level"`
}

type VSOAConfig struct {
	MSAddress string `yaml:"ms_address"`
}

var Global Config

// Load 从 YAML 文件加载配置
func Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// 默认值
	Global = Config{
		Server: ServerConfig{Port: 8080},
		Data:   DataConfig{Dir: "/var/lib/edge-setting", MetricsRetain: 7},
		Log:    LogConfig{Level: "info"},
		VSOA:   VSOAConfig{MSAddress: "vsoa://localhost:3000"},
	}

	return yaml.Unmarshal(data, &Global)
}
