package redis

import (
	"context"

	"github.com/5idu/pilot/pkg/constant"
	"github.com/5idu/pilot/pkg/singleton"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/metric"
)

var (
	ErrNotFound = redis.Nil
)

type Client struct {
	master *redis.Client
	config *Config

	metricCallbacks []metric.Registration
}

func (ins *Client) CmdOnMaster() *redis.Client {
	if ins.master == nil {
		panic(errors.New("redis: no master for " + ins.config.name))
	}
	return ins.master
}

func (ins *Client) Close() {
	if ins.master != nil {
		ins.master.Close()
	}
	if len(ins.metricCallbacks) > 0 {
		for _, cb := range ins.metricCallbacks {
			cb.Unregister()
		}
	}
}

// Singleton 单例模式
func (config *Config) Singleton() (*Client, error) {
	if val, ok := singleton.Load(constant.ModuleClientRedis, config.name); ok && val != nil {
		return val.(*Client), nil
	}

	cc, err := config.Build()
	if err != nil {
		return cc, err
	}
	singleton.Store(constant.ModuleClientRedis, config.name, cc)
	return cc, nil
}

// MustSingleton 单例模式
func (config *Config) MustSingleton() *Client {
	if val, ok := singleton.Load(constant.ModuleClientRedis, config.name); ok && val != nil {
		return val.(*Client)
	}
	cc := config.MustBuild()
	singleton.Store(constant.ModuleClientRedis, config.name, cc)
	return cc
}

// MustBuild ..
func (config *Config) MustBuild() *Client {
	cc, err := config.Build()
	if err != nil {
		panic(errors.WithMessage(err, "build redis failed"))
	}
	return cc
}

// Build ..
func (config *Config) Build() (*Client, error) {
	ins := new(Client)
	var err error
	ins.master, err = config.build(config.Addr, config.Username, config.Password)
	if err != nil {
		return ins, err
	}
	if ins.master == nil {
		return ins, errors.New("no master for " + config.name)
	}
	return ins, nil
}

func (config *Config) build(addr, user, pass string) (*redis.Client, error) {

	client := redis.NewClient(&redis.Options{
		Addr:                  addr,
		Username:              user,
		Password:              pass,
		DB:                    config.DB,
		MaxRetries:            config.MaxRetries,
		DialTimeout:           config.DialTimeout,
		ReadTimeout:           config.ReadTimeout,
		WriteTimeout:          config.WriteTimeout,
		PoolSize:              config.PoolSize,
		MinIdleConns:          config.MinIdleConns,
		ContextTimeoutEnabled: config.ContextTimeoutEnabled,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		panic(errors.WithMessage(err, "redis client ping failed"))
	}

	if config.EnableTrace {
		if err := redisotel.InstrumentTracing(client); err != nil {
			panic(errors.WithMessage(err, "new tracing failed"))
		}
	}
	if config.EnableMetric {
		if err := redisotel.InstrumentMetrics(client); err != nil {
			panic(errors.WithMessage(err, "new metrics failed"))
		}
	}
	if config.Debug {
		client.AddHook(debugInterceptor(config.name, addr, config, config.logger))
	}
	if config.EnableAccessLog {
		client.AddHook(accessInterceptor(config.name, addr, config, config.logger))
	}

	return client, nil
}
