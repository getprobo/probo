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

package gcp_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cloudgcp "go.probo.inc/probo/pkg/cloud/gcp"
)

const (
	testProviderResource = "projects/123456789012/locations/global/workloadIdentityPools/probo/providers/probo"
	testServiceAccount   = "probo-audit@my-project.iam.gserviceaccount.com"
)

func TestNewConnectorSettings(t *testing.T) {
	t.Parallel()

	t.Run("stores a canonical provider resource and email", func(t *testing.T) {
		t.Parallel()

		settings, err := cloudgcp.NewConnectorSettings(testProviderResource, testServiceAccount)
		require.NoError(t, err)
		assert.Equal(t, testProviderResource, settings.WorkloadIdentityProvider)
		assert.Equal(t, testServiceAccount, settings.ServiceAccountEmail)
	})

	t.Run("accepts pool and provider ids that start with a digit", func(t *testing.T) {
		t.Parallel()

		resource := "projects/123456789012/locations/global/workloadIdentityPools/1abc/providers/2024p"
		settings, err := cloudgcp.NewConnectorSettings(resource, testServiceAccount)
		require.NoError(t, err)
		assert.Equal(t, resource, settings.WorkloadIdentityProvider)
	})

	t.Run("trims space and strips audience prefixes", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name     string
			resource string
		}{
			{name: "https prefix", resource: "https://iam.googleapis.com/" + testProviderResource},
			{name: "scheme-relative prefix", resource: "//iam.googleapis.com/" + testProviderResource},
			{name: "surrounding space", resource: "  " + testProviderResource + "  "},
			{name: "trailing slash", resource: testProviderResource + "/"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				settings, err := cloudgcp.NewConnectorSettings(tt.resource, "  "+testServiceAccount+"  ")
				require.NoError(t, err)
				assert.Equal(t, testProviderResource, settings.WorkloadIdentityProvider)
				assert.Equal(t, testServiceAccount, settings.ServiceAccountEmail)
			})
		}
	})

	t.Run("refuses a malformed provider resource without echoing it", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name     string
			resource string
		}{
			{name: "empty", resource: ""},
			{name: "project id instead of number", resource: "projects/my-project/locations/global/workloadIdentityPools/probo/providers/probo"},
			{name: "regional location", resource: "projects/123456789012/locations/us-central1/workloadIdentityPools/probo/providers/probo"},
			{name: "missing provider", resource: "projects/123456789012/locations/global/workloadIdentityPools/probo"},
			{name: "uppercase pool", resource: "projects/123456789012/locations/global/workloadIdentityPools/Probo/providers/probo"},
			{name: "short pool id", resource: "projects/123456789012/locations/global/workloadIdentityPools/ab/providers/probo"},
			{name: "pool id ending with hyphen", resource: "projects/123456789012/locations/global/workloadIdentityPools/pool-/providers/probo"},
			{name: "provider id ending with hyphen", resource: "projects/123456789012/locations/global/workloadIdentityPools/probo/providers/probo-"},
			{name: "pool id starting with hyphen", resource: "projects/123456789012/locations/global/workloadIdentityPools/-pool/providers/probo"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				_, err := cloudgcp.NewConnectorSettings(tt.resource, testServiceAccount)
				require.Error(t, err)
				assert.Contains(t, err.Error(), "workloadIdentityProvider is not a workload identity provider resource")

				if tt.resource != "" {
					assert.NotContains(t, err.Error(), tt.resource)
				}
			})
		}
	})

	t.Run("refuses a malformed service account email without echoing it", func(t *testing.T) {
		t.Parallel()

		raw := "alice@example.com"

		_, err := cloudgcp.NewConnectorSettings(testProviderResource, raw)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "serviceAccountEmail is not a service account email")
		assert.NotContains(t, err.Error(), raw)
		assert.NotContains(t, err.Error(), "alice")
	})
}
