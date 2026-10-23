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

package html2pdf

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/remote"
	"go.gearno.de/kit/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var (
	waitUntilDocumentReady = func(ctx context.Context, t *chromedp.Target) error {
		_, err := chromedp.Evaluate[bool](`document.readyState === 'complete'`)(ctx, t)
		return err
	}

	tracerName = "go.probo.inc/probo/pkg/html2pdf"
)

type (
	RenderConfig struct {
		PageFormat               PageFormat
		Orientation              Orientation
		MarginTop                Margin
		MarginBottom             Margin
		MarginLeft               Margin
		MarginRight              Margin
		PrintBackground          bool
		Scale                    float64       // Print scale (0.1 to 2.0)
		WaitForExpression        string        // Optional JS expression polled until true before printing
		WaitForExpressionTimeout time.Duration // Max time to wait (default 15s)
		// GenerateTaggedPDF enables accessible PDF tagging (structure tree, reading
		// order). Nil defaults to true.
		GenerateTaggedPDF *bool
		// GenerateDocumentOutline embeds a document outline (bookmarks). Nil defaults
		// to true.
		GenerateDocumentOutline *bool
		// HeaderTemplate and FooterTemplate are Chrome print HTML. Either value
		// enables the print header and footer. Chrome fills elements with class
		// pageNumber and totalPages.
		HeaderTemplate string
		FooterTemplate string
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

func generateTaggedPDFEnabled(cfg RenderConfig) bool {
	if cfg.GenerateTaggedPDF != nil {
		return *cfg.GenerateTaggedPDF
	}

	return true
}

func generateDocumentOutlineEnabled(cfg RenderConfig) bool {
	if cfg.GenerateDocumentOutline != nil {
		return *cfg.GenerateDocumentOutline
	}

	return true
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

	allocCtx, cancel := remote.NewAllocator(ctx, c.addr)
	defer cancel()

	ctx, cancel = chromedp.NewContext(allocCtx)
	defer cancel()

	width, height := getPageDimensions(cfg.PageFormat, cfg.Orientation)

	marginTop := cfg.MarginTop.ToInches()
	marginBottom := cfg.MarginBottom.ToInches()
	marginLeft := cfg.MarginLeft.ToInches()
	marginRight := cfg.MarginRight.ToInches()

	scale := cfg.Scale
	if scale <= 0 {
		scale = 1.0
	}

	generateTaggedPDF := generateTaggedPDFEnabled(cfg)
	generateDocumentOutline := generateDocumentOutlineEnabled(cfg)

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
		log.Bool("generateTaggedPDF", generateTaggedPDF),
		log.Bool("generateDocumentOutline", generateDocumentOutline),
		log.Bool("displayHeaderFooter", cfg.HeaderTemplate != "" || cfg.FooterTemplate != ""),
	)

	var pdfBytes []byte

	htmlContent := string(htmlDocument)

	waitTimeout := cfg.WaitForExpressionTimeout
	if waitTimeout <= 0 {
		waitTimeout = 15 * time.Second
	}

	waitForExpr := chromedp.Func(func(ctx context.Context, t *chromedp.Target) error {
		if cfg.WaitForExpression == "" {
			return nil
		}

		deadline := time.Now().Add(waitTimeout)
		for time.Now().Before(deadline) {
			ready, err := chromedp.Evaluate[bool](cfg.WaitForExpression)(ctx, t)
			if err != nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			if ready {
				return nil
			}

			time.Sleep(100 * time.Millisecond)
		}

		return nil // proceed even on timeout
	})

	err := chromedp.Do(
		ctx,
		chromedp.Navigate("about:blank"),
		chromedp.Func(
			func(ctx context.Context, t *chromedp.Target) error {
				frameTree, err := cdp.Call(ctx, t, page.GetFrameTree, cdp.Empty{})
				if err != nil {
					return fmt.Errorf("cannot get frame tree: %w", err)
				}

				if _, err := cdp.Call(
					ctx,
					t,
					page.SetDocumentContent,
					page.SetDocumentContentParams{
						FrameID: frameTree.FrameTree.Frame.ID,
						HTML:    htmlContent,
					},
				); err != nil {
					return fmt.Errorf("cannot set document content: %w", err)
				}

				return nil
			}),
		chromedp.WaitReady("body"),
		chromedp.Func(waitUntilDocumentReady),
		waitForExpr,
		chromedp.Func(
			func(ctx context.Context, t *chromedp.Target) error {
				printParams := page.PrintToPDFParams{
					PrintBackground:         new(cfg.PrintBackground),
					PaperWidth:              width,
					PaperHeight:             height,
					MarginTop:               new(marginTop),
					MarginBottom:            new(marginBottom),
					MarginLeft:              new(marginLeft),
					MarginRight:             new(marginRight),
					Scale:                   scale,
					PreferCSSPageSize:       new(false),
					GenerateTaggedPDF:       new(generateTaggedPDF),
					GenerateDocumentOutline: new(generateDocumentOutline),
				}

				if cfg.HeaderTemplate != "" || cfg.FooterTemplate != "" {
					// Chrome draws a default date and title header when the
					// template string is empty.
					headerTemplate := cfg.HeaderTemplate
					if headerTemplate == "" {
						headerTemplate = "<span></span>"
					}

					footerTemplate := cfg.FooterTemplate
					if footerTemplate == "" {
						footerTemplate = "<span></span>"
					}

					printParams.DisplayHeaderFooter = new(true)
					printParams.HeaderTemplate = headerTemplate
					printParams.FooterTemplate = footerTemplate
				}

				res, err := cdp.Call(ctx, t, page.PrintToPDF, printParams)
				if err != nil {
					return err
				}

				pdfBytes = res.Data

				return nil
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
