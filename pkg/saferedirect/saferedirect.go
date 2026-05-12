// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

package saferedirect

import (
	"context"
	"net/http"
	"net/url"
	"strings"
)

type (
	AllowedHostFunc func(ctx context.Context, host string) bool

	SafeRedirect struct {
		allowedHost AllowedHostFunc
	}
)

func New(allowedHost AllowedHostFunc) *SafeRedirect {
	return &SafeRedirect{allowedHost: allowedHost}
}

// StaticHosts returns an AllowedHost function that matches against a fixed
// list of hosts.
func StaticHosts(hosts ...string) AllowedHostFunc {
	allowed := make(map[string]bool, len(hosts))
	for _, h := range hosts {
		if h != "" {
			allowed[h] = true
		}
	}

	return func(_ context.Context, host string) bool {
		return allowed[host]
	}
}

func (sr *SafeRedirect) Validate(ctx context.Context, redirectURL string) (string, bool) {
	if redirectURL == "" {
		return "", false
	}

	if strings.HasPrefix(redirectURL, "/") {
		if len(redirectURL) > 1 && (redirectURL[1] == '/' || redirectURL[1] == '\\') {
			return "", false
		}
		return redirectURL, true
	}

	parsedURL, err := url.Parse(redirectURL)
	if err != nil {
		return "", false
	}

	if parsedURL.IsAbs() {
		if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
			return "", false
		}

		if sr.allowedHost != nil && !sr.allowedHost(ctx, parsedURL.Host) {
			return "", false
		}

		return redirectURL, true
	}

	return "", false
}

func (sr *SafeRedirect) GetSafeRedirectURL(ctx context.Context, redirectURL, fallbackURL string) string {
	if safeURL, isValid := sr.Validate(ctx, redirectURL); isValid {
		return safeURL
	}
	return fallbackURL
}

func (sr *SafeRedirect) Redirect(w http.ResponseWriter, r *http.Request, redirectURL, fallbackURL string, statusCode int) {
	safeURL := sr.GetSafeRedirectURL(r.Context(), redirectURL, fallbackURL)
	http.Redirect(w, r, safeURL, statusCode)
}
