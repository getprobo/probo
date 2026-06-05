// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
