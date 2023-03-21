package xecho

import (
	"fmt"

	"github.com/5idu/pilot/pkg/conf"
	"github.com/5idu/pilot/pkg/constant"
	"github.com/5idu/pilot/pkg/flag"
	"github.com/5idu/pilot/pkg/xlog"

	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
)

// Config HTTP config
type Config struct {
	Host  string
	Port  int
	Debug bool
	// ServiceAddress service address in registry info, default to 'Host:Port'
	ServiceAddress string
	CertFile       string
	PrivateFile    string
	EnableTLS      bool

	EnableTrace  bool
	EnableMetric bool

	SlowQueryThresholdInMilli int64

	logger *xlog.Logger
}

// DefaultConfig ...
func DefaultConfig() *Config {
	return &Config{
		Host:                      flag.String("host"),
		Port:                      9091,
		Debug:                     false,
		SlowQueryThresholdInMilli: 500, // 500ms
		logger:                    xlog.With(xlog.String("mod", "echo.server")),
		EnableTLS:                 false,
		CertFile:                  "cert.pem",
		PrivateFile:               "private.pem",
		EnableTrace:               true,
		EnableMetric:              true,
	}
}

// StdConfig Jupiter Standard HTTP Server config
func StdConfig(name string) *Config {
	return RawConfig(constant.ConfigKey("server." + name))
}

// RawConfig ...
func RawConfig(key string) *Config {
	var config = DefaultConfig()
	if err := conf.UnmarshalKey(key, &config); err != nil && errors.Cause(err) != conf.ErrInvalidKey {
		panic(errors.WithMessage(err, "http server parse config error"))
	}
	return config
}

// WithLogger ...
func (config *Config) WithLogger(logger *xlog.Logger) *Config {
	config.logger = logger
	return config
}

// WithHost ...
func (config *Config) WithHost(host string) *Config {
	config.Host = host
	return config
}

// WithPort ...
func (config *Config) WithPort(port int) *Config {
	config.Port = port
	return config
}

func (config *Config) MustBuild() *Server {
	server, err := config.Build()
	if err != nil {
		panic(errors.WithMessage(err, "build echo server failed"))
	}
	return server
}

// Build create server instance, then initialize it with necessary interceptor
func (config *Config) Build() (*Server, error) {
	server, err := newServer(config)
	if err != nil {
		return nil, err
	}
	server.Use(recoverMiddleware(config.logger, config.SlowQueryThresholdInMilli))
	server.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowMethods:     []string{"*"},
		AllowCredentials: true,
	}))

	if config.EnableTrace {
		server.Use(traceServerInterceptor())
	}
	if config.EnableMetric {
		server.Use(metricServerInterceptor())
	}

	return server, nil
}

// Address ...
func (config *Config) Address() string {
	return fmt.Sprintf("%s:%d", config.Host, config.Port)
}
