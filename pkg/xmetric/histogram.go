package xmetric

import (
	"context"
	"time"

	"github.com/5idu/pilot/pkg/constant"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
)

// HistogramVecOpts ...
type HistogramVecOpts struct {
	Name string
	Desc string
}

// Build ...
func (opts HistogramVecOpts) Build() *histogramVec {
	histogram, err := global.Meter(constant.AppName()).Int64Histogram(
		opts.Name,
		instrument.WithUnit("microseconds"),
		instrument.WithDescription(opts.Desc),
	)
	if err != nil {
		panic(errors.WithMessage(err, "buid histogram"))
	}

	return &histogramVec{histogram: histogram}
}

// NewHistogramVec ...
func NewHistogramVec(name string, desc string) *histogramVec {
	return HistogramVecOpts{
		Name: name,
		Desc: desc,
	}.Build()
}

type histogramVec struct {
	histogram instrument.Int64Histogram
}

// Observe ...
func (v *histogramVec) Record(ctx context.Context, dur time.Duration, attrs ...attribute.KeyValue) {
	v.histogram.Record(ctx, dur.Microseconds(), attrs...)
}
