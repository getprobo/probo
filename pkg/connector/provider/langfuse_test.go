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

func TestLangfuseRegistrationMetadata(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderLangfuse)
	require.True(t, ok, "langfuse provider must be registered")

	assert.Equal(t, "Langfuse", reg.DisplayName)
	assert.NotNil(t, reg.APIKey)
	// Langfuse presents publicKey:secretKey as a full HTTP Basic credential.
	assert.Equal(t, provider.APIKeyBasicUserPass, reg.APIKey.Presentation)
	require.Len(t, reg.APIKey.ExtraSettings, 1)
	assert.Equal(t, "baseUrl", reg.APIKey.ExtraSettings[0].Key)
	assert.Equal(t, "Base URL", reg.APIKey.ExtraSettings[0].Label)
	assert.True(t, reg.APIKey.ExtraSettings[0].Required)
	// Single-tenant API-key provider: no picker, no name resolver.
	assert.Nil(t, reg.SetOrganizationSettings, "langfuse must not wire a picker store")
}
