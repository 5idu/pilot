package xmetric

import (
	"context"

	"github.com/5idu/pilot/pkg/constant"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
)

// Int64CounterVecOpts ...
type Int64CounterVecOpts struct {
	Name string
	Desc string
}

// Build ...
func (opts Int64CounterVecOpts) Build() *int64CounterVec {
	counter, err := global.Meter(constant.AppName()).Int64Counter(opts.Name,
		instrument.WithUnit("1"),
		instrument.WithDescription(opts.Desc),
	)
	if err != nil {
		panic(errors.WithMessage(err, "buid int64 counter"))
	}
	return &int64CounterVec{counter: counter}
}

// NewInt64CounterVecOpts ...
func NewInt64CounterVecOpts(name string, desc string) *int64CounterVec {
	return Int64CounterVecOpts{
		Name: name,
		Desc: desc,
	}.Build()
}

type int64CounterVec struct {
	counter instrument.Int64Counter
}

// Add ...
func (v *int64CounterVec) Inc(ctx context.Context, attrs ...attribute.KeyValue) {
	v.counter.Add(ctx, 1, attrs...)
}
