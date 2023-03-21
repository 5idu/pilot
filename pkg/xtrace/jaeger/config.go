package jaeger

import (
	"github.com/5idu/pilot/pkg/conf"
	"github.com/5idu/pilot/pkg/constant"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
)

type Config struct {
	Name     string
	Endpoint string
	Sampler  float64
}

// RawConfig ...
func RawConfig(key string) *Config {
	var config = DefaultConfig()
	if err := conf.UnmarshalKey(key, config); err != nil {
		panic(errors.WithMessage(err, "unmarshal key"))
	}
	return config
}

// DefaultConfig ...
func DefaultConfig() *Config {
	return &Config{
		Name:     constant.AppName(),
		Endpoint: "http://localhost:14268/api/traces",
		Sampler:  0,
	}
}

// Build ...
func (config *Config) Build() trace.TracerProvider {
	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(config.Endpoint)))
	if err != nil {
		panic(errors.WithMessage(err, "new jaeger"))
	}
	tp := tracesdk.NewTracerProvider(
		// Set the sampling rate based on the parent span to 100%
		tracesdk.WithSampler(tracesdk.ParentBased(tracesdk.TraceIDRatioBased(config.Sampler))),
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in an Resource.
		tracesdk.WithResource(resource.NewSchemaless(
			semconv.TelemetrySDKLanguageGo,
			semconv.ServiceNameKey.String(config.Name),
		)),
	)
	return tp
}
