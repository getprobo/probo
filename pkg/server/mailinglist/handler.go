// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

// Package mailinglist provides HTTP handlers for mailing list subscription
// management. All routes are mounted at /mailing-list/ with no API version
// prefix so they can be linked directly from emails and bookmarked by users.
//
//	GET  /mailing-list/unsubscribe   shows an unsubscribe confirmation page
//	POST /mailing-list/unsubscribe   RFC 8058 one-click unsubscribe
//	GET  /mailing-list/confirm       shows a subscription confirmation page
//	POST /mailing-list/confirm       confirms a pending subscription
package mailinglist

import (
	_ "embed"
	"errors"
	"html/template"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/mailman"
)

//go:embed templates/page.html.tmpl
var pageTmplHTML string

func NewMux(mailmanSvc *mailman.Service, tokenSecret string) *chi.Mux {
	r := chi.NewMux()

	r.Get("/unsubscribe", unsubscribeGetHandler())
	r.Post("/unsubscribe", unsubscribePostHandler(mailmanSvc, tokenSecret))

	r.Get("/confirm", confirmGetHandler())
	r.Post("/confirm", confirmPostHandler(mailmanSvc, tokenSecret))

	return r
}

func unsubscribeGetHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			renderPage(w, http.StatusBadRequest, page{
				Title:   "Invalid link",
				Heading: "Invalid link",
				Body:    "This unsubscribe link is missing required information. Please use the link from your email.",
			})
			return
		}

		renderPage(w, http.StatusOK, page{
			Title:   "Unsubscribe",
			Heading: "Unsubscribe from mailing list",
			Body:    "Click the button below to confirm that you no longer want to receive compliance updates.",
			Form: &form{
				ActionURL: template.URL("?token=" + url.QueryEscape(token)),
				Button:    "Confirm unsubscribe",
				Danger:    true,
			},
		})
	}
}

func unsubscribePostHandler(mailmanSvc *mailman.Service, tokenSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			http.Error(w, "missing token", http.StatusBadRequest)
			return
		}

		_ = r.ParseForm()

		data, err := mailman.ValidateUnsubscribeToken(tokenSecret, token)
		if err != nil {
			renderPage(w, http.StatusUnauthorized, page{
				Title:   "Invalid link",
				Heading: "Invalid or expired link",
				Body:    "This unsubscribe link is invalid or has expired.",
			})
			return
		}

		mailingListID, err := gid.ParseGID(data.MailingListID)
		if err != nil {
			http.Error(w, "invalid token data", http.StatusBadRequest)
			return
		}

		recipientEmail, err := mail.ParseAddr(data.Email)
		if err != nil {
			http.Error(w, "invalid token data", http.StatusBadRequest)
			return
		}

		if err := mailmanSvc.UnsubscribeByEmail(r.Context(), mailingListID, recipientEmail); err != nil {
			if !errors.Is(err, mailman.ErrSubscriberNotFound) {
				renderPage(w, http.StatusInternalServerError, page{
					Title:   "Something went wrong",
					Heading: "Something went wrong",
					Body:    "We could not process your request. Please try again later.",
				})
				return
			}
		}

		// Also success when already unsubscribed — unsubscribe is idempotent
		// per RFC 8058.
		renderPage(w, http.StatusOK, page{
			Title:   "Unsubscribed",
			Heading: "You've been unsubscribed",
			Body:    "You will no longer receive compliance updates from this mailing list.",
		})
	}
}

func confirmGetHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			renderPage(w, http.StatusBadRequest, page{
				Title:   "Invalid link",
				Heading: "Invalid link",
				Body:    "This confirmation link is missing required information. Please use the link from your email.",
			})
			return
		}

		renderPage(w, http.StatusOK, page{
			Title:   "Confirm subscription",
			Heading: "Confirm your subscription",
			Body:    "Click the button below to confirm that you want to receive compliance updates.",
			Form: &form{
				ActionURL: template.URL("?token=" + url.QueryEscape(token)),
				Button:    "Confirm subscription",
				Danger:    false,
			},
		})
	}
}

func confirmPostHandler(mailmanSvc *mailman.Service, tokenSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			http.Error(w, "missing token", http.StatusBadRequest)
			return
		}

		data, err := mailman.ValidateConfirmToken(tokenSecret, token)
		if err != nil {
			renderPage(w, http.StatusUnauthorized, page{
				Title:   "Invalid link",
				Heading: "Invalid or expired link",
				Body:    "This confirmation link is invalid or has expired. Confirmation links are valid for 30 days — please re-subscribe to get a new one.",
			})
			return
		}

		mailingListID, err := gid.ParseGID(data.MailingListID)
		if err != nil {
			http.Error(w, "invalid token data", http.StatusBadRequest)
			return
		}

		recipientEmail, err := mail.ParseAddr(data.Email)
		if err != nil {
			http.Error(w, "invalid token data", http.StatusBadRequest)
			return
		}

		if err := mailmanSvc.ConfirmSubscriberByEmail(r.Context(), mailingListID, recipientEmail); err != nil {
			if errors.Is(err, mailman.ErrSubscriberNotFound) {
				renderPage(w, http.StatusNotFound, page{
					Title:   "Not found",
					Heading: "Subscription not found",
					Body:    "We could not find your subscription. It may have already been cancelled or this link was already used.",
				})
				return
			}

			renderPage(w, http.StatusInternalServerError, page{
				Title:   "Something went wrong",
				Heading: "Something went wrong",
				Body:    "We could not confirm your subscription. Please try again later.",
			})
			return
		}

		renderPage(w, http.StatusOK, page{
			Title:   "Subscription confirmed",
			Heading: "Subscription confirmed",
			Body:    "You're now subscribed and will receive compliance updates.",
		})
	}
}

type form struct {
	ActionURL template.URL
	Button    string
	Danger    bool
}

type page struct {
	Title   string
	Heading string
	Body    string
	Form    *form
}

var tmpl = template.Must(template.New("page.html.tmpl").Parse(pageTmplHTML))

func renderPage(w http.ResponseWriter, status int, p page) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_ = tmpl.Execute(w, p)
}
