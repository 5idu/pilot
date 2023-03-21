package otelgrpc

import (
	"context"
	"time"

	"github.com/5idu/pilot/pkg/conf"
	"github.com/5idu/pilot/pkg/constant"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	Name     string
	Endpoint string
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
	}
}

// Build ...
func (config *Config) Build() metric.MeterProvider {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, config.Endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(errors.WithMessage(err, "new otelgrpc"))
	}

	// Set up a metric exporter
	exp, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn))
	if err != nil {
		panic(errors.WithMessage(err, "new otelgrpc"))
	}

	// // Print with a JSON encoder that indents with two spaces.
	// enc := json.NewEncoder(os.Stdout)
	// enc.SetIndent("", "  ")
	// exp, err := stdoutmetric.New(stdoutmetric.WithEncoder(enc))
	// if err != nil {
	// 	panic(err)
	// }

	mp := metricsdk.NewMeterProvider(
		metricsdk.WithReader(metricsdk.NewPeriodicReader(exp)),
		metricsdk.WithResource(resource.NewSchemaless(
			semconv.TelemetrySDKLanguageGo,
			semconv.ServiceNameKey.String(config.Name),
		)),
	)
	global.SetMeterProvider(mp)
	return mp
}
