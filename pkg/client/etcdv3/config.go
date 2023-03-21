package etcdv3

import (
	"time"

	"github.com/5idu/pilot/pkg/conf"
	"github.com/5idu/pilot/pkg/constant"
	"github.com/5idu/pilot/pkg/flag"
	"github.com/5idu/pilot/pkg/singleton"
	"github.com/5idu/pilot/pkg/xlog"

	"github.com/pkg/errors"
	"github.com/spf13/cast"
)

// Config ...
type (
	Config struct {
		Name      string   `json:"name"`
		Endpoints []string `json:"endpoints"`
		CertFile  string   `json:"certFile"`
		KeyFile   string   `json:"keyFile"`
		CaCert    string   `json:"caCert"`
		BasicAuth bool     `json:"basicAuth"`
		UserName  string   `json:"userName"`
		Password  string   `json:"-"`
		// 连接超时时间
		ConnectTimeout time.Duration `json:"connectTimeout"`
		Secure         bool          `json:"secure"`
		// 自动同步member list的间隔
		AutoSyncInterval time.Duration `json:"autoAsyncInterval"`
		TTL              int           // 单位：s
		EnableTrace      bool          `json:"enableTrace"`
		logger           *xlog.Logger
	}
)

func (config *Config) BindFlags(fs *flag.FlagSet) {
	fs.BoolVar(&config.Secure, "insecure-etcd", true, "--insecure-etcd=true")
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Endpoints:      []string{"http://localhost:2379"},
		BasicAuth:      false,
		ConnectTimeout: cast.ToDuration("5s"),
		Secure:         false,
		EnableTrace:    true,
		logger:         xlog.With(xlog.String("mod", "client.etcd")),
	}
}

// StdConfig ...
func StdConfig(name string) *Config {
	return RawConfig(constant.ConfigKey("etcdv3." + name))
}

// RawConfig ...
func RawConfig(key string) *Config {
	var config = DefaultConfig()
	config.Name = key

	if err := conf.UnmarshalKey(key, config); err != nil {
		panic(errors.WithMessage(err, "client etcd parse config failed"))
	}

	return config
}

// WithLogger ...
func (config *Config) WithLogger(logger *xlog.Logger) *Config {
	config.logger = logger
	return config
}

// Build ...
func (config *Config) Build() (*Client, error) {
	return newClient(config)
}

func (config *Config) Singleton() (*Client, error) {
	if client, ok := singleton.Load(constant.ModuleRegistryEtcd, config.Name); ok && client != nil {
		return client.(*Client), nil
	}

	client, err := config.Build()
	if err != nil {
		xlog.Error("build etcd client failed", xlog.Any("error", err))
		return nil, err
	}

	singleton.Store(constant.ModuleRegistryEtcd, config.Name, client)

	return client, nil
}

func (config *Config) MustBuild() *Client {
	client, err := config.Build()
	if err != nil {
		xlog.Panic("build etcd client failed", xlog.Any("error", err))
	}
	return client
}

func (config *Config) MustSingleton() *Client {
	client, err := config.Singleton()
	if err != nil {
		xlog.Panic("build etcd client failed", xlog.Any("error", err))
	}
	return client
}
