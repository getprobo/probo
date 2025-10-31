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

package html2pdf

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"go.gearno.de/kit/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var (
	waitUntilDocumentReady = func(ctx context.Context) error {
		var ready bool
		return chromedp.Evaluate(`document.readyState === 'complete'`, &ready).Do(ctx)
	}

	tracerName = "go.probo.inc/probo/pkg/html2pdf"
)

type (
	RenderConfig struct {
		PageFormat      PageFormat
		Orientation     Orientation
		MarginTop       Margin
		MarginBottom    Margin
		MarginLeft      Margin
		MarginRight     Margin
		PrintBackground bool
		Scale           float64 // Print scale (0.1 to 2.0)
	}

	Option func(*Converter)

	Converter struct {
		l              *log.Logger
		tracerProvider trace.TracerProvider
		tracer         trace.Tracer
		addr           string
	}
)

func WithLogger(l *log.Logger) Option {
	return func(c *Converter) {
		c.l = l
	}
}

func WithTracerProvider(tp trace.TracerProvider) Option {
	return func(c *Converter) {
		c.tracerProvider = tp
	}
}

func NewConverter(addr string, opts ...Option) *Converter {
	// Only add ws:// prefix if it's not already present
	if !strings.HasPrefix(addr, "ws://") {
		addr = "ws://" + addr
	}

	c := &Converter{
		addr:           addr,
		l:              log.NewLogger(log.WithOutput(io.Discard)),
		tracerProvider: otel.GetTracerProvider(),
	}
	for _, opt := range opts {
		opt(c)
	}

	c.l = c.l.Named("html2pdf").With(log.String("addr", addr))
	c.tracer = c.tracerProvider.Tracer(tracerName)

	return c
}

func getPageDimensions(format PageFormat, orientation Orientation) (width, height float64) {
	var w, h float64

	switch format {
	case PageFormatA4:
		w, h = 8.27, 11.69 // A4 in inches
	case PageFormatLetter:
		w, h = 8.5, 11.0 // Letter in inches
	default:
		w, h = 8.27, 11.69 // Default to A4
	}

	if orientation == OrientationLandscape {
		return h, w // Swap width and height for landscape
	}
	return w, h
}

func (c *Converter) GeneratePDF(ctx context.Context, htmlDocument []byte, cfg RenderConfig) (io.Reader, error) {
	var (
		rootSpan = trace.SpanFromContext(ctx)
		span     trace.Span
	)

	if rootSpan.IsRecording() {
		ctx, span = c.tracer.Start(
			ctx,
			"GeneratePDF",
			trace.WithSpanKind(trace.SpanKindInternal),
		)
		defer span.End()
	}

	allocCtx, cancel := chromedp.NewRemoteAllocator(ctx, c.addr)
	defer cancel()

	ctx, cancel = chromedp.NewContext(allocCtx)
	defer cancel()

	dataURL := fmt.Sprintf(
		"data:text/html;base64,%s",
		base64.StdEncoding.EncodeToString(htmlDocument),
	)

	width, height := getPageDimensions(cfg.PageFormat, cfg.Orientation)

	marginTop := cfg.MarginTop.ToInches()
	marginBottom := cfg.MarginBottom.ToInches()
	marginLeft := cfg.MarginLeft.ToInches()
	marginRight := cfg.MarginRight.ToInches()

	scale := cfg.Scale
	if scale <= 0 {
		scale = 1.0
	}

	var pdfBytes []byte

	c.l.InfoCtx(
		ctx,
		"running chromedp",
		log.Float64("width", width),
		log.Float64("height", height),
		log.Float64("marginTop", marginTop),
		log.Float64("marginBottom", marginBottom),
		log.Float64("marginLeft", marginLeft),
		log.Float64("marginRight", marginRight),
		log.Float64("scale", scale),
		log.Bool("printBackground", cfg.PrintBackground),
	)

	err := chromedp.Run(ctx,
		chromedp.Navigate(dataURL),
		chromedp.WaitReady("body"),
		chromedp.ActionFunc(waitUntilDocumentReady),
		chromedp.ActionFunc(
			func(ctx context.Context) (err error) {
				pdfBytes, _, err = page.PrintToPDF().
					WithPrintBackground(cfg.PrintBackground).
					WithPaperWidth(width).
					WithPaperHeight(height).
					WithMarginTop(marginTop).
					WithMarginBottom(marginBottom).
					WithMarginLeft(marginLeft).
					WithMarginRight(marginRight).
					WithScale(scale).
					WithPreferCSSPageSize(false).
					Do(ctx)

				return err
			},
		),
	)

	if err != nil {
		err2 := fmt.Errorf("cannot run chromedp: %w", err)

		if rootSpan.IsRecording() {
			span.RecordError(err2)
			span.SetStatus(codes.Error, err2.Error())
		}

		return nil, err2
	}

	return bytes.NewReader(pdfBytes), nil
}
