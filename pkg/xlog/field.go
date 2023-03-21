package xlog

import (
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	Field = zapcore.Field
	Opt   = zap.Option
)

var (
	// Strings ...
	Strings = zap.Strings
	// String ...
	String = zap.String
	// Any ...
	Any = zap.Any
	// Int64 ...
	Int64 = zap.Int64
	// Int ...
	Int = zap.Int
	// Int32 ...
	Int32 = zap.Int32
	// Uint ...
	Uint = zap.Uint
	// Duration ...
	Duration = zap.Duration
	// Durationp ...
	Durationp = zap.Durationp
	// Object ...
	Object = zap.Object
	// Namespace ...
	Namespace = zap.Namespace
	// Reflect ...
	Reflect = zap.Reflect
	// Skip ...
	Skip = zap.Skip()
	// ByteString ...
	ByteString = zap.ByteString
	// Time ...
	Time = zap.Time
	// Bool ...
	Bool = zap.Bool
	// Float64 ...
	Float64 = zap.Float64
)

// 耗时时间
func FieldCost(value time.Duration) Field {
	return String("cost", fmt.Sprintf("%.3f", float64(value.Round(time.Microsecond))/float64(time.Millisecond)))
}

// FieldErr ...
func FieldErr(err error) Field {
	return zap.Error(err)
}

// FieldExtra ...
func FieldExtra(value map[string]interface{}) Field {
	v, _ := sonic.MarshalString(value)
	return zap.String("extra", v)
}
