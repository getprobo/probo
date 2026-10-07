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
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
)

func TestOVHcloudRegistration(t *testing.T) {
	t.Parallel()

	reg, ok := provider.NewBuiltinRegistry().Get(coredata.ConnectorProviderOVHcloud)
	require.True(t, ok)

	assert.Equal(t, "OVHcloud", reg.DisplayName)
	assert.Equal(t, "https://www.probo.com/docs/product/access-review/ovhcloud", reg.DocumentationURL)

	// Both connect paths are offered: OAuth2 for one-click onboarding, client
	// credentials for customers who need to hold (and be able to revoke) the
	// credential themselves.
	require.NotNil(t, reg.OAuth2)
	assert.True(t, reg.SupportsClientCredentials())

	// account/all is the narrowest scope reaching every route the driver reads.
	// The bare "all" spans every product the customer owns.
	assert.Equal(t, []string{"account/all"}, reg.OAuth2.Scopes)
	assert.True(t, reg.OAuth2.RequiresPKCE)
}
