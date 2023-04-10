package redis

import (
	"time"

	"github.com/5idu/pilot/pkg/conf"
	"github.com/5idu/pilot/pkg/constant"
	"github.com/5idu/pilot/pkg/xlog"

	"github.com/pkg/errors"
	"github.com/spf13/cast"
)

// Config ...
type Config struct {
	Addr     string `json:"addr"`
	Username string `json:"username"`
	Password string `json:"password"`
	/****** for github.com/redis/go-redis/v9 ******/
	// DB default 0,not recommend
	DB int `json:"db"`
	// PoolSize applies per Stub node and not for the whole Stub.
	PoolSize int `json:"poolSize"`
	// Maximum number of retries before giving up.
	// Default is 3 retries; -1 (not 0) disables retries.
	MaxRetries int `json:"maxRetries"`
	// Minimum number of idle connections which is useful when establishing
	// new connection is slow.
	MinIdleConns int `json:"minIdleConns"`
	// Dial timeout for establishing new connections.
	// Default is 5 seconds.
	DialTimeout time.Duration `json:"dialTimeout"`
	// Timeout for socket reads. If reached, commands will fail
	// with a timeout instead of blocking. Use value 0 for no timeout and 0 for default.
	// Default is 3 seconds.
	ReadTimeout time.Duration `json:"readTimeout"`
	// Timeout for socket writes. If reached, commands will fail
	// with a timeout instead of blocking.
	// Default is ReadTimeout.
	WriteTimeout time.Duration `json:"writeTimeout"`
	// ContextTimeoutEnabled controls whether the client respects context timeouts and deadlines.
	// See https://redis.uptrace.dev/guide/go-redis-debugging.html#timeouts
	ContextTimeoutEnabled bool

	// nice option
	Debug bool `json:"debug"`

	// EnableAccessLog .. default false
	EnableAccessLog bool `json:"enableAccessLog"`
	// EnableTrace .. default true
	EnableTrace bool `json:"enableTrace"`
	// EnableMetric .. default true
	EnableMetric bool `json:"enableMetric"`

	logger *xlog.Logger
	name   string
}

// DefaultConfig default config ...
func DefaultConfig() *Config {
	return &Config{
		name:                  "default",
		Addr:                  "127.0.0.1:6379",
		DB:                    0,
		PoolSize:              200,
		MinIdleConns:          20,
		DialTimeout:           cast.ToDuration("3s"),
		ReadTimeout:           cast.ToDuration("1s"),
		WriteTimeout:          cast.ToDuration("1s"),
		ContextTimeoutEnabled: true,
		Debug:                 false,
		EnableAccessLog:       false,
		EnableTrace:           true,
		EnableMetric:          true,
		logger:                xlog.With(xlog.String("mod", "client.redis")),
	}
}

// StdConfig ...
func StdConfig(name string) *Config {
	return RawConfig(constant.ConfigKey("redis", name))
}
func RawConfig(key string) *Config {
	var config = DefaultConfig()

	if err := conf.UnmarshalKey(key, &config); err != nil {
		panic(errors.WithMessage(err, "unmarshal config: "+key))
	}
	config.name = key

	return config

}
