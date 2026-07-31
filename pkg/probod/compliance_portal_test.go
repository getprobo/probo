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

package probod

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func textHandler(body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	})
}

func TestCompliancePortalHTTPHandler(t *testing.T) {
	t.Parallel()

	t.Run(
		"serves the portal directly when TLS is terminated by a proxy",
		func(t *testing.T) {
			t.Parallel()

			handler := compliancePortalHTTPHandler(
				true,
				textHandler("portal"),
				textHandler("acme-and-redirect"),
			)

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "http://trust.example.com/", nil))

			require.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "portal", w.Body.String())
		},
	)

	t.Run(
		"serves ACME challenges and HTTPS redirects by default",
		func(t *testing.T) {
			t.Parallel()

			handler := compliancePortalHTTPHandler(
				false,
				textHandler("portal"),
				textHandler("acme-and-redirect"),
			)

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "http://trust.example.com/", nil))

			require.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "acme-and-redirect", w.Body.String())
		},
	)
}
