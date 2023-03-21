package xmongo

import (
	"time"

	"github.com/5idu/pilot/pkg/conf"
	"github.com/5idu/pilot/pkg/constant"
	"github.com/5idu/pilot/pkg/singleton"
	"github.com/5idu/pilot/pkg/xlog"

	"github.com/spf13/cast"
)

// Config ...
type (
	Config struct {
		Name string
		// DSN地址
		DSN string `json:"dsn"`
		// 创建连接的超时时间
		SocketTimeout time.Duration `json:"socketTimeout"`
		// 连接池大小(最大连接数)
		PoolLimit int `json:"poolLimit"`

		EnableTrace  bool `json:"enableTrace"`
		EnableMetric bool `json:"enableMetric"`

		logger *xlog.Logger
	}
)

// StdConfig .
func StdConfig(name string) *Config {
	return RawConfig(constant.ConfigKey("mongo." + name))
}

// RawConfig .
func RawConfig(key string) *Config {
	var config = DefaultConfig()
	config.Name = key
	if err := conf.UnmarshalKey(key, config); err != nil {
		panic(err)
	}
	return config
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		SocketTimeout: cast.ToDuration("5s"),
		PoolLimit:     5,
		EnableTrace:   true,
		EnableMetric:  true,
		logger:        xlog.With(xlog.String("mod", "store.mongo")),
	}
}

func (config *Config) Build() *Client {
	return newClient(config)
}

func (config *Config) Singleton() *Client {
	if val, ok := singleton.Load(constant.ModuleStoreMongoDB, config.Name); ok {
		return val.(*Client)
	}

	val := config.Build()
	singleton.Store(constant.ModuleStoreMongoDB, config.Name, val)

	return val
}
