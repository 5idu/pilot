package xmongo

import (
	"context"

	"github.com/5idu/pilot/pkg/xtrace"

	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
)

var mongoCmdAttributeKey = attribute.Key("mongo.cmd")

func (c *Collection) startSpan(ctx context.Context, cmd string) (context.Context, trace.Span) {
	if !c.config.EnableTrace {
		return ctx, nil
	}

	tracer := xtrace.NewTracer(trace.SpanKindClient)
	attrs := []attribute.KeyValue{
		semconv.DBSystemMongoDB,
		semconv.DBNameKey.String(c.dbname),
		semconv.DBMongoDBCollectionKey.String(c.name),
		mongoCmdAttributeKey.String(cmd),
	}
	md := metadata.New(nil)

	_, span := tracer.Start(ctx, cmd, propagation.HeaderCarrier(md), trace.WithAttributes(attrs...))

	return ctx, span
}

func (c *Collection) endSpan(span trace.Span, err error) {
	if !c.config.EnableTrace {
		return
	}
	defer span.End()

	if err == nil || err == mongo.ErrNoDocuments ||
		err == mongo.ErrNilValue || err == mongo.ErrNilDocument {
		span.SetStatus(codes.Ok, "")
		return
	}

	span.SetStatus(codes.Error, err.Error())
	span.RecordError(err)
}
