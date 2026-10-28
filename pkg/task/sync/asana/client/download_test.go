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

package client

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownloadAllowed_Buckets(t *testing.T) {
	t.Parallel()

	client := &Client{}

	allowed := []string{
		"https://asana-user-private-us-east-1.s3.us-east-1.amazonaws.com/file",
		"https://asana-user-private-us-east-1.s3.amazonaws.com/file",
		"https://asana-user-private-eu-central-1.s3.eu-central-1.amazonaws.com/file",
		"https://asana-user-private-ap-northeast-1.s3.ap-northeast-1.amazonaws.com/file",
		"https://asana-user-private-ap-southeast-2.s3.ap-southeast-2.amazonaws.com/file",
		"https://asana-user-private-me-central-1.s3.me-central-1.amazonaws.com/file",
		"https://asana-export-us-east-1.s3.us-east-1.amazonaws.com/export.json.gz",
		"https://s3.us-east-1.amazonaws.com/asana-user-private-us-east-1/assets/file",
		"https://s3.amazonaws.com/asana-export-us-east-1/export.json.gz",
	}
	for _, rawURL := range allowed {
		assert.True(t, client.downloadAllowed(rawURL), rawURL)
	}

	rejected := []string{
		"https://asana-evil.s3.amazonaws.com/file",
		"https://asana.attacker.s3.amazonaws.com/file",
		"https://notasana.s3.amazonaws.com/file",
		"https://s3.us-east-1.amazonaws.com/asana-evil/file",
		"https://s3.amazonaws.com/asana.attacker/file",
		"https://evil.s3.amazonaws.com/asana-user-private-us-east-1/file",
		"http://asana-user-private-us-east-1.s3.amazonaws.com/file",
	}
	for _, rawURL := range rejected {
		assert.False(t, client.downloadAllowed(rawURL), rawURL)
	}
}

func TestDownloadAllowed_HostOverride(t *testing.T) {
	t.Parallel()

	client := &Client{allowedDownloadHost: "files.test:8443"}

	assert.True(t, client.downloadAllowed("http://files.test:8443/a"))
	assert.True(t, client.downloadAllowed("https://files.test:8443/a"))
	assert.False(t, client.downloadAllowed("https://files.test/a"))
	assert.False(t, client.downloadAllowed("http://evil.test:8443/a"))
	assert.False(t, client.downloadAllowed("https://asana-user-private-us-east-1.s3.amazonaws.com/file"))
	assert.False(t, client.downloadAllowed("ftp://files.test:8443/a"))
}

func TestNewClient_DownloadDoesNotFollowRedirects(t *testing.T) {
	t.Parallel()

	client := NewClient(http.DefaultClient, "https://app.asana.com")
	require.NotNil(t, client.downloadHTTP)
	require.NotNil(t, client.downloadHTTP.CheckRedirect)

	err := client.downloadHTTP.CheckRedirect(nil, nil)
	require.ErrorIs(t, err, http.ErrUseLastResponse)
}

func TestReadDownloadBody_RejectsOversizedBody(t *testing.T) {
	t.Parallel()

	_, err := readDownloadBody(bytes.NewReader([]byte("abcdef")), 5)
	require.Error(t, err)
	assert.ErrorContains(t, err, "exceeds size limit")

	body, err := readDownloadBody(bytes.NewReader([]byte("abcde")), 5)
	require.NoError(t, err)
	assert.Equal(t, []byte("abcde"), body)
}

func TestDownload_FollowsAllowedRedirect(t *testing.T) {
	t.Parallel()

	var hits atomic.Int32

	var srv *httptest.Server

	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits.Add(1) == 1 {
			http.Redirect(w, r, srv.URL+"/file", http.StatusFound)

			return
		}

		w.Header().Set("Content-Type", "text/plain")
		_, _ = io.WriteString(w, "hello")
	}))
	t.Cleanup(srv.Close)

	client := downloadClientForTest(t, srv)

	body, contentType, err := client.Download(t.Context(), srv.URL+"/start")
	require.NoError(t, err)
	assert.Equal(t, []byte("hello"), body)
	assert.Equal(t, "text/plain", contentType)
	assert.Equal(t, int32(2), hits.Load())
}

func TestDownload_RejectsRedirectOffAllowlist(t *testing.T) {
	t.Parallel()

	var evilHits atomic.Int32

	evil := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		evilHits.Add(1)

		_, _ = io.WriteString(w, "secret")
	}))
	t.Cleanup(evil.Close)

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, evil.URL+"/secret", http.StatusFound)
	}))
	t.Cleanup(origin.Close)

	client := downloadClientForTest(t, origin)

	_, _, err := client.Download(t.Context(), origin.URL+"/file")
	require.Error(t, err)
	assert.ErrorContains(t, err, "redirect is not an Asana upload")
	assert.Equal(t, int32(0), evilHits.Load())
}

func TestDownload_RejectsRedirectLoop(t *testing.T) {
	t.Parallel()

	var hits atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Header().Set("Location", "/loop")
		w.WriteHeader(http.StatusFound)
	}))
	t.Cleanup(srv.Close)

	client := downloadClientForTest(t, srv)

	_, _, err := client.Download(t.Context(), srv.URL+"/loop")
	require.Error(t, err)
	assert.ErrorContains(t, err, "too many redirects")
	assert.Equal(t, int32(maxDownloadRedirects+1), hits.Load())
}

func TestDownload_RejectsNotFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "missing", http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	client := downloadClientForTest(t, srv)

	_, _, err := client.Download(t.Context(), srv.URL+"/missing")
	require.Error(t, err)
	assert.ErrorContains(t, err, "status 404")
}

func TestDownload_ResolvesRelativeLocation(t *testing.T) {
	t.Parallel()

	var hits atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits.Add(1) == 1 {
			w.Header().Set("Location", "/file")
			w.WriteHeader(http.StatusFound)

			return
		}

		w.Header().Set("Content-Type", "text/plain")
		_, _ = io.WriteString(w, "relative")
	}))
	t.Cleanup(srv.Close)

	client := downloadClientForTest(t, srv)

	body, contentType, err := client.Download(t.Context(), srv.URL+"/start")
	require.NoError(t, err)
	assert.Equal(t, []byte("relative"), body)
	assert.Equal(t, "text/plain", contentType)
	assert.Equal(t, int32(2), hits.Load())
}

func TestDownload_RejectsRedirectWithoutLocation(t *testing.T) {
	t.Parallel()

	var hits atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusFound)
	}))
	t.Cleanup(srv.Close)

	client := downloadClientForTest(t, srv)

	_, _, err := client.Download(t.Context(), srv.URL+"/start")
	require.Error(t, err)
	assert.ErrorContains(t, err, "redirect has no Location")
	assert.Equal(t, int32(1), hits.Load())
}

func downloadClientForTest(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()

	configured := NewClient(http.DefaultClient, "https://app.asana.com")
	require.NotNil(t, configured.downloadHTTP)
	require.NotNil(t, configured.downloadHTTP.CheckRedirect)

	parsed, err := url.Parse(srv.URL)
	require.NoError(t, err)

	httpClient := srv.Client()
	httpClient.CheckRedirect = configured.downloadHTTP.CheckRedirect

	return &Client{
		downloadHTTP:        httpClient,
		allowedDownloadHost: parsed.Host,
	}
}
