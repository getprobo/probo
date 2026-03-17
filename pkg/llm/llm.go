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
	"io"
	"time"

	"go.gearno.de/kit/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var tracerName = "go.probo.inc/probo/pkg/llm"

type (
	Option func(*Client)

	Client struct {
		provider       Provider
		system         string
		logger         *log.Logger
		tracerProvider trace.TracerProvider
		tracer         trace.Tracer
	}
)

func WithLogger(l *log.Logger) Option {
	return func(c *Client) {
		c.logger = l
	}
}

func WithTracerProvider(tp trace.TracerProvider) Option {
	return func(c *Client) {
		c.tracerProvider = tp
	}
}

// NewClient creates a new instrumented LLM client.
// The system parameter identifies the provider for the OTel gen_ai.provider.name
// attribute (e.g., "openai", "anthropic", "aws.bedrock").
func NewClient(provider Provider, system string, opts ...Option) *Client {
	c := &Client{
		provider:       provider,
		system:         system,
		logger:         log.NewLogger(log.WithOutput(io.Discard)),
		tracerProvider: otel.GetTracerProvider(),
	}

	for _, opt := range opts {
		opt(c)
	}

	c.logger = c.logger.Named("llm").With(log.String("system", system))
	c.tracer = c.tracerProvider.Tracer(tracerName)

	return c
}

func (c *Client) ChatCompletion(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	ctx, span := startChatSpan(ctx, c.tracer, c.system, req)

	c.logger.InfoCtx(
		ctx,
		"chat completion request",
		log.String("model", req.Model),
		log.Int("message_count", len(req.Messages)),
		log.Int("tool_count", len(req.Tools)),
	)

	start := time.Now()
	resp, err := c.provider.ChatCompletion(ctx, req)
	duration := time.Since(start)

	if err != nil {
		c.logger.ErrorCtx(
			ctx,
			"chat completion failed",
			log.String("model", req.Model),
			log.Duration("duration", duration),
			log.Error(err),
		)
		endChatSpan(span, nil, err)
		return nil, err
	}

	c.logger.InfoCtx(
		ctx,
		"chat completion response",
		log.String("model", resp.Model),
		log.Int("input_tokens", resp.Usage.InputTokens),
		log.Int("output_tokens", resp.Usage.OutputTokens),
		log.String("finish_reason", string(resp.FinishReason)),
		log.Duration("duration", duration),
	)

	endChatSpan(span, resp, nil)
	return resp, nil
}

func (c *Client) ChatCompletionStream(ctx context.Context, req *ChatCompletionRequest) (ChatCompletionStream, error) {
	ctx, span := startChatSpan(ctx, c.tracer, c.system, req)

	c.logger.InfoCtx(
		ctx,
		"chat completion stream request",
		log.String("model", req.Model),
		log.Int("message_count", len(req.Messages)),
		log.Int("tool_count", len(req.Tools)),
	)

	stream, err := c.provider.ChatCompletionStream(ctx, req)
	if err != nil {
		c.logger.ErrorCtx(
			ctx,
			"chat completion stream failed",
			log.String("model", req.Model),
			log.Error(err),
		)
		endChatSpan(span, nil, err)
		return nil, err
	}

	return newTracedStream(stream, span), nil
}
