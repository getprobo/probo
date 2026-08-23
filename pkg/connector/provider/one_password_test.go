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

// TestOnePasswordRegistrationMetadata pins the per-connect-path settings
// split. 1Password is the only registration offering both paths, and each
// needs different settings because a different driver sits behind each: a
// dialog handed the other path's list would collect fields the create
// resolver rejects, which is exactly how the API-key path was broken while
// the two shapes shared one flat list.
func TestOnePasswordRegistrationMetadata(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderOnePassword)
	require.True(t, ok, "1Password provider must be registered")

	assert.Equal(t, "1Password", reg.DisplayName)
	assert.NotNil(t, reg.APIKey)
	assert.NotNil(t, reg.ClientCredentials)

	require.Len(t, reg.APIKey.ExtraSettings, 1)
	assert.Equal(t, "scimBridgeUrl", reg.APIKey.ExtraSettings[0].Key)
	assert.Equal(t, "SCIM Bridge URL", reg.APIKey.ExtraSettings[0].Label)
	assert.True(t, reg.APIKey.ExtraSettings[0].Required)

	require.Len(t, reg.ClientCredentials.ExtraSettings, 2)
	assert.Equal(t, "accountId", reg.ClientCredentials.ExtraSettings[0].Key)
	assert.Equal(t, "Account ID", reg.ClientCredentials.ExtraSettings[0].Label)
	assert.True(t, reg.ClientCredentials.ExtraSettings[0].Required)
	assert.Equal(t, "region", reg.ClientCredentials.ExtraSettings[1].Key)
	assert.Equal(t, "Region", reg.ClientCredentials.ExtraSettings[1].Label)
	assert.True(t, reg.ClientCredentials.ExtraSettings[1].Required)
}

// TestOnePassword_NewDriver_DispatchByGrantType is the pre-merge gate
// for the 1Password closure. The OnePassword registration dispatches
// between two drivers on GrantType(), which is "" for an API-key
// connection — this test asserts every connection shape reaching the
// closure constructs the driver whose settings the matching
// per-path settings list collects.
