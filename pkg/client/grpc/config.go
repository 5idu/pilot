package grpc

import (
	"time"

	"github.com/5idu/pilot/pkg/conf"
	"github.com/5idu/pilot/pkg/constant"
	"github.com/5idu/pilot/pkg/singleton"
	"github.com/5idu/pilot/pkg/xlog"

	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/keepalive"
)

// Config ...
type Config struct {
	Name           string // config's name
	BalancerName   string
	Addr           string
	DialTimeout    time.Duration
	ReadTimeout    time.Duration
	KeepAlive      *keepalive.ClientParameters
	RegistryConfig string

	logger      *xlog.Logger
	dialOptions []grpc.DialOption

	SlowThreshold time.Duration

	EnableTimeoutLog       bool
	EnableAccessLog        bool
	EnableTrace            bool
	EnableMetric           bool
	AccessInterceptorLevel string
}

// DefaultConfig ...
func DefaultConfig() *Config {
	return &Config{
		BalancerName:           roundrobin.Name, // round robin by default
		DialTimeout:            cast.ToDuration("3s"),
		ReadTimeout:            cast.ToDuration("1s"),
		SlowThreshold:          cast.ToDuration("600ms"),
		AccessInterceptorLevel: "info",
		EnableTimeoutLog:       true,
		EnableAccessLog:        true,
		EnableTrace:            true,
		EnableMetric:           true,
		KeepAlive: &keepalive.ClientParameters{
			Time:                5 * time.Minute,
			Timeout:             20 * time.Second,
			PermitWithoutStream: true,
		},
		RegistryConfig: constant.ConfigKey("registry.default"),
	}
}

// StdConfig ...
func StdConfig(name string) *Config {
	return RawConfig(constant.ConfigKey("grpc." + name))
}

// RawConfig ...
func RawConfig(key string) *Config {
	var config = DefaultConfig()
	config.Name = key
	if err := conf.UnmarshalKey(key, &config); err != nil {
		panic(errors.WithMessage(err, "client grpc parse config failed"))
	}
	return config
}

// WithLogger ...
func (config *Config) WithLogger(logger *xlog.Logger) *Config {
	config.logger = logger
	return config
}

// WithDialOption ...
func (config *Config) WithDialOption(opts ...grpc.DialOption) *Config {
	if config.dialOptions == nil {
		config.dialOptions = make([]grpc.DialOption, 0)
	}
	config.dialOptions = append(config.dialOptions, opts...)
	return config
}

// Build ...
func (config *Config) Build() *grpc.ClientConn {
	config.logger = xlog.With(xlog.String("mod", "client.grpc"))

	if config.EnableTimeoutLog {
		config.dialOptions = append(config.dialOptions,
			grpc.WithChainUnaryInterceptor(timeoutUnaryClientInterceptor(config.logger, config.ReadTimeout, config.SlowThreshold)),
		)
	}

	if config.EnableAccessLog {
		config.dialOptions = append(config.dialOptions,
			grpc.WithChainUnaryInterceptor(loggerUnaryClientInterceptor(config.logger, config.Name, config.AccessInterceptorLevel)),
		)
	}

	if config.EnableTrace {
		config.dialOptions = append(config.dialOptions,
			grpc.WithChainUnaryInterceptor(traceUnaryClientInterceptor()),
		)
	}

	if config.EnableMetric {
		config.dialOptions = append(config.dialOptions,
			grpc.WithChainUnaryInterceptor(metricUnaryClientInterceptor(config.Name)),
		)
	}

	return newGRPCClient(config)
}

// Singleton returns a singleton client conn.
func (config *Config) Singleton() (*grpc.ClientConn, error) {
	if val, ok := singleton.Load(constant.ModuleClientGrpc, config.Name); ok && val != nil {
		return val.(*grpc.ClientConn), nil
	}

	cc := config.Build()
	singleton.Store(constant.ModuleClientGrpc, config.Name, cc)

	return cc, nil
}

// MustSingleton panics when error found.
func (config *Config) MustSingleton() *grpc.ClientConn {
	cc, err := config.Singleton()
	if err != nil {
		panic(errors.WithMessage(err, "client grpc build client conn failed"))
	}

	return cc
}
