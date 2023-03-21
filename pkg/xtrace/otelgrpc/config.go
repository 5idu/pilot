package otelgrpc

import (
	"context"
	"os"
	"time"

	"github.com/5idu/pilot/pkg/conf"
	"github.com/5idu/pilot/pkg/constant"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
		Endpoint: "localhost:4317",
		Sampler:  0,
	}
}

// Build ...
func (config *Config) Build() trace.TracerProvider {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, config.Endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(errors.WithMessage(err, "new otelgrpc"))
	}

	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		panic(errors.WithMessage(err, "new otelgrpc"))
	}

	hostname, _ := os.Hostname()
	tp := tracesdk.NewTracerProvider(
		// Set the sampling rate based on the parent span to 100%
		tracesdk.WithSampler(tracesdk.ParentBased(tracesdk.TraceIDRatioBased(config.Sampler))),
		// Always be sure to batch in production.
		tracesdk.WithBatcher(traceExporter),
		// Record information about this application in an Resource.
		tracesdk.WithResource(resource.NewSchemaless(
			semconv.TelemetrySDKLanguageGo,
			semconv.ServiceNameKey.String(config.Name),
			semconv.HostNameKey.String(hostname),
		)),
	)
	otel.SetTracerProvider(tp)
	return tp
}
