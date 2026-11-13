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

package mdm

import (
	"net/http"
	"os"
	"slices"
	"strconv"
	"testing"

	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/recorder"
)

// replaySafeRequestHeaders is the allowlist of request headers a cassette may
// keep. Anything else, the Authorization header first, is dropped on save.
var replaySafeRequestHeaders = []string{
	"Accept",
	"Accept-Encoding",
	"Content-Type",
}

// replaySafeResponseHeaders is the allowlist of response headers a cassette
// may keep. Provider edges answer with cookies and request IDs that can tie a
// recording back to its tenant.
var replaySafeResponseHeaders = []string{
	"Content-Type",
	"Content-Length",
}

// newRecorder records against the live provider when envVar is set and
// replays the committed cassette otherwise. Sanitizers run before save, after
// the secret-stripping hook, so a recording run still asserts on live data.
func newRecorder(
	t *testing.T,
	cassettePath string,
	envVar string,
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
		recorder.WithMatcher(cassette.NewDefaultMatcher(cassette.WithIgnoreAuthorization(), cassette.WithIgnoreHeaders("User-Agent"))),
		recorder.WithHook(stripCassetteSecrets, recorder.BeforeSaveHook),
	}

	for _, sanitize := range sanitizers {
		opts = append(opts, recorder.WithHook(sanitize, recorder.BeforeSaveHook))
	}

	rec, err := recorder.New(cassettePath, opts...)
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

func stripCassetteSecrets(i *cassette.Interaction) error {
	for name := range i.Request.Headers {
		if !slices.Contains(replaySafeRequestHeaders, http.CanonicalHeaderKey(name)) {
			i.Request.Headers.Del(name)
		}
	}

	for name := range i.Response.Headers {
		if !slices.Contains(replaySafeResponseHeaders, http.CanonicalHeaderKey(name)) {
			i.Response.Headers.Del(name)
		}
	}

	return nil
}

// replaceCassetteBody swaps a recorded response body and keeps its length
// fields in step with it.
func replaceCassetteBody(i *cassette.Interaction, body string) {
	i.Response.Body = body
	i.Response.ContentLength = int64(len(body))

	if _, ok := i.Response.Headers["Content-Length"]; ok {
		i.Response.Headers.Set("Content-Length", strconv.Itoa(len(body)))
	}
}

type bearerRoundTripper struct {
	token     string
	transport http.RoundTripper
}

func (rt *bearerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+rt.token)

	return rt.transport.RoundTrip(req)
}

// newVCRClient returns a client that sends token as a bearer token while
// recording. Replays send none.
func newVCRClient(rec *recorder.Recorder, token string) *http.Client {
	transport := rec.GetDefaultClient().Transport
	if token != "" {
		transport = &bearerRoundTripper{token: token, transport: transport}
	}

	return &http.Client{Transport: transport}
}
