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

func confirmGetHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			renderPage(
				w,
				http.StatusBadRequest,
				page{
					Title:   "Invalid link",
					Heading: "Invalid link",
					Body:    "This confirmation link is missing required information. Please use the link from your email.",
				},
			)
			return
		}

		renderPage(
			w,
			http.StatusOK,
			// The confirmation should happen too quickly for the user to notice.
			// This content is only shown as a fallback.
			page{
				Title:   "Confirm subscription",
				Heading: "Confirm your subscription",
				Body:    "Your confirmation should be processed automatically. If it isn’t, click the button below to confirm that you want to receive updates.",
				Form: &form{
					ActionURL:  template.URL("?token=" + url.QueryEscape(token)),
					Button:     "Confirm subscription",
					AutoSubmit: true,
					Danger:     false,
				},
			},
		)
	}
}

func confirmPostHandler(mailmanSvc *mailman.Service, tokenSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			renderPage(
				w,
				http.StatusBadRequest,
				page{
					Title:   "Invalid link",
					Heading: "Invalid link",
					Body:    "This confirmation link is missing required information. Please use the link from your email.",
				},
			)
			return
		}

		data, err := mailman.ValidateConfirmToken(tokenSecret, token)
		if err != nil {
			renderPage(
				w,
				http.StatusUnauthorized,
				page{
					Title:   "Invalid link",
					Heading: "Invalid or expired link",
					Body:    "This confirmation link is invalid or has expired. Confirmation links are valid for 30 days — please re-subscribe to get a new one.",
				},
			)
			return
		}

		if err := mailmanSvc.ConfirmSubscriberByEmail(r.Context(), data.MailingListID, data.Email); err != nil {
			if errors.Is(err, mailman.ErrSubscriberNotFound) {
				renderPage(
					w,
					http.StatusNotFound, page{
						Title:   "Not found",
						Heading: "Subscription not found",
						Body:    "We could not find your subscription. It may have already been cancelled or this link was already used.",
					},
				)
				return
			}

			renderPage(
				w,
				http.StatusInternalServerError,
				page{
					Title:   "Something went wrong",
					Heading: "Something went wrong",
					Body:    "We could not confirm your subscription. Please try again later.",
				},
			)
			return
		}

		renderPage(
			w,
			http.StatusOK,
			page{
				Title:   "Subscription confirmed",
				Heading: "Subscription confirmed",
				Body:    "You're now subscribed and will receive updates.",
			},
		)
	}
}
