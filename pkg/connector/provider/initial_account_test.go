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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
)

func TestRegistration_ResolveInitialAccount(t *testing.T) {
	t.Parallel()

	registry := provider.NewBuiltinRegistry()

	t.Run("aws role arn", func(t *testing.T) {
		t.Parallel()

		reg, ok := registry.Get(coredata.ConnectorProviderAWS)
		require.True(t, ok)

		c := &coredata.Connector{Provider: coredata.ConnectorProviderAWS}
		require.NoError(t, c.SetSettings(coredata.AWSConnectorSettings{
			RoleARN: "arn:aws-us-gov:iam::123456789012:role/ProboAudit",
		}))

		id, name, err := reg.ResolveInitialAccount(c)
		require.NoError(t, err)
		assert.Equal(t, "123456789012", id)
		assert.Equal(t, "123456789012", name)
	})

	t.Run("gcp project from provider", func(t *testing.T) {
		t.Parallel()

		reg, ok := registry.Get(coredata.ConnectorProviderGCP)
		require.True(t, ok)

		c := &coredata.Connector{Provider: coredata.ConnectorProviderGCP}
		require.NoError(t, c.SetSettings(coredata.GCPConnectorSettings{
			WorkloadIdentityProvider: "projects/9876543210/locations/global/workloadIdentityPools/probo/providers/probo",
		}))

		id, name, err := reg.ResolveInitialAccount(c)
		require.NoError(t, err)
		assert.Equal(t, "9876543210", id)
		assert.Equal(t, "9876543210", name)
	})

	t.Run("picker saas empty until org is known", func(t *testing.T) {
		t.Parallel()

		reg, ok := registry.Get(coredata.ConnectorProviderGitHub)
		require.True(t, ok)

		c := &coredata.Connector{Provider: coredata.ConnectorProviderGitHub}
		id, name, err := reg.ResolveInitialAccount(c)
		require.NoError(t, err)
		assert.Equal(t, "", id)
		assert.Equal(t, "", name)
	})

	t.Run("github app uses installation id and org name", func(t *testing.T) {
		t.Parallel()

		reg, ok := registry.Get(coredata.ConnectorProviderGitHub)
		require.True(t, ok)

		c := &coredata.Connector{
			Provider:   coredata.ConnectorProviderGitHub,
			Connection: &connector.GitHubAppConnection{InstallationID: 42},
		}
		require.NoError(t, c.SetSettings(coredata.GitHubConnectorSettings{Organization: "acme"}))

		id, name, err := reg.ResolveInitialAccount(c)
		require.NoError(t, err)
		assert.Equal(t, "42", id)
		assert.Equal(t, "acme", name)
	})

	t.Run("corrupt settings are not an empty tenant", func(t *testing.T) {
		t.Parallel()

		reg, ok := registry.Get(coredata.ConnectorProviderAzure)
		require.True(t, ok)

		c := &coredata.Connector{
			Provider:    coredata.ConnectorProviderAzure,
			RawSettings: []byte(`{"subscription_id":1}`),
		}

		id, name, err := reg.ResolveInitialAccount(c)
		require.Error(t, err)
		assert.ErrorContains(t, err, "cannot read connector settings")
		assert.Equal(t, "", id)
		assert.Equal(t, "", name)
	})

	t.Run("onepassword users api uses account id", func(t *testing.T) {
		t.Parallel()

		reg, ok := registry.Get(coredata.ConnectorProviderOnePassword)
		require.True(t, ok)

		c := &coredata.Connector{Provider: coredata.ConnectorProviderOnePassword}
		require.NoError(t, c.SetSettings(coredata.OnePasswordUsersAPISettings{
			AccountID: "acct_123",
			Region:    "us",
		}))

		id, name, err := reg.ResolveInitialAccount(c)
		require.NoError(t, err)
		assert.Equal(t, "acct_123", id)
		assert.Equal(t, "acct_123", name)
	})

	t.Run("onepassword api key uses scim bridge url", func(t *testing.T) {
		t.Parallel()

		reg, ok := registry.Get(coredata.ConnectorProviderOnePassword)
		require.True(t, ok)

		c := &coredata.Connector{Provider: coredata.ConnectorProviderOnePassword}
		require.NoError(t, c.SetSettings(coredata.OnePasswordConnectorSettings{
			SCIMBridgeURL: "https://scim.example.test",
		}))

		id, name, err := reg.ResolveInitialAccount(c)
		require.NoError(t, err)
		assert.Equal(t, "https://scim.example.test", id)
		assert.Equal(t, "https://scim.example.test", name)
	})

	t.Run("onepassword without settings is empty", func(t *testing.T) {
		t.Parallel()

		reg, ok := registry.Get(coredata.ConnectorProviderOnePassword)
		require.True(t, ok)

		c := &coredata.Connector{Provider: coredata.ConnectorProviderOnePassword}
		id, name, err := reg.ResolveInitialAccount(c)
		require.NoError(t, err)
		assert.Equal(t, "", id)
		assert.Equal(t, "", name)
	})
}
