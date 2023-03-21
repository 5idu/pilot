package request

import (
	"github.com/5idu/pilot/pkg/xtrace"

	"github.com/imroc/req/v3"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
)

func traceInterceptor(config *Config, c *req.Client) {
	tracer := xtrace.NewTracer(trace.SpanKindClient)
	attrs := []attribute.KeyValue{
		semconv.HTTPHostKey.String(config.Host),
		semconv.HTTPServerNameKey.String("request"),
		semconv.ServiceNameKey.String(config.name),
	}

	c.WrapRoundTripFunc(func(rt req.RoundTripper) req.RoundTripFunc {
		return func(req *req.Request) (resp *req.Response, err error) {
			ctx := req.Context()

			_, span := tracer.Start(ctx, req.URL.Path, propagation.HeaderCarrier(req.Headers), trace.WithAttributes(attrs...))
			span.SetAttributes(
				attribute.String("http.scheme", req.URL.Scheme),
				attribute.String("http.target", req.URL.String()),
				attribute.String("http.method", req.Method),
				attribute.String("http.req.header", req.HeaderToString()),
			)
			defer span.End()

			resp, err = rt.RoundTrip(req)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
			if resp.Response != nil {
				span.SetAttributes(
					semconv.HTTPStatusCodeKey.Int64(int64(resp.StatusCode)),
				)
			}
			return
		}
	})
}
