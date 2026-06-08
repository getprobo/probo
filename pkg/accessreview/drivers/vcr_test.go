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

package drivers

import (
	"encoding/base64"
	"net/http"
	"os"
	"testing"

	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/recorder"
)

// versionedClientHeaders are HTTP headers that encode SDK or client library
// versions. They are ignored by the matcher so cassettes keep replaying after
// dependency bumps.
var versionedClientHeaders = []string{"User-Agent", "X-Goog-Api-Client"}

// newRecorder creates a go-vcr recorder for the given cassette path. When
// the env var is non-empty the recorder runs in record mode, otherwise
// it replays from the committed cassette. A BeforeSave hook strips the
// Authorization header so tokens are never persisted.
func newRecorder(t *testing.T, cassettePath string, envVar string) *recorder.Recorder {
	t.Helper()

	mode := recorder.ModeReplayOnly
	if os.Getenv(envVar) != "" {
		mode = recorder.ModeRecordOnly
	}

	rec, err := recorder.New(
		cassettePath,
		recorder.WithMode(mode),
		recorder.WithSkipRequestLatency(true),
		recorder.WithMatcher(cassette.NewDefaultMatcher(
			cassette.WithIgnoreAuthorization(),
			cassette.WithIgnoreHeaders(versionedClientHeaders...),
		)),
		recorder.WithHook(func(i *cassette.Interaction) error {
			i.Request.Headers.Del("Authorization")
			// Providers like Anthropic (x-api-key) and SigNoz
			// (SIGNOZ-API-KEY) authenticate via a custom header rather
			// than Authorization; strip those too so a re-record never
			// persists a raw key.
			i.Request.Headers.Del("X-Api-Key")
			i.Request.Headers.Del("Signoz-Api-Key")

			return nil
		}, recorder.BeforeSaveHook),
	)
	if err != nil {
		if mode == recorder.ModeReplayOnly {
			t.Skipf("cassette not found (record with %s env var): %v", envVar, err)
		}

		t.Fatalf("cannot create vcr recorder: %v", err)
	}

	t.Cleanup(func() {
		if err := rec.Stop(); err != nil {
			t.Errorf("cannot stop vcr recorder: %v", err)
		}
	})

	return rec
}

// authRoundTripper wraps a transport and injects an Authorization header
// into each request. The authValue is set as-is (caller provides "Bearer xxx"
// or a raw API key depending on the provider).
type authRoundTripper struct {
	authValue string
	transport http.RoundTripper
}

func (rt *authRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if rt.authValue != "" {
		req.Header.Set("Authorization", rt.authValue)
	}

	return rt.transport.RoundTrip(req)
}

// bearerAuth returns "Bearer <token>" if the token is non-empty, or "" otherwise.
func bearerAuth(token string) string {
	if token == "" {
		return ""
	}

	return "Bearer " + token
}

// basicAuth returns the HTTP Basic auth header value for a username with
// an empty password ("Basic base64(<username>:)"), or "" if the username
// is empty. Cursor presents its admin API key as the Basic auth username.
func basicAuth(username string) string {
	if username == "" {
		return ""
	}

	return "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"))
}

// newVCRClient creates an *http.Client backed by the recorder's transport,
// with an optional Authorization header injected into requests (for recording
// mode). The authValue should be the complete header value, e.g.
// "Bearer xxx" or a raw API key like "lin_api_xxx".
func newVCRClient(rec *recorder.Recorder, authValue string) *http.Client {
	transport := rec.GetDefaultClient().Transport
	if authValue != "" {
		transport = &authRoundTripper{
			authValue: authValue,
			transport: transport,
		}
	}

	return &http.Client{Transport: transport}
}

// headerRoundTripper injects a value into an arbitrary request header.
// Used for providers (e.g. Anthropic) that authenticate with a custom
// header instead of Authorization.
type headerRoundTripper struct {
	header    string
	value     string
	transport http.RoundTripper
}

func (rt *headerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if rt.value != "" {
		req.Header.Set(rt.header, rt.value)
	}

	return rt.transport.RoundTrip(req)
}

// newVCRClientWithHeader is like newVCRClient but injects the auth value
// into a named header (e.g. "x-api-key") instead of Authorization, for
// providers that do not use Bearer auth. The header is stripped from the
// cassette by newRecorder's BeforeSave hook.
func newVCRClientWithHeader(rec *recorder.Recorder, header, value string) *http.Client {
	transport := rec.GetDefaultClient().Transport
	if value != "" {
		transport = &headerRoundTripper{
			header:    header,
			value:     value,
			transport: transport,
		}
	}

	return &http.Client{Transport: transport}
}
