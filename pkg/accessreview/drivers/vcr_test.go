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

package drivers

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"

	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/recorder"
)

// versionedClientHeaders are HTTP headers that encode SDK or client library
// versions. They are ignored by the matcher so cassettes keep replaying after
// dependency bumps.
var versionedClientHeaders = []string{"User-Agent", "X-Goog-Api-Client"}

// awsSigningHeaders are written by SigV4 and the AWS SDK on every request.
// They must not be persisted, and they cannot be replayed: dates, signatures
// and invocation IDs change each call.
var awsSigningHeaders = []string{
	"X-Amz-Date",
	"X-Amz-Security-Token",
	"X-Amz-Content-Sha256",
	"Amz-Sdk-Invocation-Id",
	"Amz-Sdk-Request",
}

// newRecorder creates a go-vcr recorder for the given cassette path. When
// the env var is non-empty the recorder runs in record mode, otherwise
// it replays from the committed cassette. A BeforeSave hook strips the
// Authorization header so tokens are never persisted.
//
// Optional sanitizers run as further BeforeSave hooks. A provider whose
// live response carries real member identity must rewrite identity only:
// status, headers and JSON shape are the contract under test. A provider
// may also strip request headers the shared secret hook does not know.
func newRecorder(
	t *testing.T,
	cassettePath string,
	envVar string,
	sanitizers ...func(*cassette.Interaction) error,
) *recorder.Recorder {
	t.Helper()

	return newRecorderWithMatcher(
		t,
		cassettePath,
		envVar,
		cassette.NewDefaultMatcher(
			cassette.WithIgnoreAuthorization(),
			cassette.WithIgnoreHeaders(versionedClientHeaders...),
		),
		sanitizers...,
	)
}

func newRecorderWithMatcher(
	t *testing.T,
	cassettePath string,
	envVar string,
	matcher cassette.MatcherFunc,
	sanitizers ...func(*cassette.Interaction) error,
) *recorder.Recorder {
	t.Helper()

	mode := recorder.ModeReplayOnly
	if os.Getenv(envVar) != "" {
		mode = recorder.ModeRecordOnly
	}

	opts := []recorder.Option{
		recorder.WithMode(mode),
		recorder.WithSkipRequestLatency(true),
		recorder.WithMatcher(matcher),
		recorder.WithHook(stripCassetteSecrets, recorder.BeforeSaveHook),
	}

	for _, sanitize := range sanitizers {
		opts = append(opts, recorder.WithHook(sanitize, recorder.BeforeSaveHook))
	}

	rec, err := recorder.New(cassettePath, opts...)
	if err != nil {
		if mode == recorder.ModeReplayOnly {
			if envVar == "" {
				t.Skipf("cassette not found: %v", err)
			} else {
				t.Skipf("cassette not found (record with %s env var): %v", envVar, err)
			}
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

// stripCassetteSecrets removes request headers that carry credentials so a
// cassette can never be committed with a live secret.
func stripCassetteSecrets(i *cassette.Interaction) error {
	i.Request.Headers.Del("Authorization")
	// Providers like Anthropic (x-api-key), SigNoz (SIGNOZ-API-KEY) and
	// Brevo (api-key) authenticate via a custom header rather than
	// Authorization; strip those too so a re-record never persists a raw key.
	i.Request.Headers.Del("X-Api-Key")
	i.Request.Headers.Del("Signoz-Api-Key")
	i.Request.Headers.Del("Api-Key")
	// Scaleway authenticates with the secret key in X-Auth-Token.
	i.Request.Headers.Del("X-Auth-Token")
	// Dotfile authenticates with the key in X-DOTFILE-API-KEY
	// (canonicalized to X-Dotfile-Api-Key).
	i.Request.Headers.Del("X-Dotfile-Api-Key")

	return nil
}

func sanitizeAWSSigningHeaders(i *cassette.Interaction) error {
	for _, header := range awsSigningHeaders {
		i.Request.Headers.Del(header)
	}

	return nil
}

// newAWSRecorder replays a hand-authored AWS cassette. It never records:
// recording would require live AWS credentials, which these tests refuse to
// read from the environment.
func newAWSRecorder(t *testing.T, cassettePath string) *recorder.Recorder {
	t.Helper()

	return newRecorderWithMatcher(
		t,
		cassettePath,
		"",
		awsAPIMatcher,
		sanitizeAWSSigningHeaders,
	)
}

// awsAPIMatcher matches AWS SDK POSTs without relying on SigV4 headers or
// byte-for-byte bodies. IAM Query is identified by form Action; SSO Admin
// and Identity Store use AWS JSON and are identified by X-Amz-Target;
// Account Management is REST-JSON and is identified by path.
func awsAPIMatcher(r *http.Request, i cassette.Request) bool {
	if r.Method != i.Method {
		return false
	}

	host := r.URL.Host
	if host == "" {
		host = r.Host
	}

	cassetteURL, err := url.Parse(i.URL)
	if err != nil {
		return false
	}

	if host != cassetteURL.Host {
		return false
	}

	target := r.Header.Get("X-Amz-Target")
	if target != "" {
		return target == i.Headers.Get("X-Amz-Target")
	}

	if awsIAMQueryAction(r, i) {
		return true
	}

	return r.URL.Path == cassetteURL.Path && r.URL.Path != "/" && r.URL.Path != ""
}

func awsIAMQueryAction(r *http.Request, i cassette.Request) bool {
	var body []byte

	if r.Body != nil {
		var err error

		body, err = io.ReadAll(r.Body)
		if err != nil {
			return false
		}

		r.Body = io.NopCloser(bytes.NewReader(body))
	}

	values, err := url.ParseQuery(string(body))
	if err != nil {
		return false
	}

	action := values.Get("Action")

	return action != "" && action == i.Form.Get("Action")
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

// basicAuthUserPass returns the HTTP Basic auth header value for a credential
// that already holds the "username:password" pair ("Basic
// base64(<credential>)"), or "" if the credential is empty. ClickHouse
// Cloud (keyId:keySecret) and Langfuse (publicKey:secretKey) present such
// a credential. The matcher ignores Authorization, so this only matters
// when re-recording.
func basicAuthUserPass(credential string) string {
	if credential == "" {
		return ""
	}

	return "Basic " + base64.StdEncoding.EncodeToString([]byte(credential))
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

// roundTripFunc is a test helper that adapts a function to http.RoundTripper
// for stubbing HTTP responses outside VCR cassettes.
type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
