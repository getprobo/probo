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
	"slices"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// replaySafeRequestHeaders are the request headers a cassette may keep. It is
// an allowlist, and that direction is the whole point: a credential travels in
// a header whose name only its provider knows, so a list of headers to REMOVE
// is fail-open — the first provider to authenticate through a name nobody
// thought of writes its key into a cassette verbatim, and this repository is
// public. ElevenLabs was that provider: xi-api-key canonicalizes to
// Xi-Api-Key, which is not the X-Api-Key that Anthropic uses, so the older
// removal list did not cover it.
//
// Inverting it makes the failure loud instead. Everything here is either
// content negotiation or a provider's API-version/routing pin, which the
// request matcher compares and which carries no secret. A provider that needs
// a new one gets a test that cannot find its interaction — a failure that
// stops at CI, unlike a leaked key.
var replaySafeRequestHeaders = []string{
	"Accept",
	"Accept-Encoding",
	"Content-Type",
	"User-Agent",
	// Provider API-version and routing pins. They select a response shape, so
	// the matcher must still see them.
	"Anthropic-Version",
	"Intercom-Version",
	"Notion-Version",
	"Square-Version",
	"X-Crisp-Tier",
	"X-Github-Api-Version",
	"X-Goog-Api-Client",
	"X-Amz-Target",
	// Heroku pages with Range/Next-Range, so the request header selects which
	// page comes back and the matcher has to see it.
	"Range",
}

// stripCassetteSecrets drops every request header that is not known to be safe
// to persist, so a cassette cannot be committed with a live secret.
func stripCassetteSecrets(i *cassette.Interaction) error {
	for name := range i.Request.Headers {
		if !slices.Contains(replaySafeRequestHeaders, http.CanonicalHeaderKey(name)) {
			i.Request.Headers.Del(name)
		}
	}

	return nil
}

// replaceCassetteBody swaps a recorded response body and keeps every statement
// of its length in step: the interaction's own field and the Content-Length
// header the provider sent. A sanitizer that sets only the field leaves a
// header describing the length of a body that is no longer there, which is
// harmless to a decoder reading to EOF and misleading to everyone else.
func replaceCassetteBody(i *cassette.Interaction, body string) {
	i.Response.Body = body
	i.Response.ContentLength = int64(len(body))

	if _, ok := i.Response.Headers["Content-Length"]; ok {
		i.Response.Headers.Set("Content-Length", strconv.Itoa(len(body)))
	}
}

// newGCPRecorder replays a hand-authored GCP cassette. It never records:
// recording would require live GCP credentials, which these tests refuse to
// read from the environment.
func newGCPRecorder(t *testing.T, cassettePath string) *recorder.Recorder {
	t.Helper()

	return newRecorderWithMatcher(
		t,
		cassettePath,
		"",
		gcpAPIMatcher,
	)
}

// gcpAPIMatcher matches Google API REST calls by method, host, path, and
// pageToken. Client-default query parameters (alt, prettyPrint, pageSize,
// keyTypes) are ignored so cassettes stay stable across library versions.
func gcpAPIMatcher(r *http.Request, i cassette.Request) bool {
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

	if host != cassetteURL.Host || r.URL.Path != cassetteURL.Path {
		return false
	}

	return r.URL.Query().Get("pageToken") == cassetteURL.Query().Get("pageToken")
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
	)
}

// awsAPIMatcher matches AWS SDK POSTs without relying on SigV4 headers or
// byte-for-byte bodies. IAM Query is identified by form Action; SSO Admin,
// Identity Store, and Organizations use AWS JSON and are identified by
// X-Amz-Target; Account Management is REST-JSON and is identified by path.
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

func TestGCPAPIMatcher(t *testing.T) {
	t.Parallel()

	const (
		serviceAccounts = "https://iam.googleapis.com/v1/projects/123456789012/serviceAccounts"
		serviceKeys     = "https://iam.googleapis.com/v1/projects/-/serviceAccounts/ci@my-project.iam.gserviceaccount.com/keys"
		resourceManager = "https://cloudresourcemanager.googleapis.com/v1/projects/123456789012:getIamPolicy"
	)

	tests := []struct {
		name        string
		method      string
		requestURL  string
		cassetteURL string
		want        bool
	}{
		{
			name:        "same path without page token",
			method:      http.MethodGet,
			requestURL:  serviceAccounts,
			cassetteURL: serviceAccounts,
			want:        true,
		},
		{
			name:        "same path and page token",
			method:      http.MethodGet,
			requestURL:  serviceAccounts + "?pageToken=abc",
			cassetteURL: serviceAccounts + "?pageToken=abc",
			want:        true,
		},
		{
			name:        "same path with different page tokens",
			method:      http.MethodGet,
			requestURL:  serviceAccounts + "?pageToken=abc",
			cassetteURL: serviceAccounts + "?pageToken=def",
			want:        false,
		},
		{
			name:        "page token only on the request",
			method:      http.MethodGet,
			requestURL:  serviceAccounts + "?pageToken=abc",
			cassetteURL: serviceAccounts,
			want:        false,
		},
		{
			name:        "page token only on the cassette",
			method:      http.MethodGet,
			requestURL:  serviceAccounts,
			cassetteURL: serviceAccounts + "?pageToken=abc",
			want:        false,
		},
		{
			name:        "ignores client-default query parameters",
			method:      http.MethodGet,
			requestURL:  serviceAccounts + "?alt=json&prettyPrint=false&pageSize=100&keyTypes=USER_MANAGED",
			cassetteURL: serviceAccounts,
			want:        true,
		},
		{
			name:        "matches page token while ignoring client defaults",
			method:      http.MethodGet,
			requestURL:  serviceAccounts + "?alt=json&pageSize=100&pageToken=abc",
			cassetteURL: serviceAccounts + "?pageToken=abc",
			want:        true,
		},
		{
			name:        "different path",
			method:      http.MethodGet,
			requestURL:  serviceAccounts,
			cassetteURL: serviceKeys,
			want:        false,
		},
		{
			name:        "different host",
			method:      http.MethodPost,
			requestURL:  resourceManager,
			cassetteURL: serviceAccounts,
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				req, err := http.NewRequest(tt.method, tt.requestURL, nil)
				require.NoError(t, err)

				got := gcpAPIMatcher(
					req,
					cassette.Request{Method: tt.method, URL: tt.cassetteURL},
				)
				assert.Equal(t, tt.want, got)
			},
		)
	}
}
