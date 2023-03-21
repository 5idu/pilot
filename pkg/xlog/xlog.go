package xlog

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/5idu/pilot/pkg/xtrace"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var logger *Logger

type Logger struct {
	ctx  context.Context
	zlog *zap.Logger
}

// Default returns default logger
func Default() *Logger {
	return logger
}

func newLoggerCore(c *Config) zapcore.Core {
	hook := newLogWriter(c)
	lvl := zap.NewAtomicLevelAt(getzaplogLevel(c.Level))

	encoderConfig := newZapEncoder()
	var encoder zapcore.Encoder
	if c.Format == FormatJSON {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(hook)),
		lvl,
	)
	return core
}

func newLogWriter(c *Config) io.Writer {
	if c.Dir == "" || c.Dir == "-" {
		return os.Stdout
	}
	return &lumberjack.Logger{
		Filename:   c.Filename(),
		Compress:   c.Compress,
		MaxBackups: c.MaxBackup,
		MaxSize:    c.MaxSize,
		MaxAge:     c.MaxAge,
		LocalTime:  true,
	}
}

func newZapEncoder() zapcore.EncoderConfig {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "line",
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder, // 小写编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,    // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, // 全路径编码器
		EncodeName:     zapcore.FullNameEncoder,
	}
	return encoderConfig
}

func newLoggerOptions() []zap.Option {
	// 开启开发模式，堆栈跟踪
	caller := zap.AddCaller()
	callerskip := zap.AddCallerSkip(1)
	// 开发者
	development := zap.Development()
	options := []zap.Option{
		caller,
		callerskip,
		development,
	}
	return options
}

func getzaplogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "panic":
		return zap.PanicLevel
	case "fatal":
		return zap.FatalLevel
	default:
		return zap.InfoLevel
	}
}

func (l *Logger) With(fields ...Field) *Logger {
	l.zlog = l.zlog.With(fields...)
	return l
}

func (l *Logger) clone() *Logger {
	copy := *l
	return &copy
}

func With(fields ...Field) *Logger {
	l := logger.clone()
	l.zlog = l.zlog.With(fields...)
	return l
}

func WithContext(ctx context.Context, fields ...Field) *Logger {
	l := logger.clone()
	l.ctx = ctx
	l.zlog = l.zlog.With(fields...)
	return l
}

func (l *Logger) buildFields(fields []Field) []Field {
	if fields == nil {
		fields = make([]Field, 0)
	}
	if l.ctx == nil {
		return fields
	}

	traceID := xtrace.TraceIDFromContext(l.ctx)
	if len(traceID) > 0 {
		fields = append(fields, String("request_id", traceID))
	}

	return fields
}

// Debug output log
func (l *Logger) Debug(msg string, fields ...Field) {
	l.zlog.Debug(msg, l.buildFields(fields)...)
}

// Debugf .
func (l *Logger) Debugf(msg string, args ...interface{}) {
	l.zlog.Debug(fmt.Sprintf(msg, args...), l.buildFields(nil)...)
}

// Info output log
func (l *Logger) Info(msg string, fields ...Field) {
	l.zlog.Info(msg, l.buildFields(fields)...)
}

// Infof .
func (l *Logger) Infof(msg string, args ...interface{}) {
	l.zlog.Info(fmt.Sprintf(msg, args...), l.buildFields(nil)...)
}

// Warn output log
func (l *Logger) Warn(msg string, fields ...Field) {
	l.zlog.Warn(msg, l.buildFields(fields)...)
}

// Warnf .
func (l *Logger) Warnf(msg string, args ...interface{}) {
	l.zlog.Warn(fmt.Sprintf(msg, args...), l.buildFields(nil)...)
}

// Error output log
func (l *Logger) Error(msg string, fields ...Field) {
	l.zlog.Error(msg, l.buildFields(fields)...)
}

// Errorf .
func (l *Logger) Errorf(msg string, args ...interface{}) {
	l.zlog.Error(fmt.Sprintf(msg, args...), l.buildFields(nil)...)
}

// Panic output panic
func (l *Logger) Panic(msg string, fields ...Field) {
	l.zlog.Panic(msg, l.buildFields(fields)...)
}

// Panicf .
func (l *Logger) Panicf(msg string, args ...interface{}) {
	l.zlog.Panic(fmt.Sprintf(msg, args...), l.buildFields(nil)...)
}

// Fatal output log
func (l *Logger) Fatal(msg string, fields ...Field) {
	l.zlog.Fatal(msg, l.buildFields(fields)...)
}

// Fatalf .
func (l *Logger) Fatalf(msg string, args ...interface{}) {
	l.zlog.Fatal(fmt.Sprintf(msg, args...), l.buildFields(nil)...)
}

// Sync flush buffered logs
func (l *Logger) Sync() {
	l.zlog.Sync()
}

// Debug output log
func Debug(msg string, fields ...Field) {
	logger.Debug(msg, fields...)
}

// Debugf .
func Debugf(msg string, args ...interface{}) {
	logger.Debugf(msg, args...)
}

// Info output log
func Info(msg string, fields ...Field) {
	logger.Info(msg, fields...)
}

// Infof .
func Infof(msg string, args ...interface{}) {
	logger.Infof(msg, args...)
}

// Warn output log
func Warn(msg string, fields ...Field) {
	logger.Warn(msg, fields...)
}

// Warnf .
func Warnf(msg string, args ...interface{}) {
	logger.Warnf(msg, args...)
}

// Error output log
func Error(msg string, fields ...Field) {
	logger.Error(msg, fields...)
}

// Errorf .
func Errorf(msg string, args ...interface{}) {
	logger.Errorf(msg, args...)
}

// Panic output panic
func Panic(msg string, fields ...Field) {
	logger.Panic(msg, fields...)
}

// Panicf .
func Panicf(msg string, args ...interface{}) {
	logger.Panicf(msg, args...)
}

// Fatal output log
func Fatal(msg string, fields ...Field) {
	logger.Fatal(msg, fields...)
}

// Fatalf .
func Fatalf(msg string, args ...interface{}) {
	logger.Fatalf(msg, args...)
}

// Sync flush buffered logs
func Sync() {
	logger.zlog.Sync()
}
