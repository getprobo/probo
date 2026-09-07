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
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/httpclient"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
)

func TestLangfuseRegistrationMetadata(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderLangfuse)
	require.True(t, ok, "langfuse provider must be registered")

	assert.Equal(t, "Langfuse", reg.DisplayName)
	assert.True(t, reg.SupportsAPIKey())
	// Langfuse presents publicKey:secretKey as a full HTTP Basic credential.
	assert.Equal(
		t,
		provider.APIKeyAuth{Mode: provider.APIKeyAuthBasicUserPass},
		reg.APIKey.Auth,
	)
	require.Len(t, reg.APIKeyExtraSettings(), 1)
	assert.Equal(t, "baseUrl", reg.APIKeyExtraSettings()[0].Key)
	assert.Equal(t, "Base URL", reg.APIKeyExtraSettings()[0].Label)
	assert.True(t, reg.APIKeyExtraSettings()[0].Required)
	// Single-tenant API-key provider: no picker, no name resolver.
	assert.Nil(t, reg.NewNameResolver, "langfuse must not wire a name resolver")
	assert.Nil(t, reg.SetOrganizationSettings, "langfuse must not wire a picker store")
	require.NotNil(t, reg.ClassifyRejection, "langfuse must tell its two 403s apart")
	require.NotNil(t, reg.APIKey.KeyFormat, "langfuse must declare the shape of a pasted key")
	assert.Equal(t, "pk-lf-…:sk-lf-…", reg.APIKey.KeyFormat.Example)
}

func TestLangfuseKeyFormat(t *testing.T) {
	t.Parallel()

	// The check exists to catch a credential that was pasted wrong, which is
	// otherwise indistinguishable from a dead one: the transport base64s the
	// value verbatim, so a missing half authenticates as nothing.
	//
	// What it must NOT do is claim to know the key's scope. Langfuse mints
	// organization and project keys with the same prefixes, so the first two
	// cases below are both well-formed and only one of them can list an
	// organization's members. Only the provider can say which.
	cases := []struct {
		name string
		key  string
		want bool
	}{
		{name: "organization key pair", key: "pk-lf-1111:sk-lf-2222", want: true},
		{name: "project key pair is well formed too", key: "pk-lf-3333:sk-lf-4444", want: true},
		{name: "public key alone", key: "pk-lf-1111"},
		{name: "secret key alone", key: "sk-lf-2222"},
		{name: "halves pasted in the wrong order", key: "sk-lf-2222:pk-lf-1111"},
		{name: "separated by something else", key: "pk-lf-1111 sk-lf-2222"},
		{name: "missing the secret half", key: "pk-lf-1111:"},
		{name: "missing the public half", key: ":sk-lf-2222"},
		{name: "prefixes only", key: "pk-lf-:sk-lf-"},
		{name: "a third colon", key: "pk-lf-1111:sk-lf-2222:extra"},
		{name: "both pairs pasted at once", key: "pk-lf-1111:sk-lf-2222:pk-lf-3333:sk-lf-4444"},
		{name: "some other provider's key", key: "sk-ant-api03-abcdef"},
		{name: "empty", key: ""},
	}

	r := provider.NewBuiltinRegistry()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := r.ValidateAPIKey(coredata.ConnectorProviderLangfuse, tc.key)

			if tc.want {
				require.NoError(t, err)

				return
			}

			require.Error(t, err)
			// The message reaches the customer, so it shows the shape it
			// wanted and never what they pasted.
			assert.Contains(t, err.Error(), "pk-lf-…:sk-lf-…")

			if tc.key != "" {
				assert.NotContains(t, err.Error(), tc.key)
			}
		})
	}
}

// langfuseRoundTripFunc answers a probe with a canned response.
type langfuseRoundTripFunc func(*http.Request) (*http.Response, error)

func (f langfuseRoundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestLangfuseProbeClassifiesRejection(t *testing.T) {
	t.Parallel()

	// The bodies are Langfuse's own, from
	// web/src/pages/api/public/organizations/memberships/index.ts. It checks
	// the key's scope before the organization's plan, and a plan without the
	// admin-api entitlement hides the organization API-keys tab altogether —
	// so a customer who cannot reach the entitlement pastes a project key and
	// gets the scope error, which is a credential they can still fix.
	cases := []struct {
		name         string
		status       int
		body         string
		wantRejected bool
		wantRefused  bool
	}{
		{
			name:         "project key instead of organization key is a bad credential",
			status:       http.StatusForbidden,
			body:         `{"error":"Invalid API key. Organization-scoped API key required for this operation."}`,
			wantRejected: true,
		},
		{
			name:         "plan gate is a refused operation",
			status:       http.StatusForbidden,
			body:         `{"error":"This feature is not available on your current plan."}`,
			wantRejected: true,
			wantRefused:  true,
		},
		{
			name:         "dead key is a bad credential",
			status:       http.StatusUnauthorized,
			body:         `{"error":"Invalid credentials. Confirm that you've configured the correct host."}`,
			wantRejected: true,
		},
		{
			name:   "a listable organization is connected",
			status: http.StatusOK,
			body:   `{"memberships":[]}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := provider.NewBuiltinRegistry()

			raw, err := json.Marshal(&coredata.LangfuseConnectorSettings{
				BaseURL: "https://cloud.langfuse.com",
			})
			require.NoError(t, err)

			client := &http.Client{Transport: langfuseRoundTripFunc(func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, "/api/public/organizations/memberships", req.URL.Path)

				return &http.Response{
					StatusCode: tc.status,
					Body:       io.NopCloser(strings.NewReader(tc.body)),
					Header:     make(http.Header),
				}, nil
			})}

			err = r.ProbeConnection(
				context.Background(),
				client,
				&coredata.Connector{
					Provider:    coredata.ConnectorProviderLangfuse,
					RawSettings: raw,
				},
			)

			if !tc.wantRejected {
				require.NoError(t, err)

				return
			}

			var rejected *provider.CredentialRejectedError

			require.ErrorAs(t, err, &rejected)
			assert.Equal(t, tc.status, rejected.StatusCode)
			assert.Equal(t, tc.wantRefused, rejected.OperationRefused)
		})
	}
}

func TestLangfuseNewDriver(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderLangfuse)
	require.True(t, ok, "langfuse provider must be registered")
	require.NotNil(t, reg.NewDriver, "langfuse NewDriver closure must be wired")

	t.Run("creates driver with valid base_url", func(t *testing.T) {
		t.Parallel()

		raw, err := json.Marshal(&coredata.LangfuseConnectorSettings{
			BaseURL: "https://cloud.langfuse.com",
		})
		require.NoError(t, err)

		conn := &coredata.Connector{
			Provider:    coredata.ConnectorProviderLangfuse,
			RawSettings: raw,
		}

		drv, err := reg.NewDriver(context.Background(), httpclient.DefaultClient(httpclient.WithSSRFProtection()), conn, nil, reg.Endpoints)
		require.NoError(t, err)
		assert.IsType(t, &drivers.LangfuseDriver{}, drv)
	})

	t.Run("errors when base_url is missing", func(t *testing.T) {
		t.Parallel()

		conn := &coredata.Connector{
			Provider:    coredata.ConnectorProviderLangfuse,
			RawSettings: []byte(`{}`),
		}

		_, err := reg.NewDriver(context.Background(), httpclient.DefaultClient(httpclient.WithSSRFProtection()), conn, nil, reg.Endpoints)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "base_url is required")
	})

	t.Run("errors when base_url is invalid", func(t *testing.T) {
		t.Parallel()

		raw, err := json.Marshal(&coredata.LangfuseConnectorSettings{
			BaseURL: "ftp://cloud.langfuse.com",
		})
		require.NoError(t, err)

		conn := &coredata.Connector{
			Provider:    coredata.ConnectorProviderLangfuse,
			RawSettings: raw,
		}

		_, err = reg.NewDriver(context.Background(), httpclient.DefaultClient(httpclient.WithSSRFProtection()), conn, nil, reg.Endpoints)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "base_url must be an http(s) URL")
	})
}
