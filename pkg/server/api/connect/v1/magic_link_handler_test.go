// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package connect_v1_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/securecookie"
	connect_v1 "go.probo.inc/probo/pkg/server/api/connect/v1"
)

func newTestMagicLinkHandler(t *testing.T) *connect_v1.MagicLinkHandler {
	t.Helper()

	return connect_v1.NewMagicLinkHandler(
		nil,
		baseurl.MustParse("https://auth.example.com"),
		securecookie.Config{Secret: "01234567890123456789012345678901"},
		log.NewLogger(log.WithOutput(io.Discard)),
		nil,
	)
}

func postMagicLinkForm(t *testing.T, handler http.HandlerFunc, values url.Values) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/connect/v1/magic-link/send",
		strings.NewReader(values.Encode()),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rec := httptest.NewRecorder()
	handler(rec, req)

	return rec
}

func TestMagicLinkHandler_SendHandler_Validation(t *testing.T) {
	t.Parallel()

	handler := newTestMagicLinkHandler(t)

	t.Run("rejects missing continue", func(t *testing.T) {
		t.Parallel()

		rec := postMagicLinkForm(
			t,
			handler.SendHandler,
			url.Values{
				"email": {"user@example.com"},
			},
		)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("rejects invalid email", func(t *testing.T) {
		t.Parallel()

		rec := postMagicLinkForm(
			t,
			handler.SendHandler,
			url.Values{
				"email":    {"not-an-email"},
				"continue": {"/"},
			},
		)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("rejects unsafe continue URL", func(t *testing.T) {
		t.Parallel()

		rec := postMagicLinkForm(
			t,
			handler.SendHandler,
			url.Values{
				"email":    {"user@example.com"},
				"continue": {"//evil.example.com"},
			},
		)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
