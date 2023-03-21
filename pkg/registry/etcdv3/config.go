package etcdv3

import (
	"time"

	"github.com/5idu/pilot/pkg/client/etcdv3"
	"github.com/5idu/pilot/pkg/conf"
	"github.com/5idu/pilot/pkg/constant"
	"github.com/5idu/pilot/pkg/registry"
	"github.com/5idu/pilot/pkg/singleton"
	"github.com/5idu/pilot/pkg/xlog"

	"github.com/spf13/cast"
	"go.uber.org/zap"
)

// StdConfig ...
func StdConfig(name string) *Config {
	return RawConfig(constant.ConfigKey("registry." + name))
}

// RawConfig ...
func RawConfig(key string) *Config {
	var config = DefaultConfig()
	// 解析最外层配置
	if err := conf.UnmarshalKey(key, &config); err != nil {
		xlog.Panic("unmarshal key", xlog.String("mod", "registry.etcd"), xlog.Any("error", err), xlog.String("key", key), xlog.Any("config", config))
	}
	// 解析嵌套配置
	if err := conf.UnmarshalKey(key, &config.Config); err != nil {
		xlog.Panic("unmarshal key", xlog.String("mod", "registry.etcd"), xlog.Any("error", err), xlog.String("key", key), xlog.Any("config", config))
	}
	return config
}

// DefaultConfig ...
func DefaultConfig() *Config {
	return &Config{
		Config:      etcdv3.DefaultConfig(),
		ReadTimeout: time.Second * 3,
		logger:      xlog.With(xlog.String("mod", "registry.etcd")),
		ServiceTTL:  cast.ToDuration("60s"),
	}
}

// Config ...
type Config struct {
	*etcdv3.Config
	ReadTimeout time.Duration
	ConfigKey   string
	ServiceTTL  time.Duration
	logger      *xlog.Logger
}

// Build ...
func (config Config) Build() (registry.Registry, error) {
	if config.ConfigKey != "" {
		config.Config = etcdv3.RawConfig(config.ConfigKey)
	}
	return newETCDRegistry(&config)
}

func (config Config) MustBuild() registry.Registry {
	reg, err := config.Build()
	if err != nil {
		xlog.Panic("build registry failed", zap.Error(err))
	}
	return reg
}

func (config *Config) Singleton() (registry.Registry, error) {
	if val, ok := singleton.Load(constant.ModuleClientEtcd, config.ConfigKey); ok {
		return val.(registry.Registry), nil
	}

	reg, err := config.Build()
	if err != nil {
		return nil, err
	}

	singleton.Store(constant.ModuleClientEtcd, config.ConfigKey, reg)

	return reg, nil
}

func (config *Config) MustSingleton() registry.Registry {
	reg, err := config.Singleton()
	if err != nil {
		xlog.Panic("build registry failed", zap.Error(err))
	}

	return reg
}
