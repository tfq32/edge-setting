package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config 全局配置
type Config struct {
	Server ServerConfig `yaml:"server"`
	Data   DataConfig   `yaml:"data"`
	Log    LogConfig    `yaml:"log"`
	VSOA   VSOAConfig   `yaml:"vsoa"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type DataConfig struct {
	Dir string `yaml:"dir"`
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
		Server: ServerConfig{Port: 10000},
		Data:   DataConfig{Dir: "data"},
		Log:    LogConfig{Level: "info"},
		VSOA:   VSOAConfig{MSAddress: "localhost:3001"},
	}

	return yaml.Unmarshal(data, &Global)
}
