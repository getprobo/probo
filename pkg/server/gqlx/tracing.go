// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package gqlx

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
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

func (t TracingExtension) InterceptField(ctx context.Context, next graphql.Resolver) (any, error) {
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

	operationType := string(requestContext.Operation.Operation)

	operationName := requestContext.OperationName
	if operationName == "" {
		operationName = "unnamed"
	}

	var (
		errs     gqlerror.List
		finished bool
	)

	withSpan := func(ctx context.Context) context.Context {
		// gqlgen invokes the response handler with a different ctx than the
		// one passed to next(...). Re-attach the span so the logger can
		// extract trace_id.
		if operationSpan != nil {
			return trace.ContextWithSpan(ctx, operationSpan)
		}

		return ctx
	}

	finish := func() {
		if finished {
			return
		}

		finished = true

		if operationSpan != nil {
			operationSpan.End()
		}
	}

	report := func(ctx context.Context) {
		if finished {
			return
		}

		duration := time.Since(startTime)

		if len(errs) > 0 {
			t.logger.ErrorCtx(
				ctx,
				fmt.Sprintf("%s %s failed %s", operationType, operationName, duration.String()),
				log.String("graphql_operation_name", operationName),
				log.String("graphql_operation_type", operationType),
				log.Duration("graphql_operation_duration", duration),
				log.Int("graphql_operation_error_count", len(errs)),
				log.String("graphql_operation_error_codes", graphqlErrorCodes(errs)),
			)
		} else {
			t.logger.InfoCtx(
				ctx,
				fmt.Sprintf("%s %s succeed %s", operationType, operationName, duration.String()),
				log.String("graphql_operation_name", operationName),
				log.String("graphql_operation_type", operationType),
				log.Duration("graphql_operation_duration", duration),
			)
		}

		finish()
	}

	// A single-response transport such as POST stops after the first
	// payload even when it says hasNext, so the request closes whatever the
	// stream did not.
	if operations := openOperationsFromContext(ctx); operations != nil {
		operations.add(func() { report(withSpan(spanCtx)) })
	}

	return func(ctx context.Context) *graphql.Response {
		ctx = withSpan(ctx)

		resp := handler(ctx)

		// A nil response also ends a stream, so only a context error is a cancel.
		if resp == nil {
			if err := ctx.Err(); err != nil {
				duration := time.Since(startTime)

				t.logger.InfoCtx(
					ctx,
					fmt.Sprintf("%s %s canceled %s", operationType, operationName, duration.String()),
					log.String("graphql_operation_name", operationName),
					log.String("graphql_operation_type", operationType),
					log.Duration("graphql_operation_duration", duration),
					log.Error(err),
				)
			}

			finish()

			return nil
		}

		errs = append(errs, resp.Errors...)

		// With @defer the operation yields one response per payload; only the
		// last one, without hasNext, marks the end of the operation.
		if resp.HasNext == nil || !*resp.HasNext {
			report(ctx)
		}

		return resp
	}
}

var safeGraphQLErrorCodes = map[string]struct{}{
	"ACCOUNT_ALREADY_ACTIVATED": {},
	"ACTIVATION_NOT_REQUIRED":   {},
	"ALREADY_AUTHENTICATED":     {},
	"ASSUMPTION_REQUIRED":       {},
	"CONFLICT":                  {},
	"FORBIDDEN":                 {},
	"FULL_NAME_REQUIRED":        {},
	"INTERNAL":                  {},
	"INVALID":                   {},
	"MEMBERSHIP_REQUIRED":       {},
	"NDA_SIGNATURE_REQUIRED":    {},
	"NOT_FOUND":                 {},
	"TOKEN_ALREADY_USED":        {},
	"TOKEN_EXPIRED":             {},
	"UNAUTHENTICATED":           {},
	"UNAVAILABLE":               {},
}

// graphqlErrorCodes retains only server-owned, stable codes. gqlerror
// extensions can otherwise contain submitted validation values.
func graphqlErrorCodes(errs gqlerror.List) string {
	codes := make([]string, 0, len(errs))

	for _, err := range errs {
		code, _ := err.Extensions["code"].(string)
		if _, ok := safeGraphQLErrorCodes[code]; !ok {
			code = "UNKNOWN"
		}

		codes = append(codes, code)
	}

	slices.Sort(codes)

	return strings.Join(slices.Compact(codes), ",")
}

type openOperations struct {
	mu      sync.Mutex
	closers []func()
}

var openOperationsKey = &ctxKey{name: "open_operations"}

func withOpenOperations(ctx context.Context) (context.Context, *openOperations) {
	operations := &openOperations{}

	return context.WithValue(ctx, openOperationsKey, operations), operations
}

func openOperationsFromContext(ctx context.Context) *openOperations {
	operations, _ := ctx.Value(openOperationsKey).(*openOperations)

	return operations
}

func (o *openOperations) add(closer func()) {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.closers = append(o.closers, closer)
}

// close runs after the transport has returned, so no response handler is
// still running for these operations.
func (o *openOperations) close() {
	o.mu.Lock()
	closers := o.closers
	o.closers = nil
	o.mu.Unlock()

	for _, closer := range closers {
		closer()
	}
}
