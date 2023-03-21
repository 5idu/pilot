package xmetric

import (
	"github.com/5idu/pilot/pkg/constant"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
)

// UpDownCounterObserverVecOpts ...
type UpDownCounterObserverVecOpts struct {
	Name string
	Desc string
}

// Build ...
func (opts UpDownCounterObserverVecOpts) Build() *upDownCounterObserverVec {
	counter, err := global.Meter(constant.AppName()).Int64ObservableUpDownCounter(opts.Name,
		instrument.WithDescription(opts.Desc),
	)
	if err != nil {
		panic(errors.WithMessage(err, "buid int64 observe updown counter"))
	}
	return &upDownCounterObserverVec{counter}
}

// NewUpDownCounterObserverVecOpts ...
func NewUpDownCounterObserverVecOpts(name string, desc string) *upDownCounterObserverVec {
	return UpDownCounterObserverVecOpts{
		Name: name,
		Desc: desc,
	}.Build()
}

type upDownCounterObserverVec struct {
	instrument.Int64ObservableUpDownCounter
}

// Add ...
func (counter *upDownCounterObserverVec) Observe(f metric.Callback) (metric.Registration, error) {
	return global.Meter(constant.AppName()).RegisterCallback(f, counter.Int64ObservableUpDownCounter)
}
