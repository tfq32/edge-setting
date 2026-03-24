package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config 全局配置
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Log      LogConfig      `yaml:"log"`
	VSOA     VSOAConfig     `yaml:"vsoa"`
	Data     DataConfig     `yaml:"data"`
}

type ServerConfig struct {
	Port     int    `yaml:"port"`
	HTTPS    bool   `yaml:"https"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

type DatabaseConfig struct {
	EdgePath string `yaml:"edge_path"`
}

type DataConfig struct {
	MetricsRetain int `yaml:"metrics_retain_days"`
}

type LogConfig struct {
	Level string `yaml:"level"`
}

type VSOAConfig struct {
	MSAddress string `yaml:"ms_address"`
}

var AppConfig Config

// Load 从 YAML 文件加载配置
func Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// 设置默认值
	AppConfig = Config{
		Server:   ServerConfig{Port: 8080},
		Database: DatabaseConfig{EdgePath: "./edge-setting.db"},
		Data:     DataConfig{MetricsRetain: 7},
		Log:      LogConfig{Level: "info"},
		VSOA:     VSOAConfig{MSAddress: "vsoa://localhost:3000"},
	}

	return yaml.Unmarshal(data, &AppConfig)
}
