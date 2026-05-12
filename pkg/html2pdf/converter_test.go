// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
)

func TestNewConverter(t *testing.T) {
	tests := []struct {
		name string
		addr string
		opts []Option
	}{
		{
			name: "converter without options",
			addr: "ws://localhost:9222",
			opts: nil,
		},
		{
			name: "converter with logger option",
			addr: "ws://localhost:9222",
			opts: []Option{WithLogger(log.NewLogger())},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewConverter(tt.addr, tt.opts...)

			assert.NotNil(t, converter)
			assert.Equal(t, tt.addr, converter.addr)
			assert.NotNil(t, converter.l)
		})
	}
}

func TestWithLogger(t *testing.T) {
	logger := log.NewLogger()
	converter := &Converter{}

	option := WithLogger(logger)
	option(converter)

	assert.NotNil(t, converter.l)
}

func TestGetPageDimensions(t *testing.T) {
	tests := []struct {
		name        string
		format      PageFormat
		orientation Orientation
		wantWidth   float64
		wantHeight  float64
	}{
		{
			name:        "A4 portrait",
			format:      PageFormatA4,
			orientation: OrientationPortrait,
			wantWidth:   8.27,
			wantHeight:  11.69,
		},
		{
			name:        "A4 landscape",
			format:      PageFormatA4,
			orientation: OrientationLandscape,
			wantWidth:   11.69,
			wantHeight:  8.27,
		},
		{
			name:        "Letter portrait",
			format:      PageFormatLetter,
			orientation: OrientationPortrait,
			wantWidth:   8.5,
			wantHeight:  11.0,
		},
		{
			name:        "Letter landscape",
			format:      PageFormatLetter,
			orientation: OrientationLandscape,
			wantWidth:   11.0,
			wantHeight:  8.5,
		},
		{
			name:        "Unknown format defaults to A4 portrait",
			format:      PageFormat("unknown"),
			orientation: OrientationPortrait,
			wantWidth:   8.27,
			wantHeight:  11.69,
		},
		{
			name:        "Unknown format defaults to A4 landscape",
			format:      PageFormat("unknown"),
			orientation: OrientationLandscape,
			wantWidth:   11.69,
			wantHeight:  8.27,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			width, height := getPageDimensions(tt.format, tt.orientation)

			assert.Equal(t, tt.wantWidth, width)
			assert.Equal(t, tt.wantHeight, height)
		})
	}
}

func TestGeneratePDF_InvalidChrome(t *testing.T) {
	converter := NewConverter("ws://invalid:9999")

	htmlContent := []byte("<html><body><h1>Test</h1></body></html>")
	cfg := RenderConfig{
		PageFormat:      PageFormatA4,
		Orientation:     OrientationPortrait,
		MarginTop:       NewMarginInches(1.0),
		MarginBottom:    NewMarginInches(1.0),
		MarginLeft:      NewMarginInches(1.0),
		MarginRight:     NewMarginInches(1.0),
		PrintBackground: true,
		Scale:           1.0,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := converter.GeneratePDF(ctx, htmlContent, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot run chromedp")
}

func TestGeneratePDF_WithValidChrome(t *testing.T) {
	// Skip this test if Chrome is not available
	chromeAddr := os.Getenv("CHROME_WS_URL")
	if chromeAddr == "" {
		t.Skip("Skipping test: CHROME_WS_URL environment variable not set")
	}

	converter := NewConverter(chromeAddr)

	htmlContent := []byte(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>Test Document</title>
			<style>
				body { font-family: Arial, sans-serif; margin: 0; padding: 20px; }
				h1 { color: #333; }
				.content { background-color: #f0f0f0; padding: 10px; }
			</style>
		</head>
		<body>
			<h1>Test PDF Generation</h1>
			<div class="content">
				<p>This is a test document to verify PDF generation works correctly.</p>
				<p>It includes some basic styling and multiple elements.</p>
			</div>
		</body>
		</html>
	`)

	cfg := RenderConfig{
		PageFormat:      PageFormatA4,
		Orientation:     OrientationPortrait,
		MarginTop:       NewMarginInches(1.0),
		MarginBottom:    NewMarginInches(1.0),
		MarginLeft:      NewMarginInches(1.0),
		MarginRight:     NewMarginInches(1.0),
		PrintBackground: true,
		Scale:           1.0,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	reader, err := converter.GeneratePDF(ctx, htmlContent, cfg)
	require.NoError(t, err)
	require.NotNil(t, reader)

	// Read the PDF content to verify it's not empty
	pdfContent, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Greater(t, len(pdfContent), 0)

	// Verify it looks like a PDF file (starts with PDF magic bytes)
	assert.True(t, bytes.HasPrefix(pdfContent, []byte("%PDF-")))
}

func TestGeneratePDF_ScaleHandling(t *testing.T) {
	chromeAddr := os.Getenv("CHROME_WS_URL")
	if chromeAddr == "" {
		t.Skip("Skipping test: CHROME_WS_URL environment variable not set")
	}

	converter := NewConverter(chromeAddr)
	htmlContent := []byte("<html><body><h1>Scale Test</h1></body></html>")

	tests := []struct {
		name      string
		scale     float64
		wantScale float64
	}{
		{
			name:      "zero scale defaults to 1.0",
			scale:     0.0,
			wantScale: 1.0,
		},
		{
			name:      "negative scale defaults to 1.0",
			scale:     -0.5,
			wantScale: 1.0,
		},
		{
			name:      "valid scale preserved",
			scale:     1.5,
			wantScale: 1.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := RenderConfig{
				PageFormat:      PageFormatA4,
				Orientation:     OrientationPortrait,
				MarginTop:       NewMarginInches(1.0),
				MarginBottom:    NewMarginInches(1.0),
				MarginLeft:      NewMarginInches(1.0),
				MarginRight:     NewMarginInches(1.0),
				PrintBackground: false,
				Scale:           tt.scale,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			reader, err := converter.GeneratePDF(ctx, htmlContent, cfg)
			require.NoError(t, err)
			require.NotNil(t, reader)

			// Just verify we get a valid PDF - detailed scale verification would require
			// more complex PDF parsing
			pdfContent, err := io.ReadAll(reader)
			require.NoError(t, err)
			assert.Greater(t, len(pdfContent), 0)
			assert.True(t, bytes.HasPrefix(pdfContent, []byte("%PDF-")))
		})
	}
}

func TestRenderConfig_AllFormatsAndOrientations(t *testing.T) {
	chromeAddr := os.Getenv("CHROME_WS_URL")
	if chromeAddr == "" {
		t.Skip("Skipping test: CHROME_WS_URL environment variable not set")
	}

	converter := NewConverter(chromeAddr)
	htmlContent := []byte("<html><body><h1>Format Test</h1></body></html>")

	tests := []struct {
		name        string
		format      PageFormat
		orientation Orientation
	}{
		{
			name:        "A4 Portrait",
			format:      PageFormatA4,
			orientation: OrientationPortrait,
		},
		{
			name:        "A4 Landscape",
			format:      PageFormatA4,
			orientation: OrientationLandscape,
		},
		{
			name:        "Letter Portrait",
			format:      PageFormatLetter,
			orientation: OrientationPortrait,
		},
		{
			name:        "Letter Landscape",
			format:      PageFormatLetter,
			orientation: OrientationLandscape,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := RenderConfig{
				PageFormat:      tt.format,
				Orientation:     tt.orientation,
				MarginTop:       NewMarginInches(0.5),
				MarginBottom:    NewMarginInches(0.5),
				MarginLeft:      NewMarginInches(0.5),
				MarginRight:     NewMarginInches(0.5),
				PrintBackground: true,
				Scale:           1.0,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			reader, err := converter.GeneratePDF(ctx, htmlContent, cfg)
			require.NoError(t, err)
			require.NotNil(t, reader)

			pdfContent, err := io.ReadAll(reader)
			require.NoError(t, err)
			assert.Greater(t, len(pdfContent), 0)
			assert.True(t, bytes.HasPrefix(pdfContent, []byte("%PDF-")))
		})
	}
}

func TestGeneratePDF_WithDifferentMargins(t *testing.T) {
	chromeAddr := os.Getenv("CHROME_WS_URL")
	if chromeAddr == "" {
		t.Skip("Skipping test: CHROME_WS_URL environment variable not set")
	}

	converter := NewConverter(chromeAddr)
	htmlContent := []byte("<html><body><h1>Margin Test</h1></body></html>")

	tests := []struct {
		name         string
		marginTop    Margin
		marginBottom Margin
		marginLeft   Margin
		marginRight  Margin
	}{
		{
			name:         "equal margins in inches",
			marginTop:    NewMarginInches(1.0),
			marginBottom: NewMarginInches(1.0),
			marginLeft:   NewMarginInches(1.0),
			marginRight:  NewMarginInches(1.0),
		},
		{
			name:         "mixed margin units",
			marginTop:    NewMarginInches(0.5),
			marginBottom: NewMarginMillimeters(12.7),
			marginLeft:   NewMarginCentimeters(1.27),
			marginRight:  NewMarginPoints(36.0),
		},
		{
			name:         "zero margins",
			marginTop:    NewMarginInches(0.0),
			marginBottom: NewMarginInches(0.0),
			marginLeft:   NewMarginInches(0.0),
			marginRight:  NewMarginInches(0.0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := RenderConfig{
				PageFormat:      PageFormatA4,
				Orientation:     OrientationPortrait,
				MarginTop:       tt.marginTop,
				MarginBottom:    tt.marginBottom,
				MarginLeft:      tt.marginLeft,
				MarginRight:     tt.marginRight,
				PrintBackground: false,
				Scale:           1.0,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			reader, err := converter.GeneratePDF(ctx, htmlContent, cfg)
			require.NoError(t, err)
			require.NotNil(t, reader)

			pdfContent, err := io.ReadAll(reader)
			require.NoError(t, err)
			assert.Greater(t, len(pdfContent), 0)
			assert.True(t, bytes.HasPrefix(pdfContent, []byte("%PDF-")))
		})
	}
}

func BenchmarkGetPageDimensions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getPageDimensions(PageFormatA4, OrientationPortrait)
	}
}

func BenchmarkNewConverter(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewConverter("ws://localhost:9222")
	}
}
