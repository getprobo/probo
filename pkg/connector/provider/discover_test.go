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
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/cloud"
	cloudaws "go.probo.inc/probo/pkg/cloud/aws"
	cloudazure "go.probo.inc/probo/pkg/cloud/azure"
	cloudgcp "go.probo.inc/probo/pkg/cloud/gcp"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
)

// cannedTransport answers every request with the next scripted body,
// regardless of host, so a pager can be driven without a network or a
// recorded cassette.
type cannedTransport struct {
	bodies    []string
	requests  int
	contentTy string
}

func (c *cannedTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if c.requests >= len(c.bodies) {
		return nil, fmt.Errorf("unexpected request %d to %s", c.requests+1, r.URL)
	}

	body := c.bodies[c.requests]
	c.requests++

	contentType := c.contentTy
	if contentType == "" {
		contentType = "application/json"
	}

	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{"Content-Type": []string{contentType}},
		Request:    r,
	}, nil
}

func cannedClient(bodies ...string) (*http.Client, *cannedTransport) {
	transport := &cannedTransport{bodies: bodies}

	return &http.Client{Transport: transport}, transport
}

func awsDiscoverSession(t *testing.T, client *http.Client) *cloudaws.Session {
	t.Helper()

	return cloudaws.NewSessionFromConfig(
		"111111111111",
		cloudaws.CommercialPartition,
		awssdk.Config{
			Region:           cloudaws.DefaultCommercialRegion,
			Credentials:      awssdk.NewCredentialsCache(credentials.NewStaticCredentialsProvider("AKIATESTING", "testing-secret", "")),
			HTTPClient:       client,
			RetryMaxAttempts: 1,
		},
	)
}

// TestDiscoverAccounts covers one paginated walk per cloud, plus the GCP
// scope. Pagination is the property: a first-page-only implementation reviews
// a subset of the organization and looks correct doing it, and no end-to-end
// test can reach a real cloud to notice.
func TestDiscoverAccounts(t *testing.T) {
	t.Parallel()

	registry := provider.NewBuiltinRegistry()

	t.Run("aws walks every page of ListAccounts", func(t *testing.T) {
		t.Parallel()

		reg, ok := registry.Get(coredata.ConnectorProviderAWS)
		require.True(t, ok)

		client, transport := cannedClient(
			`{"Accounts":[{"Id":"111111111111","Name":"Alpha","Status":"ACTIVE"},{"Id":"222222222222","Name":"Suspended","Status":"SUSPENDED"}],"NextToken":"page2"}`,
			`{"Accounts":[{"Id":"333333333333","Name":"Gamma","Status":"ACTIVE"}]}`,
		)

		accounts, err := reg.WorkloadIdentity.DiscoverAccounts(
			context.Background(),
			awsDiscoverSession(t, client),
			&coredata.Connector{Provider: coredata.ConnectorProviderAWS},
		)
		require.NoError(t, err)

		assert.Equal(t, 2, transport.requests, "the second page must be fetched")
		// A suspended account cannot be assumed into, so offering it would
		// only produce a source that fails its first fetch.
		assert.Equal(
			t,
			[]cloud.Account{
				{ID: "111111111111", Name: "Alpha"},
				{ID: "333333333333", Name: "Gamma"},
			},
			accounts,
		)
	})

	t.Run("azure walks every page of the subscriptions pager", func(t *testing.T) {
		t.Parallel()

		reg, ok := registry.Get(coredata.ConnectorProviderAzure)
		require.True(t, ok)

		client, transport := cannedClient(
			`{"value":[{"subscriptionId":"00000000-0000-0000-0000-000000000001","displayName":"Production"}],"nextLink":"https://management.azure.com/subscriptions?page=2"}`,
			`{"value":[{"subscriptionId":"00000000-0000-0000-0000-000000000002","displayName":"Staging"}]}`,
		)

		accounts, err := reg.WorkloadIdentity.DiscoverAccounts(
			context.Background(),
			cloudazure.NewSessionFromToken("", "token", cloudazure.WithHTTPClient(client)),
			&coredata.Connector{Provider: coredata.ConnectorProviderAzure},
		)
		require.NoError(t, err)

		assert.Equal(t, 2, transport.requests, "the second page must be fetched")
		assert.Equal(
			t,
			[]cloud.Account{
				{ID: "00000000-0000-0000-0000-000000000001", Name: "Production"},
				{ID: "00000000-0000-0000-0000-000000000002", Name: "Staging"},
			},
			accounts,
		)
	})

	t.Run("gcp searches under Parent and returns project numbers", func(t *testing.T) {
		t.Parallel()

		reg, ok := registry.Get(coredata.ConnectorProviderGCP)
		require.True(t, ok)

		conn := &coredata.Connector{Provider: coredata.ConnectorProviderGCP}
		require.NoError(t, conn.SetSettings(coredata.GCPConnectorSettings{
			WorkloadIdentityProvider: "projects/42/locations/global/workloadIdentityPools/probo/providers/probo",
			ServiceAccountEmail:      "probo-audit@hub.iam.gserviceaccount.com",
			Parent:                   "organizations/12345",
		}))

		var scopes []string

		client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			scopes = append(scopes, r.URL.Path)

			return &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(strings.NewReader(
					`{"results":[{"name":"//cloudresourcemanager.googleapis.com/projects/200","displayName":"two-folders-deep"}]}`,
				)),
				Header:  http.Header{"Content-Type": []string{"application/json"}},
				Request: r,
			}, nil
		})}

		accounts, err := reg.WorkloadIdentity.DiscoverAccounts(
			context.Background(),
			cloudgcp.NewSessionFromToken("42", "token", cloudgcp.WithHTTPClient(client)),
			conn,
		)
		require.NoError(t, err)

		// Scoping to the project instead of Parent would silently reproduce
		// the project-only blind spot organization install exists to close.
		require.Len(t, scopes, 1)
		assert.Contains(t, scopes[0], "organizations/12345")
		assert.Equal(t, []cloud.Account{{ID: "200", Name: "two-folders-deep"}}, accounts)
	})
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
