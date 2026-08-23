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
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
)

// Every connector provider must have a driver here. A provider the catalog can
// connect but this registry cannot review would be offered in the access-source
// picker and then fail its first campaign fetch.
//
// A provider that genuinely reviews nothing should be removed from the
// ConnectorProvider constants rather than left absent here.
func TestEveryProviderReviewsAccounts(t *testing.T) {
	t.Parallel()

	registry := NewBuiltinRegistry()

	for _, p := range coredata.ConnectorProviders() {
		t.Run(string(p), func(t *testing.T) {
			t.Parallel()

			assert.Truef(t, registry.Supports(p), "provider %q registers no access review driver", p)
		})
	}
}

func TestRegistryRegister(t *testing.T) {
	t.Parallel()

	factory := Factory(func(context.Context, *provider.Handle, *log.Logger) (Driver, error) {
		return nil, nil
	})

	t.Run("rejects a missing provider", func(t *testing.T) {
		t.Parallel()

		require.Error(t, NewRegistry().Register("", factory))
	})

	t.Run("rejects a nil factory", func(t *testing.T) {
		t.Parallel()

		require.Error(t, NewRegistry().Register(coredata.ConnectorProviderSlack, nil))
	})

	t.Run("rejects a duplicate", func(t *testing.T) {
		t.Parallel()

		r := NewRegistry()
		require.NoError(t, r.Register(coredata.ConnectorProviderSlack, factory))
		require.Error(t, r.Register(coredata.ConnectorProviderSlack, factory))
	})
}

// A connector Probo can open but not review must fail by name, so the campaign
// error says which provider rather than reporting a nil driver later.
func TestRegistryNewUnsupportedProvider(t *testing.T) {
	t.Parallel()

	reg := &provider.Registration{
		Provider:    coredata.ConnectorProviderSlack,
		DisplayName: "Slack",
	}
	conn := &coredata.Connector{Provider: coredata.ConnectorProviderSlack}

	driver, err := NewRegistry().New(
		context.Background(),
		provider.NewHTTPHandleForTest(reg, conn, http.DefaultClient),
		log.NewLogger(),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reviews no accounts")
	assert.Nil(t, driver)
}

// A capability a provider does not have must read as absent, not as a driver
// that answers and panics: capable drops a nil optional rather than embedding
// it, so the assertion behind Organizations and InstanceName tells the truth.
func TestCapableOmitsAbsentCapabilities(t *testing.T) {
	t.Parallel()

	bare := Driver(&stubDriver{})

	t.Run("no optional capability", func(t *testing.T) {
		t.Parallel()

		driver := capable(bare, nil, nil)

		_, isResolver := driver.(NameResolver)
		assert.False(t, isResolver)

		_, isLister := driver.(OrganizationLister)
		assert.False(t, isLister)

		orgs, err := Organizations(context.Background(), driver)
		require.NoError(t, err)
		assert.Empty(t, orgs)

		name, err := InstanceName(context.Background(), driver)
		require.NoError(t, err)
		assert.Empty(t, name)
	})

	t.Run("name only", func(t *testing.T) {
		t.Parallel()

		driver := capable(bare, stubNameResolver("acme"), nil)

		name, err := InstanceName(context.Background(), driver)
		require.NoError(t, err)
		assert.Equal(t, "acme", name)

		_, isLister := driver.(OrganizationLister)
		assert.False(t, isLister)
	})

	t.Run("organizations only", func(t *testing.T) {
		t.Parallel()

		driver := capable(bare, nil, stubOrganizationLister("acme"))

		orgs, err := Organizations(context.Background(), driver)
		require.NoError(t, err)
		assert.Equal(t, []Organization{{Slug: "acme"}}, orgs)

		_, isResolver := driver.(NameResolver)
		assert.False(t, isResolver)
	})

	t.Run("both", func(t *testing.T) {
		t.Parallel()

		driver := capable(bare, stubNameResolver("acme"), stubOrganizationLister("acme"))

		name, err := InstanceName(context.Background(), driver)
		require.NoError(t, err)
		assert.Equal(t, "acme", name)

		orgs, err := Organizations(context.Background(), driver)
		require.NoError(t, err)
		assert.Len(t, orgs, 1)
	})
}

type stubDriver struct{}

func (*stubDriver) ListAccounts(context.Context) ([]AccountRecord, error) {
	return nil, nil
}

type stubNameResolver string

func (s stubNameResolver) ResolveInstanceName(context.Context) (string, error) {
	return string(s), nil
}

func stubOrganizationLister(slug string) OrganizationLister {
	return organizationListerFunc(func(context.Context) ([]Organization, error) {
		return []Organization{{Slug: slug}}, nil
	})
}
