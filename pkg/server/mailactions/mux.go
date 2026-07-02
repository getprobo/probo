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

// Package mailactions provides HTTP handlers for mailing list subscription
// management. All routes are mounted at /mail-actions/ with no API version
// prefix so they can be linked directly from emails and bookmarked by users.
//
//	GET  /mail-actions/unsubscribe   shows an unsubscribe confirmation page
//	POST /mail-actions/unsubscribe   RFC 8058 one-click unsubscribe
//	GET  /mail-actions/confirm       shows a subscription confirmation page
//	POST /mail-actions/confirm       confirms a pending subscription
package mailactions

import (
	"github.com/go-chi/chi/v5"
	"go.probo.inc/probo/pkg/mailman"
)

func NewMux(mailmanSvc *mailman.Service, tokenSecret string) *chi.Mux {
	r := chi.NewMux()

	r.Get("/unsubscribe", unsubscribeGetHandler())
	r.Post("/unsubscribe", unsubscribePostHandler(mailmanSvc, tokenSecret))

	r.Get("/confirm", confirmGetHandler())
	r.Post("/confirm", confirmPostHandler(mailmanSvc, tokenSecret))

	return r
}
