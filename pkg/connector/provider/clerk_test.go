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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/httpclient"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
)

func TestClerkRegistration(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderClerk)
	require.True(t, ok, "clerk provider must be registered")

	assert.Equal(t, "Clerk", reg.DisplayName)
	assert.True(t, reg.SupportsAPIKey)
	assert.Equal(t, "", reg.APIKeyHeader)
	assert.False(t, reg.APIKeyBasicAuth)
	require.NotNil(t, reg.NewDriver, "clerk NewDriver closure must be wired")

	drv, err := reg.NewDriver(
		context.Background(),
		httpclient.DefaultClient(httpclient.WithSSRFProtection()),
		&coredata.Connector{Provider: coredata.ConnectorProviderClerk},
		nil,
	)
	require.NoError(t, err)
	assert.IsType(t, &drivers.ClerkDriver{}, drv)
}
