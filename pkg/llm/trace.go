// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package llm

import (
	"context"
	"fmt"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
)

func startChatSpan(ctx context.Context, tracer trace.Tracer, system string, req *ChatCompletionRequest) (context.Context, trace.Span) {
	spanName := fmt.Sprintf("chat %s", req.Model)

	attrs := []attribute.KeyValue{
		semconv.GenAIOperationNameChat,
		semconv.GenAIProviderNameKey.String(system),
		semconv.GenAIRequestModel(req.Model),
	}
	if req.Temperature != nil {
		attrs = append(attrs, semconv.GenAIRequestTemperature(*req.Temperature))
	}
	if req.MaxTokens != nil {
		attrs = append(attrs, semconv.GenAIRequestMaxTokens(*req.MaxTokens))
	}
	if req.TopP != nil {
		attrs = append(attrs, semconv.GenAIRequestTopP(*req.TopP))
	}
	if len(req.StopSequences) > 0 {
		attrs = append(attrs, semconv.GenAIRequestStopSequences(req.StopSequences...))
	}

	return tracer.Start(
		ctx,
		spanName,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attrs...),
	)
}

func endChatSpan(span trace.Span, resp *ChatCompletionResponse, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.End()
		return
	}

	span.SetAttributes(
		semconv.GenAIResponseModel(resp.Model),
		semconv.GenAIUsageInputTokens(resp.Usage.InputTokens),
		semconv.GenAIUsageOutputTokens(resp.Usage.OutputTokens),
		semconv.GenAIResponseFinishReasons(string(resp.FinishReason)),
	)
	span.End()
}

// tracedStream wraps a ChatCompletionStream and manages the OTel span
// lifecycle for streaming calls. The span is ended when Close is called
// or when Next returns false (whichever comes first).
type tracedStream struct {
	inner        ChatCompletionStream
	span         trace.Span
	lastEvent    ChatCompletionStreamEvent
	closeOnce    sync.Once
	finishReason *FinishReason
	usage        *Usage
}

func newTracedStream(inner ChatCompletionStream, span trace.Span) *tracedStream {
	return &tracedStream{
		inner: inner,
		span:  span,
	}
}

func (s *tracedStream) Next() bool {
	if !s.inner.Next() {
		s.finalizeSpan()
		return false
	}
	s.lastEvent = s.inner.Event()
	if s.lastEvent.FinishReason != nil {
		s.finishReason = s.lastEvent.FinishReason
	}
	if s.lastEvent.Usage != nil {
		s.usage = s.lastEvent.Usage
	}
	return true
}

func (s *tracedStream) Event() ChatCompletionStreamEvent {
	return s.lastEvent
}

func (s *tracedStream) Err() error {
	return s.inner.Err()
}

func (s *tracedStream) Close() error {
	err := s.inner.Close()
	if err != nil {
		s.span.RecordError(err)
		s.span.SetStatus(codes.Error, err.Error())
	}
	s.finalizeSpan()
	return err
}

func (s *tracedStream) finalizeSpan() {
	s.closeOnce.Do(func() {
		if err := s.inner.Err(); err != nil {
			s.span.RecordError(err)
			s.span.SetStatus(codes.Error, err.Error())
			s.span.End()
			return
		}

		var attrs []attribute.KeyValue
		if s.usage != nil {
			attrs = append(attrs,
				semconv.GenAIUsageInputTokens(s.usage.InputTokens),
				semconv.GenAIUsageOutputTokens(s.usage.OutputTokens),
			)
		}
		if s.finishReason != nil {
			attrs = append(attrs, semconv.GenAIResponseFinishReasons(string(*s.finishReason)))
		}
		if len(attrs) > 0 {
			s.span.SetAttributes(attrs...)
		}
		s.span.End()
	})
}
