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
