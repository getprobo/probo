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
