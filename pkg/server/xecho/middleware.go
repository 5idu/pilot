package xecho

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/5idu/pilot/pkg/xlog"
	"github.com/5idu/pilot/pkg/xmetric"
	"github.com/5idu/pilot/pkg/xtrace"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
)

// RecoverMiddleware ...
func recoverMiddleware(logger *xlog.Logger, slowQueryThresholdInMilli int64) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) (err error) {
			var beg = time.Now()

			defer func() {
				extra := map[string]interface{}{
					"cost":   time.Since(beg).Seconds(),
					"method": ctx.Request().Method,
					"code":   ctx.Response().Status,
					"host":   ctx.Request().Host,
					"path":   ctx.Request().URL.Path,
				}
				if rec := recover(); rec != nil {
					log.Println(getCurrentGoroutineStack())
					switch rec := rec.(type) {
					case error:
						err = rec
					default:
						err = fmt.Errorf("%v", rec)
					}
				}
				if slowQueryThresholdInMilli > 0 {
					if cost := int64(time.Since(beg)) / 1e6; cost > slowQueryThresholdInMilli {
						extra["slow"] = cost
					}
				}
				if err != nil {
					extra["erros"] = err.Error()
					logger.Error("access", xlog.FieldExtra(extra))
					return
				}
				logger.Info("access", xlog.FieldExtra(extra))
			}()

			return next(ctx)
		}
	}
}

// getCurrentGoroutineStack 获取当前Goroutine的调用栈，便于排查panic异常
func getCurrentGoroutineStack() string {
	var buf [4096]byte
	n := runtime.Stack(buf[:], false)
	return string(buf[:n])
}

func traceServerInterceptor() echo.MiddlewareFunc {
	tracer := xtrace.NewTracer(trace.SpanKindServer)
	attrs := []attribute.KeyValue{
		semconv.HTTPServerNameKey.String("xecho"),
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			ctx, span := tracer.Start(c.Request().Context(), c.Path(), propagation.HeaderCarrier(c.Request().Header), trace.WithAttributes(attrs...))
			span.SetAttributes(
				semconv.HTTPServerAttributesFromHTTPRequest("", c.Request().URL.Path, c.Request())...)

			c.SetRequest(c.Request().WithContext(ctx))
			defer span.End()
			return next(c)
		}
	}
}

func metricServerInterceptor() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			beg := time.Now()
			err = next(c)
			method := c.Request().Method
			path := c.Path()
			peer := c.RealIP()
			xmetric.EchoServerDuration.Record(c.Request().Context(), time.Since(beg),
				attribute.String("method", method),
				attribute.String("path", path),
				attribute.String("peer", peer),
				attribute.String("status", http.StatusText(c.Response().Status)),
			)
			return err
		}
	}
}
