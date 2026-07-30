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

package complianceportal_v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRewriteContinueURLLocale(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		continueURL string
		locale      string
		want        string
	}{
		{
			name:        "rewrites relative path locale",
			continueURL: "/es/documents",
			locale:      "fr",
			want:        "/fr/documents",
		},
		{
			name:        "preserves query markers",
			continueURL: "/es/documents?request-document-id=abc",
			locale:      "fr",
			want:        "/fr/documents?request-document-id=abc",
		},
		{
			name:        "rewrites absolute url path",
			continueURL: "https://acme.probopage.localhost/es/documents?subscribe=true",
			locale:      "fr",
			want:        "https://acme.probopage.localhost/fr/documents?subscribe=true",
		},
		{
			name:        "prepends locale on unprefixed path",
			continueURL: "/overview",
			locale:      "fr",
			want:        "/fr/overview",
		},
		{
			name:        "rewrites locale-only path",
			continueURL: "/es",
			locale:      "fr",
			want:        "/fr",
		},
		{
			name:        "leaves url unchanged for unsupported locale",
			continueURL: "/es/documents",
			locale:      "xx",
			want:        "/es/documents",
		},
		{
			name:        "leaves url unchanged when already matching",
			continueURL: "/fr/documents",
			locale:      "fr",
			want:        "/fr/documents",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := rewriteContinueURLLocale(tt.continueURL, tt.locale)
			assert.Equal(t, tt.want, got)
		})
	}
}
