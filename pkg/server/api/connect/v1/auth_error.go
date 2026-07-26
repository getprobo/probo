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

package connect_v1

import (
	"net/http"
	"net/url"
)

const (
	authErrorPersonalAccountNotAllowed = "personal_account_not_allowed"
	authErrorEmailNotVerified          = "email_not_verified"
	authErrorInvalidState              = "invalid_state"
	authErrorAuthenticationFailed      = "authentication_failed"
	authErrorMagicLinkExpired          = "magic_link_expired"
	authErrorMagicLinkAlreadyUsed      = "magic_link_already_used"
	authErrorMagicLinkInvalid          = "magic_link_invalid"
)

func redirectAuthError(w http.ResponseWriter, r *http.Request, code string, continueURL string) {
	q := url.Values{}
	q.Set("error", code)
	if continueURL != "" {
		q.Set("continue", continueURL)
	}

	redirectURL := url.URL{
		Path:     "/auth/error",
		RawQuery: q.Encode(),
	}

	http.Redirect(w, r, redirectURL.String(), http.StatusFound)
}
