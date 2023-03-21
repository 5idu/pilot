package request

import (
	"errors"
	"strings"
	"time"

	"github.com/5idu/pilot/pkg/conf"
	"github.com/5idu/pilot/pkg/constant"
	"github.com/5idu/pilot/pkg/xlog"
	"github.com/5idu/pilot/pkg/xmetric"

	"github.com/imroc/req/v3"
	"github.com/spf13/cast"
	"go.opentelemetry.io/otel/attribute"
)

var errSlowCommand = errors.New("http request slow command")

// Client ...
type Client = req.Client

// Config ...
type (
	// Config options
	Config struct {
		// 应用名称
		name string
		// Debug 开关
		Debug bool `json:"debug"`
		// 链路追踪开关
		EnableTrace bool `json:"enableTrace"`
		// 指标采集开关
		EnableMetric bool `json:"enableMetric"`
		// 失败重试次数
		RetryCount int `json:"retryCount"`
		// 失败重试的间隔时间
		RetryWaitTime time.Duration `json:"retryWaitTime"`
		// 请求超时时间
		Timeout time.Duration `json:"timeout"`
		// 慢日志阈值
		SlowThreshold time.Duration `json:"slowThreshold"`
		// 目标服务 base url
		Host string `json:"host"`
		// auth
		// like: {"type":"basic", "username": "", "password": ""}
		// 鉴权类型，目前只实现：basic、apikey
		Auth map[string]string `json:"auth"`
		// 日志
		logger *xlog.Logger
	}
)

// StdConfig 返回标准配置
func StdConfig(name string) *Config {
	return RawConfig(constant.ConfigKey("request." + name))
}

// RawConfig 返回配置
func RawConfig(key string) *Config {
	var config = DefaultConfig()
	if err := conf.UnmarshalKey(key, &config); err != nil {
		xlog.Panic("unmarshal config", xlog.FieldExtra(map[string]interface{}{"key": key}))
	}
	config.name = strings.Split(key, ".")[2]
	return config
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Debug:         false,
		EnableTrace:   true,
		EnableMetric:  true,
		RetryCount:    0,
		RetryWaitTime: cast.ToDuration("100ms"),
		SlowThreshold: cast.ToDuration("500ms"),
		Timeout:       cast.ToDuration("3000ms"),
		logger:        xlog.With(xlog.String("mod", "client.request")),
	}
}

func (config *Config) Build() (*req.Client, error) {
	if config.Host == "" {
		return nil, errors.New("addr not found")
	}

	c := req.C().
		SetBaseURL(config.Host).
		SetTimeout(config.Timeout).
		SetCommonRetryCount(config.RetryCount).
		SetRedirectPolicy(req.NoRedirectPolicy())

	if len(config.Auth) > 0 {
		authType := strings.TrimSpace(config.Auth["type"])
		switch authType {
		case "basic":
			c.SetCommonBasicAuth(config.Auth["username"], config.Auth["password"])
		case "apikey":
			addTo := strings.TrimSpace(config.Auth["addTo"])
			switch addTo {
			case "header":
				c.SetCommonHeader(config.Auth["key"], config.Auth["value"])
			default:
				c.SetCommonQueryParam(config.Auth["key"], config.Auth["value"])
			}
		default:
			return nil, errors.New("auth type not found")
		}
	}

	if config.Debug {
		c.EnableDebugLog()
	}
	if config.EnableTrace {
		traceInterceptor(config, c)
	}

	if config.RetryWaitTime != time.Duration(0) {
		c.SetCommonRetryInterval(func(resp *req.Response, attempt int) time.Duration {
			return config.RetryWaitTime
		})
	}

	if config.SlowThreshold > time.Duration(0) {
		c.OnAfterResponse(func(client *req.Client, resp *req.Response) error {
			// 慢日志
			if resp.TotalTime() > config.SlowThreshold {
				config.logger.Error("slow",
					xlog.FieldExtra(map[string]interface{}{
						"error":       errSlowCommand.Error(),
						"method":      resp.Request.Method,
						"cost":        resp.TotalTime(),
						"addr":        resp.Request.URL.String(),
						"status_code": resp.StatusCode,
					}),
				)
			}
			return nil
		})
	}

	if config.EnableMetric {
		c.OnAfterResponse(func(client *req.Client, resp *req.Response) error {
			attrs := []attribute.KeyValue{
				attribute.String("host", resp.Request.URL.Host),
				attribute.String("path", resp.Request.URL.Path),
				attribute.String("method", resp.Request.Method),
				attribute.String("status", resp.Status),
			}
			if resp.IsErrorState() {
				xmetric.HttpRequestFault.Inc(resp.Request.Context(), attrs...)
			}
			xmetric.HttpRequestDuration.Record(resp.Request.Context(), resp.TotalTime(), attrs...)
			return nil
		})
	}

	return c, nil
}

func (c *Config) MustBuild() *req.Client {
	cc, err := c.Build()
	if err != nil {
		c.logger.Panic("request build failed", xlog.FieldExtra(map[string]interface{}{"error": err.Error(), "config": c}))
	}

	return cc
}
