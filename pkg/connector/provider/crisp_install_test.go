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

package provider_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
)

const (
	crispInstallWebsiteID = "e8592878-c0d0-4632-b2f7-7d882f288d43"
	crispInstallPluginID  = "e979a1c3-2c41-4e93-a8ed-410ace27318e"
	crispInstallToken     = "s3cr3t-subscription-token"
)

// installHostRewriter points the driver's pinned Crisp host at a test server.
// GetCrispSubscription reaches crispDefaultBaseURL directly — that is why Crisp
// declares EndpointOverrideUnsupported — so there is no Endpoints field to move
// and the transport is the only seam.
type installHostRewriter struct {
	target string
}

func (h *installHostRewriter) RoundTrip(r *http.Request) (*http.Response, error) {
	u, err := url.Parse(h.target)
	if err != nil {
		return nil, err
	}

	r2 := r.Clone(r.Context())
	r2.URL.Scheme = u.Scheme
	r2.URL.Host = u.Host

	return http.DefaultTransport.RoundTrip(r2)
}

func crispRegistrationForTest(t *testing.T) *provider.Registration {
	t.Helper()

	reg, ok := provider.NewBuiltinRegistry().Get(coredata.ConnectorProviderCrisp)
	require.True(t, ok, "crisp is not registered")
	require.NotNil(t, reg.Install)
	require.NotNil(t, reg.Install.Verify)

	return reg
}

// crispSubscriptionServer answers the plugin-subscription endpoint and records
// the website id Verify actually put in the path.
func crispSubscriptionServer(t *testing.T, seenPath *string) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if seenPath != nil {
			*seenPath = r.URL.Path
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(
			`{"error":false,"reason":"resolved","data":{"plugin_id":"` +
				crispInstallPluginID + `","token":"` + crispInstallToken + `"}}`,
		))
	}))
	t.Cleanup(srv.Close)

	return srv
}

// TestCrispInstallVerify_CanonicalizesWebsiteID pins the uuid.Parse round-trip.
// Crisp's own dashboard and every hand-typed link agree on the canonical
// spelling, but uuid.Parse accepts five and returns an equal value for all of
// them. Persisting whichever one arrived would give one website five distinct
// settings ->> 'website_id' values, five advisory-lock keys, and five
// connectors — the idempotency the whole ceremony rests on, defeated by
// punctuation.
func TestCrispInstallVerify_CanonicalizesWebsiteID(t *testing.T) {
	t.Parallel()

	spellings := map[string]string{
		"canonical": crispInstallWebsiteID,
		"uppercase": "E8592878-C0D0-4632-B2F7-7D882F288D43",
		"braced":    "{e8592878-c0d0-4632-b2f7-7d882f288d43}",
		"urn":       "urn:uuid:e8592878-c0d0-4632-b2f7-7d882f288d43",
		"bare hex":  "e8592878c0d04632b2f77d882f288d43",
	}

	for name, spelling := range spellings {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			reg := crispRegistrationForTest(t)
			seenPath := ""
			srv := crispSubscriptionServer(t, &seenPath)
			client := &http.Client{Transport: &installHostRewriter{target: srv.URL}}

			resourceID, err := reg.Install.Verify(
				t.Context(),
				client,
				crispInstallPluginID,
				url.Values{
					"website_id": {spelling},
					"token":      {crispInstallToken},
				},
			)
			require.NoError(t, err)

			// Both the persisted id AND the id the vendor call used must be
			// canonical: a canonical return over a non-canonical lookup would
			// verify a different website than the one it records.
			assert.Equal(t, crispInstallWebsiteID, resourceID)
			assert.Contains(t, seenPath, crispInstallWebsiteID)
		})
	}
}

// TestCrispInstallVerify_Terminal covers every refusal that must BURN the
// customer's single-use state: retrying cannot turn any of them into a success,
// and releasing would hand a forged proof an unlimited number of attempts.
func TestCrispInstallVerify_Terminal(t *testing.T) {
	t.Parallel()

	t.Run("malformed website id", func(t *testing.T) {
		t.Parallel()

		reg := crispRegistrationForTest(t)

		_, err := reg.Install.Verify(
			t.Context(),
			&http.Client{},
			crispInstallPluginID,
			url.Values{"website_id": {"not-a-uuid"}, "token": {crispInstallToken}},
		)
		require.Error(t, err)
		assert.NotErrorIs(t, err, provider.ErrInstallVerificationTransient)
	})

	t.Run("missing proof", func(t *testing.T) {
		t.Parallel()

		reg := crispRegistrationForTest(t)

		_, err := reg.Install.Verify(
			t.Context(),
			&http.Client{},
			crispInstallPluginID,
			url.Values{"website_id": {crispInstallWebsiteID}},
		)
		require.Error(t, err)
		assert.NotErrorIs(t, err, provider.ErrInstallVerificationTransient)
	})

	t.Run("proof does not match the subscription", func(t *testing.T) {
		t.Parallel()

		reg := crispRegistrationForTest(t)
		srv := crispSubscriptionServer(t, nil)
		client := &http.Client{Transport: &installHostRewriter{target: srv.URL}}

		_, err := reg.Install.Verify(
			t.Context(),
			client,
			crispInstallPluginID,
			url.Values{
				"website_id": {crispInstallWebsiteID},
				"token":      {"forged-token"},
			},
		)
		require.Error(t, err)
		assert.NotErrorIs(t, err, provider.ErrInstallVerificationTransient)
	})

	// 401/403 mean Probo's own plugin credential is wrong, revoked or
	// cross-wired with the plugin id. A retry makes the identical request.
	for name, status := range map[string]int{
		"401 on Probo's own credential": http.StatusUnauthorized,
		"403 on Probo's own credential": http.StatusForbidden,
		"400 bad request":               http.StatusBadRequest,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			reg := crispRegistrationForTest(t)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(status)
			}))
			t.Cleanup(srv.Close)

			client := &http.Client{Transport: &installHostRewriter{target: srv.URL}}

			_, err := reg.Install.Verify(
				t.Context(),
				client,
				crispInstallPluginID,
				url.Values{
					"website_id": {crispInstallWebsiteID},
					"token":      {crispInstallToken},
				},
			)
			require.Error(t, err)
			assert.NotErrorIs(t, err, provider.ErrInstallVerificationTransient)
		})
	}

	// 404 is Crisp saying Probo's plugin is not subscribed to this website. No
	// retry installs it for them.
	t.Run("404 plugin not subscribed", func(t *testing.T) {
		t.Parallel()

		reg := crispRegistrationForTest(t)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		t.Cleanup(srv.Close)

		client := &http.Client{Transport: &installHostRewriter{target: srv.URL}}

		_, err := reg.Install.Verify(
			t.Context(),
			client,
			crispInstallPluginID,
			url.Values{
				"website_id": {crispInstallWebsiteID},
				"token":      {crispInstallToken},
			},
		)
		require.ErrorIs(t, err, drivers.ErrCrispPluginNotSubscribed)
		assert.NotErrorIs(t, err, provider.ErrInstallVerificationTransient)
	})
}

// TestCrispInstallVerify_Transient covers the failures that must RELEASE the
// state, so a vendor blip does not cost the customer the rest of their
// ten-minute window.
func TestCrispInstallVerify_Transient(t *testing.T) {
	t.Parallel()

	for name, status := range map[string]int{
		"429 rate limited":   http.StatusTooManyRequests,
		"500 vendor fault":   http.StatusInternalServerError,
		"502 bad gateway":    http.StatusBadGateway,
		"503 unavailable":    http.StatusServiceUnavailable,
		"504 vendor timeout": http.StatusGatewayTimeout,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			reg := crispRegistrationForTest(t)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(status)
			}))
			t.Cleanup(srv.Close)

			client := &http.Client{Transport: &installHostRewriter{target: srv.URL}}

			_, err := reg.Install.Verify(
				t.Context(),
				client,
				crispInstallPluginID,
				url.Values{
					"website_id": {crispInstallWebsiteID},
					"token":      {crispInstallToken},
				},
			)
			require.ErrorIs(t, err, provider.ErrInstallVerificationTransient)
		})
	}

	// No status reached us at all: a dial, TLS or timeout failure. Releasing is
	// the safe direction — the state still expires on its own.
	t.Run("transport failure", func(t *testing.T) {
		t.Parallel()

		reg := crispRegistrationForTest(t)
		srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		target := srv.URL

		srv.Close() // nothing is listening now

		client := &http.Client{Transport: &installHostRewriter{target: target}}

		_, err := reg.Install.Verify(
			t.Context(),
			client,
			crispInstallPluginID,
			url.Values{
				"website_id": {crispInstallWebsiteID},
				"token":      {crispInstallToken},
			},
		)
		require.ErrorIs(t, err, provider.ErrInstallVerificationTransient)
	})
}
