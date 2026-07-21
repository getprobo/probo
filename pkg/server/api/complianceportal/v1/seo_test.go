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

package complianceportal_v1_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	complianceportal_v1 "go.probo.inc/probo/pkg/server/api/complianceportal/v1"
)

func TestSEOFromRequest(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest(
		http.MethodGet,
		"https://acme.probopage.localhost/fr/documents",
		nil,
	)
	require.NoError(t, err)

	lang, canonical, hreflang := complianceportal_v1.SEOFromRequest(
		req,
		"https://acme.probopage.localhost",
	)
	assert.Equal(t, "fr", lang)
	assert.Equal(t, "https://acme.probopage.localhost/fr/documents", canonical)
	require.NotEmpty(t, hreflang)

	var xDefault string
	var enHref string
	for _, link := range hreflang {
		if link.Lang == "x-default" {
			xDefault = link.Href
		}
		if link.Lang == "en" {
			enHref = link.Href
		}
	}
	assert.Equal(t, "https://acme.probopage.localhost/en/documents", enHref)
	assert.Equal(t, enHref, xDefault)
}
