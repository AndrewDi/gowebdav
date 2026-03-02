package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config 表示WebDAV客户端的配置
type Config struct {
	Endpoint      string `json:"endpoint" yaml:"endpoint"`
	Username      string `json:"username" yaml:"username"`
	Password      string `json:"password" yaml:"password"`
	RootPath      string `json:"rootPath" yaml:"rootPath"`
	HTTPProxy     string `json:"httpProxy" yaml:"httpProxy"`
	HTTPSProxy    string `json:"httpsProxy" yaml:"httpsProxy"`
	ProxyUsername string `json:"proxyUsername" yaml:"proxyUsername"`
	ProxyPassword string `json:"proxyPassword" yaml:"proxyPassword"`
}

// LoadConfig 从文件加载配置
func LoadConfig(configPath string) (*Config, error) {
	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 配置文件不存在，返回默认配置
		return &Config{}, nil
	}

	// 读取配置文件
	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 根据文件扩展名选择解析方式
	config := &Config{}
	ext := strings.ToLower(filepath.Ext(configPath))

	switch ext {
	case ".json":
		err = json.Unmarshal(content, config)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(content, config)
	default:
		return nil, fmt.Errorf("unsupported config file format: %s", ext)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// SaveConfig 保存配置到文件
func SaveConfig(config *Config, configPath string) error {
	// 确保目录存在
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 根据文件扩展名选择序列化方式
	var content []byte
	var err error
	ext := strings.ToLower(filepath.Ext(configPath))

	switch ext {
	case ".json":
		content, err = json.MarshalIndent(config, "", "  ")
	case ".yaml", ".yml":
		content, err = yaml.Marshal(config)
	default:
		return fmt.Errorf("unsupported config file format: %s", ext)
	}

	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetDefaultConfigPath 获取默认配置文件路径
func GetDefaultConfigPath() string {
	// 获取用户主目录
	home, err := os.UserHomeDir()
	if err != nil {
		return "./config.yaml"
	}

	// 返回默认配置文件路径
	return filepath.Join(home, ".webdav", "config.yaml")
}
