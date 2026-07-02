// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
