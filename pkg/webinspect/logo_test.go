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

package webinspect_test

import (
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/webinspect"
)

func parseTestHTML(t *testing.T, rawURL string, body string) *webinspect.PageInfo {
	t.Helper()

	u, err := url.Parse(rawURL)
	require.NoError(t, err)

	info, err := webinspect.ParseHTML(u, strings.NewReader(body))
	require.NoError(t, err)

	return info
}

func TestFindLogoURL_SVGPreferred(t *testing.T) {
	t.Parallel()

	info := parseTestHTML(t, "https://example.com", `<html><head>
		<link rel="icon" type="image/svg+xml" href="/favicon.svg">
		<link rel="apple-touch-icon" href="/apple-touch-icon.png">
		<link rel="icon" sizes="192x192" href="/icon-192.png">
	</head><body></body></html>`)

	got, err := webinspect.FindLogoURL(info)
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/favicon.svg", got)
}

func TestFindLogoURL_AppleTouchIconSecond(t *testing.T) {
	t.Parallel()

	info := parseTestHTML(t, "https://example.com", `<html><head>
		<link rel="apple-touch-icon" sizes="180x180" href="/apple-touch-icon.png">
		<link rel="icon" sizes="32x32" href="/favicon-32.png">
	</head><body></body></html>`)

	got, err := webinspect.FindLogoURL(info)
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/apple-touch-icon.png", got)
}

func TestFindLogoURL_AppleTouchIconLargest(t *testing.T) {
	t.Parallel()

	info := parseTestHTML(t, "https://example.com", `<html><head>
		<link rel="apple-touch-icon" sizes="120x120" href="/touch-120.png">
		<link rel="apple-touch-icon" sizes="180x180" href="/touch-180.png">
		<link rel="apple-touch-icon" sizes="152x152" href="/touch-152.png">
	</head><body></body></html>`)

	got, err := webinspect.FindLogoURL(info)
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/touch-180.png", got)
}

func TestFindLogoURL_LargestIcon(t *testing.T) {
	t.Parallel()

	info := parseTestHTML(t, "https://example.com", `<html><head>
		<link rel="icon" sizes="32x32" href="/icon-32.png">
		<link rel="icon" sizes="192x192" href="/icon-192.png">
		<link rel="icon" sizes="16x16" href="/icon-16.png">
	</head><body></body></html>`)

	got, err := webinspect.FindLogoURL(info)
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/icon-192.png", got)
}

func TestFindLogoURL_MsTileImage(t *testing.T) {
	t.Parallel()

	info := parseTestHTML(t, "https://example.com", `<html><head>
		<meta name="msapplication-TileImage" content="/mstile-144.png">
	</head><body></body></html>`)

	got, err := webinspect.FindLogoURL(info)
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/mstile-144.png", got)
}

func TestFindLogoURL_RelativeHrefResolved(t *testing.T) {
	t.Parallel()

	info := parseTestHTML(t, "https://cdn.example.com/app/", `<html><head>
		<link rel="icon" type="image/svg+xml" href="../assets/logo.svg">
	</head><body></body></html>`)

	got, err := webinspect.FindLogoURL(info)
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example.com/assets/logo.svg", got)
}

func TestFindLogoURL_NoHead(t *testing.T) {
	t.Parallel()

	info := parseTestHTML(t, "https://example.com", `<html><body><p>no head</p></body></html>`)

	_, err := webinspect.FindLogoURL(info)
	assert.Error(t, err)
}

func TestFindLogoURL_NothingFound(t *testing.T) {
	t.Parallel()

	info := parseTestHTML(t, "https://example.com", `<html><head></head><body></body></html>`)

	_, err := webinspect.FindLogoURL(info)
	assert.Error(t, err)
}

func TestExtensionForMIME(t *testing.T) {
	t.Parallel()

	tests := []struct {
		contentType string
		expected    string
	}{
		{"image/svg+xml", ".svg"},
		{"image/png", ".png"},
		{"image/jpeg", ".jpg"},
		{"image/gif", ".gif"},
		{"image/webp", ".webp"},
		{"image/x-icon", ".ico"},
		{"image/vnd.microsoft.icon", ".ico"},
		{"image/png; charset=utf-8", ".png"},
		{"application/octet-stream", ".png"},
	}

	for _, tt := range tests {
		t.Run(
			tt.contentType,
			func(t *testing.T) {
				t.Parallel()
				assert.Equal(t, tt.expected, webinspect.ExtensionForMIME(tt.contentType))
			},
		)
	}
}
