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

package mailactions

import (
	"errors"
	"html/template"
	"net/http"
	"net/url"

	"go.probo.inc/probo/pkg/mailman"
)

func unsubscribeGetHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			renderPage(
				w,
				http.StatusBadRequest,
				page{
					Title:   "Invalid link",
					Heading: "Invalid link",
					Body:    "This unsubscribe link is missing required information. Please use the link from your email.",
				},
			)
			return
		}

		renderPage(
			w,
			http.StatusOK,
			page{
				Title:   "Unsubscribe",
				Heading: "Unsubscribe from mailing list",
				Body:    "Click the button below to confirm that you no longer want to receive updates.",
				Form: &form{
					ActionURL: template.URL("?token=" + url.QueryEscape(token)),
					Button:    "Confirm unsubscribe",
					Danger:    true,
				},
			},
		)
	}
}

func unsubscribePostHandler(mailmanSvc *mailman.Service, tokenSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			renderPage(
				w,
				http.StatusBadRequest,
				page{
					Title:   "Invalid link",
					Heading: "Invalid link",
					Body:    "This unsubscribe link is missing required information. Please use the link from your email.",
				},
			)
			return
		}

		data, err := mailman.ValidateUnsubscribeToken(tokenSecret, token)
		if err != nil {
			renderPage(
				w,
				http.StatusUnauthorized,
				page{
					Title:   "Invalid link",
					Heading: "Invalid or expired link",
					Body:    "This unsubscribe link is invalid or has expired.",
				},
			)
			return
		}

		if err := mailmanSvc.UnsubscribeByEmail(r.Context(), data.MailingListID, data.Email); err != nil {
			if !errors.Is(err, mailman.ErrSubscriberNotFound) {
				renderPage(
					w,
					http.StatusInternalServerError,
					page{
						Title:   "Something went wrong",
						Heading: "Something went wrong",
						Body:    "We could not process your request. Please try again later.",
					},
				)
				return
			}
		}

		// Also success when already unsubscribed — unsubscribe is idempotent
		// per RFC 8058.
		renderPage(
			w,
			http.StatusOK,
			page{
				Title:   "Unsubscribed",
				Heading: "You've been unsubscribed",
				Body:    "You will no longer receive updates.",
			},
		)
	}
}
