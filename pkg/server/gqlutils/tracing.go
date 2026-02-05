// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package gqlutils

import (
	"context"
	"fmt"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"go.gearno.de/kit/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type TracingExtension struct {
	logger *log.Logger
}

func NewTracingExtension(logger *log.Logger) TracingExtension {
	return TracingExtension{logger: logger}
}

func (t TracingExtension) ExtensionName() string {
	return "Tracing"
}

func (t TracingExtension) Validate(schema graphql.ExecutableSchema) error {
	return nil
}

func (t TracingExtension) InterceptField(ctx context.Context, next graphql.Resolver) (interface{}, error) {
	rootSpan := trace.SpanFromContext(ctx)

	if rootSpan.IsRecording() {
		tracer := otel.Tracer("graphql-field")
		fieldContext := graphql.GetFieldContext(ctx)

		ctx, span := tracer.Start(ctx, "GraphQL Field: "+fieldContext.Field.Name)
		defer span.End()

		span.SetAttributes(
			attribute.String("graphql.field.name", fieldContext.Field.Name),
			attribute.String("graphql.field.path", fieldContext.Path().String()),
			attribute.String("graphql.field.object", fieldContext.Object),
		)

		result, err := next(ctx)

		if err != nil {
			span.RecordError(err)
		}

		return result, err
	}

	return next(ctx)
}

func (t TracingExtension) InterceptOperation(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	requestContext := graphql.GetOperationContext(ctx)
	startTime := time.Now()

	rootSpan := trace.SpanFromContext(ctx)
	spanCtx := ctx
	var operationSpan trace.Span

	if rootSpan.IsRecording() {
		tracer := otel.Tracer("graphql-operation")
		operationName := "GraphQL Operation"
		if requestContext.OperationName != "" {
			operationName = "GraphQL " + requestContext.OperationName
		}

		spanCtx, operationSpan = tracer.Start(ctx, operationName)

		operationSpan.SetAttributes(
			attribute.String("graphql.operation_name", requestContext.OperationName),
			attribute.String("graphql.operation_type", string(requestContext.Operation.Operation)),
			attribute.String("graphql.query", requestContext.RawQuery),
		)
	}

	handler := next(spanCtx)

	return func(ctx context.Context) *graphql.Response {
		// gqlgen invokes the response handler with a different ctx than the one
		// passed to next(...). Re-attach the span so the logger can extract trace_id.
		if operationSpan != nil {
			ctx = trace.ContextWithSpan(ctx, operationSpan)
			defer operationSpan.End()
		}

		resp := handler(ctx)
		duration := time.Since(startTime)

		operationType := string(requestContext.Operation.Operation)
		operationName := requestContext.OperationName
		if operationName == "" {
			operationName = "unnamed"
		}

		if resp.Errors != nil {
			t.logger.ErrorCtx(ctx,
				fmt.Sprintf("%s %s failed %s", operationType, operationName, duration.String()),
				log.String("graphql_operation_name", operationName),
				log.String("graphql_operation_type", operationType),
				log.Duration("graphql_operation_duration", duration),
				log.Any("graphql_operation_errors", resp.Errors),
			)
		} else {
			t.logger.InfoCtx(ctx,
				fmt.Sprintf("%s %s succeed %s", operationType, operationName, duration.String()),
				log.String("graphql_operation_name", operationName),
				log.String("graphql_operation_type", operationType),
				log.Duration("graphql_operation_duration", duration),
			)
		}

		return resp
	}
}
