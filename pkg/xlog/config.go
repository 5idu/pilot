package xlog

import (
	"fmt"
	"log"

	"github.com/5idu/pilot/pkg/conf"
	"github.com/5idu/pilot/pkg/constant"

	"go.uber.org/zap"
)

func init() {
	conf.OnLoaded(func(c *conf.Configuration) {
		prefix := constant.GetConfigPrefix()
		log.Print("hook config, init logger")

		key := prefix + ".logger.default"
		log.Printf("reload default logger with configKey: %s\n", key)
		logger = RawConfig(key).Build()
	})
}

const (
	// FormatText format log text
	FormatText = "text"
	// FormatJSON format log json
	FormatJSON = "json"
)

// Config ...
type Config struct {
	Dir       string // 日志输出目录
	Name      string // 日志文件名称
	Level     string
	Compress  bool
	MaxBackup int
	MaxSize   int
	MaxAge    int
	Format    string
	configKey string
}

// Filename ...
func (config *Config) Filename() string {
	return fmt.Sprintf("%s/%s", config.Dir, config.Name)
}

// RawConfig ...
func RawConfig(key string) *Config {
	var config = DefaultConfig()
	config, _ = conf.UnmarshalWithExpect(key, config).(*Config)
	config.configKey = key
	return config
}

// StdConfig Standard logger config
func StdConfig(prefix, name string) *Config {
	return RawConfig(prefix + ".logger." + name)
}

// DefaultConfig for application.
func DefaultConfig() *Config {
	return &Config{
		Name:      "log.json",
		Dir:       "-",
		Level:     "info",
		MaxSize:   500, // 500M
		MaxAge:    1,   // 1 day
		MaxBackup: 10,  // 10 backup
		Format:    FormatJSON,
	}
}

// Build ...
func (config Config) Build() *Logger {
	core := newLoggerCore(&config)
	zapopts := newLoggerOptions()
	return &Logger{
		zlog: zap.New(core, zapopts...),
	}
}
