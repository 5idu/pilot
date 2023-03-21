package xgrpc

import (
	"fmt"

	"github.com/5idu/pilot/pkg/conf"
	"github.com/5idu/pilot/pkg/constant"
	"github.com/5idu/pilot/pkg/flag"
	"github.com/5idu/pilot/pkg/xlog"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// Config ...
type Config struct {
	Name string `json:"name"`
	Host string `json:"host"`
	Port int    `json:"port"`
	// Network network type, tcp4 by default
	Network string `json:"network"`
	// EnableAccessLog enable Access Interceptor, true by default
	EnableAccessLog bool
	// EnableTrace enable Trace Interceptor, true by default
	EnableTrace bool
	// EnableMetric enable Metric Interceptor, true by default
	EnableMetric bool
	// SlowQueryThresholdInMilli, request will be colored if cost over this threshold value
	SlowQueryThresholdInMilli int64
	// ServiceAddress service address in registry info, default to 'Host:Port'
	ServiceAddress string
	// EnableTLS
	EnableTLS bool
	// CaFile
	CaFile string
	// CertFile
	CertFile string
	// PrivateFile
	PrivateFile string

	Labels map[string]string `json:"labels"`

	serverOptions      []grpc.ServerOption
	streamInterceptors []grpc.StreamServerInterceptor
	unaryInterceptors  []grpc.UnaryServerInterceptor

	logger *xlog.Logger
}

// StdConfig represents Standard gRPC Server config
// which will parse config by conf package,
// panic if no config key found in conf
func StdConfig(name string) *Config {
	return RawConfig(constant.ConfigKey("server." + name))
}

// RawConfig ...
func RawConfig(key string) *Config {
	var config = DefaultConfig()
	if err := conf.UnmarshalKey(key, &config); err != nil {
		panic(errors.WithMessage(err, "grpc server parse config failed"))
	}
	return config
}

// DefaultConfig represents default config
// User should construct config base on DefaultConfig
func DefaultConfig() *Config {
	return &Config{
		Network:                   "tcp4",
		Host:                      flag.String("host"),
		Port:                      9092,
		EnableAccessLog:           true,
		EnableTLS:                 false,
		EnableTrace:               true,
		EnableMetric:              true,
		SlowQueryThresholdInMilli: 500,
		logger:                    xlog.With(xlog.String("mod", "grpc.server")),
		serverOptions:             []grpc.ServerOption{},
		streamInterceptors:        []grpc.StreamServerInterceptor{},
		unaryInterceptors:         []grpc.UnaryServerInterceptor{},
	}
}

// WithServerOption inject server option to grpc server
// User should not inject interceptor option, which is recommend by WithStreamInterceptor
// and WithUnaryInterceptor
func (config *Config) WithServerOption(options ...grpc.ServerOption) *Config {
	if config.serverOptions == nil {
		config.serverOptions = make([]grpc.ServerOption, 0)
	}
	config.serverOptions = append(config.serverOptions, options...)
	return config
}

// WithStreamInterceptor inject stream interceptors to server option
func (config *Config) WithStreamInterceptor(intes ...grpc.StreamServerInterceptor) *Config {
	if config.streamInterceptors == nil {
		config.streamInterceptors = make([]grpc.StreamServerInterceptor, 0)
	}

	config.streamInterceptors = append(config.streamInterceptors, intes...)
	return config
}

// WithUnaryInterceptor inject unary interceptors to server option
func (config *Config) WithUnaryInterceptor(intes ...grpc.UnaryServerInterceptor) *Config {
	if config.unaryInterceptors == nil {
		config.unaryInterceptors = make([]grpc.UnaryServerInterceptor, 0)
	}

	config.unaryInterceptors = append(config.unaryInterceptors, intes...)
	return config
}

func (config *Config) MustBuild() *Server {
	server, err := config.Build()
	if err != nil {
		xlog.Panic("build xgrpc server", xlog.Any("error", err))
	}
	return server
}

// Build ...
func (config *Config) Build() (*Server, error) {
	if config.EnableTrace {
		config.unaryInterceptors = append(config.unaryInterceptors, NewTraceUnaryServerInterceptor())
		config.streamInterceptors = append(config.streamInterceptors, NewTraceStreamServerInterceptor())
	}
	if config.EnableMetric {
		config.unaryInterceptors = append(config.unaryInterceptors, metricUnaryServerInterceptor)
		config.streamInterceptors = append(config.streamInterceptors, metricStreamServerInterceptor)
	}
	return newServer(config)
}

// WithLogger ...
func (config *Config) WithLogger(logger *xlog.Logger) *Config {
	config.logger = logger
	return config
}

// Address ...
func (config Config) Address() string {
	return fmt.Sprintf("%s:%d", config.Host, config.Port)
}
