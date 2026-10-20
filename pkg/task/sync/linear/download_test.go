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

package linear

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownloadAsset_RejectsForeignHost(t *testing.T) {
	t.Parallel()

	client := NewClient(http.DefaultClient, "https://api.linear.app")

	_, _, err := client.DownloadAsset(t.Context(), "https://example.com/diagram.png")
	require.Error(t, err)
	assert.ErrorContains(t, err, "not a Linear upload")
}

func TestDownloadAsset_ReadsBody(t *testing.T) {
	t.Parallel()

	var (
		mu   sync.Mutex
		auth string
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		auth = r.Header.Get("Authorization")
		mu.Unlock()

		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte("png-bytes"))
	}))
	t.Cleanup(srv.Close)

	transport := srv.Client().Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	client := NewClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			req = req.Clone(req.Context())
			req.Header.Set("Authorization", "Bearer oauth-token")

			return transport.RoundTrip(req)
		}),
	}, srv.URL)

	assetURL := srv.URL + "/diagram.png"
	client.allowedAssetHost = mustHostname(t, assetURL)

	body, contentType, err := client.DownloadAsset(t.Context(), assetURL)
	require.NoError(t, err)
	assert.Equal(t, []byte("png-bytes"), body)
	assert.Equal(t, "image/png", contentType)

	mu.Lock()
	gotAuth := auth
	mu.Unlock()

	assert.Equal(t, "Bearer oauth-token", gotAuth)
}

func TestDownloadAsset_RedirectOmitsToken(t *testing.T) {
	t.Parallel()

	var (
		mu         sync.Mutex
		originAuth string
		cdnAuth    string
		cdnBody    = []byte("png-bytes")
	)

	cdn := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		cdnAuth = r.Header.Get("Authorization")
		mu.Unlock()

		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(cdnBody)
	}))
	t.Cleanup(cdn.Close)

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		originAuth = r.Header.Get("Authorization")
		mu.Unlock()

		http.Redirect(w, r, cdn.URL+"/file.png", http.StatusFound)
	}))
	t.Cleanup(origin.Close)

	originTransport := origin.Client().Transport
	if originTransport == nil {
		originTransport = http.DefaultTransport
	}

	client := NewClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			req = req.Clone(req.Context())
			req.Header.Set("Authorization", "Bearer oauth-token")

			return originTransport.RoundTrip(req)
		}),
	}, origin.URL)
	client.allowedAssetHost = mustHostname(t, origin.URL)
	client.uploadHTTP = cdn.Client()

	body, contentType, err := client.DownloadAsset(t.Context(), origin.URL+"/diagram.png")
	require.NoError(t, err)
	assert.Equal(t, cdnBody, body)
	assert.Equal(t, "image/png", contentType)

	mu.Lock()
	gotOriginAuth, gotCDNAuth := originAuth, cdnAuth
	mu.Unlock()

	assert.Equal(t, "Bearer oauth-token", gotOriginAuth)
	assert.Empty(t, gotCDNAuth)
}

func TestDownloadAsset_FollowsThreeRedirects(t *testing.T) {
	t.Parallel()

	var (
		mu      sync.Mutex
		cdnHops int
		cdn     *httptest.Server
	)

	cdn = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		cdnHops++
		hop := cdnHops
		mu.Unlock()

		if hop < 3 {
			http.Redirect(w, r, cdn.URL+"/file.png", http.StatusFound)

			return
		}

		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte("png-bytes"))
	}))
	t.Cleanup(cdn.Close)

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, cdn.URL+"/file.png", http.StatusFound)
	}))
	t.Cleanup(origin.Close)

	originTransport := origin.Client().Transport
	if originTransport == nil {
		originTransport = http.DefaultTransport
	}

	client := NewClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			req = req.Clone(req.Context())
			req.Header.Set("Authorization", "Bearer oauth-token")

			return originTransport.RoundTrip(req)
		}),
	}, origin.URL)
	client.allowedAssetHost = mustHostname(t, origin.URL)
	client.uploadHTTP = cdn.Client()

	body, _, err := client.DownloadAsset(t.Context(), origin.URL+"/diagram.png")
	require.NoError(t, err)
	assert.Equal(t, []byte("png-bytes"), body)
}

func TestDownloadAsset_EmptyBodyKeepsStatus(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	client := NewClient(srv.Client(), srv.URL)
	client.allowedAssetHost = mustHostname(t, srv.URL)

	_, _, err := client.DownloadAsset(t.Context(), srv.URL+"/empty.png")
	require.Error(t, err)

	var rejected *AssetRejectedError
	require.ErrorAs(t, err, &rejected)
	assert.Equal(t, http.StatusOK, rejected.Status)
}

func TestDownloadAsset_NotFoundIsUnavailable(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	client := NewClient(srv.Client(), srv.URL)
	client.allowedAssetHost = mustHostname(t, srv.URL)

	_, _, err := client.DownloadAsset(t.Context(), srv.URL+"/missing.png")
	require.Error(t, err)

	var unavailable *AssetUnavailableError
	require.ErrorAs(t, err, &unavailable)
	assert.Equal(t, http.StatusNotFound, unavailable.Status)
}

func mustHostname(t *testing.T, raw string) string {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, raw, nil)
	require.NoError(t, err)

	return req.URL.Hostname()
}

func TestDownloadAsset_RejectsOversizedBody(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.Copy(w, io.LimitReader(neverEnding('x'), maxLinearUploadBytes+1))
	}))
	t.Cleanup(srv.Close)

	client := NewClient(srv.Client(), srv.URL)
	client.allowedAssetHost = mustHostname(t, srv.URL)

	_, _, err := client.DownloadAsset(t.Context(), srv.URL+"/huge.png")
	require.Error(t, err)

	var rejected *AssetRejectedError
	require.ErrorAs(t, err, &rejected)
	assert.Equal(t, http.StatusRequestEntityTooLarge, rejected.Status)
}

type neverEnding byte

func (b neverEnding) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = byte(b)
	}

	return len(p), nil
}
