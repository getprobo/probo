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

package saferedirect_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.probo.inc/probo/pkg/saferedirect"
)

func TestSafeRedirect_Validate(t *testing.T) {
	tests := []struct {
		name            string
		allowedHost     string
		redirectURL     string
		expectedURL     string
		expectedIsValid bool
	}{
		{
			name:            "empty redirect URL",
			allowedHost:     "example.com",
			redirectURL:     "",
			expectedURL:     "",
			expectedIsValid: false,
		},
		{
			name:            "relative URL",
			allowedHost:     "example.com",
			redirectURL:     "/dashboard",
			expectedURL:     "/dashboard",
			expectedIsValid: true,
		},
		{
			name:            "allowed absolute URL",
			allowedHost:     "example.com",
			redirectURL:     "https://example.com/dashboard",
			expectedURL:     "https://example.com/dashboard",
			expectedIsValid: true,
		},
		{
			name:            "disallowed host",
			allowedHost:     "example.com",
			redirectURL:     "https://evil.com/phishing",
			expectedURL:     "",
			expectedIsValid: false,
		},
		{
			name:            "disallowed scheme (javascript:)",
			allowedHost:     "example.com",
			redirectURL:     "javascript:alert('xss')",
			expectedURL:     "",
			expectedIsValid: false,
		},
		{
			name:            "disallowed scheme (data:)",
			allowedHost:     "example.com",
			redirectURL:     "data:text/html;base64,PHNjcmlwdD5hbGVydCgnWFNTJyk8L3NjcmlwdD4=",
			expectedURL:     "",
			expectedIsValid: false,
		},
		{
			name:            "no allowed host restriction",
			allowedHost:     "",
			redirectURL:     "https://any-domain.com/page",
			expectedURL:     "https://any-domain.com/page",
			expectedIsValid: true,
		},
		{
			name:            "invalid URL",
			allowedHost:     "example.com",
			redirectURL:     "https://[invalid-url",
			expectedURL:     "",
			expectedIsValid: false,
		},
		{
			name:            "double slash attack",
			allowedHost:     "example.com",
			redirectURL:     "//evil.com/phishing",
			expectedURL:     "",
			expectedIsValid: false,
		},
		{
			name:            "slash-backslash attack",
			allowedHost:     "example.com",
			redirectURL:     "/\\evil.com/phishing",
			expectedURL:     "",
			expectedIsValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sr := saferedirect.SafeRedirect{
				AllowedHost: tt.allowedHost,
			}

			gotURL, gotIsValid := sr.Validate(tt.redirectURL)
			if gotIsValid != tt.expectedIsValid {
				t.Errorf("Validate() isValid = %v, want %v", gotIsValid, tt.expectedIsValid)
			}
			if gotURL != tt.expectedURL {
				t.Errorf("Validate() url = %v, want %v", gotURL, tt.expectedURL)
			}
		})
	}
}

func TestSafeRedirect_GetSafeRedirectURL(t *testing.T) {
	tests := []struct {
		name        string
		allowedHost string
		redirectURL string
		fallbackURL string
		expectedURL string
	}{
		{
			name:        "safe redirect URL",
			allowedHost: "example.com",
			redirectURL: "/dashboard",
			fallbackURL: "/home",
			expectedURL: "/dashboard",
		},
		{
			name:        "unsafe redirect URL",
			allowedHost: "example.com",
			redirectURL: "https://evil.com/phishing",
			fallbackURL: "/home",
			expectedURL: "/home",
		},
		{
			name:        "empty redirect URL",
			allowedHost: "example.com",
			redirectURL: "",
			fallbackURL: "/home",
			expectedURL: "/home",
		},
		{
			name:        "double slash attack",
			allowedHost: "example.com",
			redirectURL: "//evil.com/phishing",
			fallbackURL: "/home",
			expectedURL: "/home",
		},
		{
			name:        "slash-backslash attack",
			allowedHost: "example.com",
			redirectURL: "/\\evil.com/phishing",
			fallbackURL: "/home",
			expectedURL: "/home",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sr := saferedirect.SafeRedirect{
				AllowedHost: tt.allowedHost,
			}

			gotURL := sr.GetSafeRedirectURL(tt.redirectURL, tt.fallbackURL)
			if gotURL != tt.expectedURL {
				t.Errorf("GetSafeRedirectURL() = %v, want %v", gotURL, tt.expectedURL)
			}
		})
	}
}

func TestSafeRedirect_Redirect(t *testing.T) {
	tests := []struct {
		name           string
		allowedHost    string
		redirectURL    string
		fallbackURL    string
		expectedStatus int
		expectedURL    string
	}{
		{
			name:           "safe redirect URL",
			allowedHost:    "example.com",
			redirectURL:    "/dashboard",
			fallbackURL:    "/home",
			expectedStatus: http.StatusFound,
			expectedURL:    "/dashboard",
		},
		{
			name:           "unsafe redirect URL",
			allowedHost:    "example.com",
			redirectURL:    "https://evil.com/phishing",
			fallbackURL:    "/home",
			expectedStatus: http.StatusFound,
			expectedURL:    "/home",
		},
		{
			name:           "empty redirect URL",
			allowedHost:    "example.com",
			redirectURL:    "",
			fallbackURL:    "/home",
			expectedStatus: http.StatusFound,
			expectedURL:    "/home",
		},
		{
			name:           "double slash attack",
			allowedHost:    "example.com",
			redirectURL:    "//evil.com/phishing",
			fallbackURL:    "/home",
			expectedStatus: http.StatusFound,
			expectedURL:    "/home",
		},
		{
			name:           "slash-backslash attack",
			allowedHost:    "example.com",
			redirectURL:    "/\\evil.com/phishing",
			fallbackURL:    "/home",
			expectedStatus: http.StatusFound,
			expectedURL:    "/home",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sr := saferedirect.SafeRedirect{
				AllowedHost: tt.allowedHost,
			}

			// Create a test HTTP recorder to capture the response
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "http://test.com", nil)

			sr.Redirect(w, r, tt.redirectURL, tt.fallbackURL, tt.expectedStatus)

			// Check that we got the expected status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Redirect() status = %v, want %v", w.Code, tt.expectedStatus)
			}

			// Check that the Location header contains the expected URL
			location := w.Header().Get("Location")
			if location != tt.expectedURL {
				t.Errorf("Redirect() location = %v, want %v", location, tt.expectedURL)
			}
		})
	}
}

func TestSafeRedirect_RedirectFromQuery(t *testing.T) {
	tests := []struct {
		name           string
		allowedHost    string
		queryParam     string
		queryValue     string
		fallbackURL    string
		expectedStatus int
		expectedURL    string
	}{
		{
			name:           "safe continue param",
			allowedHost:    "example.com",
			queryParam:     "continue",
			queryValue:     "/dashboard",
			fallbackURL:    "/home",
			expectedStatus: http.StatusFound,
			expectedURL:    "/dashboard",
		},
		{
			name:           "unsafe continue param",
			allowedHost:    "example.com",
			queryParam:     "continue",
			queryValue:     "https://evil.com/phishing",
			fallbackURL:    "/home",
			expectedStatus: http.StatusFound,
			expectedURL:    "/home",
		},
		{
			name:           "missing continue param",
			allowedHost:    "example.com",
			queryParam:     "continue",
			queryValue:     "",
			fallbackURL:    "/home",
			expectedStatus: http.StatusFound,
			expectedURL:    "/home",
		},
		{
			name:           "different query param name",
			allowedHost:    "example.com",
			queryParam:     "next",
			queryValue:     "/profile",
			fallbackURL:    "/home",
			expectedStatus: http.StatusFound,
			expectedURL:    "/profile",
		},
		{
			name:           "double slash attack in query",
			allowedHost:    "example.com",
			queryParam:     "continue",
			queryValue:     "//evil.com/phishing",
			fallbackURL:    "/home",
			expectedStatus: http.StatusFound,
			expectedURL:    "/home",
		},
		{
			name:           "slash-backslash attack in query",
			allowedHost:    "example.com",
			queryParam:     "continue",
			queryValue:     "/\\evil.com/phishing",
			fallbackURL:    "/home",
			expectedStatus: http.StatusFound,
			expectedURL:    "/home",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sr := saferedirect.SafeRedirect{
				AllowedHost: tt.allowedHost,
			}

			w := httptest.NewRecorder()

			url := "http://test.com"
			if tt.queryValue != "" {
				url = "http://test.com?" + tt.queryParam + "=" + tt.queryValue
			}
			r := httptest.NewRequest("GET", url, nil)

			sr.RedirectFromQuery(w, r, tt.queryParam, tt.fallbackURL, tt.expectedStatus)

			if w.Code != tt.expectedStatus {
				t.Errorf("RedirectFromQuery() status = %v, want %v", w.Code, tt.expectedStatus)
			}

			location := w.Header().Get("Location")
			if location != tt.expectedURL {
				t.Errorf("RedirectFromQuery() location = %v, want %v", location, tt.expectedURL)
			}
		})
	}
}
