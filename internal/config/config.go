package config

import (
	"sync"

	"github.com/spf13/viper"
)

type SMTPConfig struct {
	Host     string `mapstructure:"host" yaml:"host"`
	Port     int    `mapstructure:"port" yaml:"port"`
	Username string `mapstructure:"username" yaml:"username"`
	Password string `mapstructure:"password" yaml:"password"`
	From     string `mapstructure:"from" yaml:"from"`
}

type UploadConfig struct {
	Dir       string `mapstructure:"dir" yaml:"dir"`
	URLPrefix string `mapstructure:"url_prefix" yaml:"url_prefix"`
}

type FeatureConfig struct {
	EnableEmailVerify bool `mapstructure:"enable_email_verify" yaml:"enable_email_verify"`
	EnableMarkdownAPI bool `mapstructure:"enable_markdown_api" yaml:"enable_markdown_api"`
}

type LimitConfig struct {
	Read int64 `mapstructure:"read" yaml:"read"`
	Rate int64 `mapstructure:"rate" yaml:"rate"`
}

type Config struct {
	Port    int           `yaml:"port"`
	Limit   LimitConfig   `mapstructure:"limit" yaml:"limit"`
	SMTP    SMTPConfig    `mapstructure:"smtp" yaml:"smtp"`
	Upload  UploadConfig  `mapstructure:"upload" yaml:"upload"`
	Feature FeatureConfig `mapstructure:"feature" yaml:"feature"`
}

var (
	mu      sync.RWMutex
	current = defaultConfig()
)

func defaultConfig() *Config {
	return &Config{
		Port: 8080,
		Limit: LimitConfig{
			Read: 100,
			Rate: 1,
		},
		SMTP: SMTPConfig{
			Host:     "",
			Port:     587,
			Username: "",
			Password: "",
			From:     "",
		},
		Upload: UploadConfig{
			Dir:       "uploads",
			URLPrefix: "/uploads",
		},
		Feature: FeatureConfig{
			EnableEmailVerify: true,
			EnableMarkdownAPI: true,
		},
	}
}

func setDefaults() {
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("limit.read", 100)
	viper.SetDefault("limit.rate", 1)
	// 兼容旧配置键
	viper.SetDefault("read_limit", 100)
	viper.SetDefault("rate", 1)
	viper.SetDefault("smtp.host", "")
	viper.SetDefault("smtp.port", 587)
	viper.SetDefault("smtp.username", "")
	viper.SetDefault("smtp.password", "")
	viper.SetDefault("smtp.from", "")
	viper.SetDefault("upload.dir", "uploads")
	viper.SetDefault("upload.url_prefix", "/uploads")
	viper.SetDefault("feature.enable_email_verify", true)
	viper.SetDefault("feature.enable_markdown_api", true)
}

func LoadFromViper() (*Config, error) {
	setDefaults()
	cfg := defaultConfig()
	cfg.Port = viper.GetInt("server.port")
	_ = viper.Unmarshal(&cfg)
	// 兼容旧配置: 若未配置 limit.*，则回退到 read_limit/rate
	if !viper.IsSet("limit.read") && viper.IsSet("read_limit") {
		cfg.Limit.Read = viper.GetInt64("read_limit")
	}
	if !viper.IsSet("limit.rate") && viper.IsSet("rate") {
		cfg.Limit.Rate = viper.GetInt64("rate")
	}
	Set(cfg)
	return cfg, nil
}

func Set(cfg *Config) {
	if cfg == nil {
		return
	}
	mu.Lock()
	current = cfg
	mu.Unlock()
}

func Get() *Config {
	mu.RLock()
	defer mu.RUnlock()
	return current
}
