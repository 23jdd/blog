package config

type Config struct {
	Port      int   `yaml:"port"`
	ReadLimit int64 `yaml:"read_limit"`
	Rate      int64 `yaml:"rate"`
}
