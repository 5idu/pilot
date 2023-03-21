package xmetric

import (
	"github.com/5idu/pilot/pkg/constant"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
)

// Int64CounterObserverVecOpts ...
type Int64CounterObserverVecOpts struct {
	Name string
	Desc string
}

// Build ...
func (opts Int64CounterObserverVecOpts) Build() *int64CounterObserverVec {
	counter, err := global.Meter(constant.AppName()).Int64ObservableCounter(opts.Name,
		instrument.WithDescription(opts.Desc),
	)
	if err != nil {
		panic(errors.WithMessage(err, "buid int64 observe counter"))
	}
	return &int64CounterObserverVec{counter}
}

// NewInt64CounterObserverVecOpts ...
func NewInt64CounterObserverVecOpts(name string, desc string) *int64CounterObserverVec {
	return Int64CounterObserverVecOpts{
		Name: name,
		Desc: desc,
	}.Build()
}

type int64CounterObserverVec struct {
	instrument.Int64ObservableCounter
}

// Add ...
func (counter *int64CounterObserverVec) Observe(f metric.Callback) (metric.Registration, error) {
	return global.Meter(constant.AppName()).RegisterCallback(f, counter.Int64ObservableCounter)
}
