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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/httpclient"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
)

func TestRenderNewDriver(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderRender)
	require.True(t, ok, "render provider must be registered")

	t.Run("creates driver with valid owner id", func(t *testing.T) {
		t.Parallel()

		raw, err := json.Marshal(&coredata.RenderConnectorSettings{
			OwnerID: "tea-csp8nlbgbbvc73a8nn9g",
		})
		require.NoError(t, err)

		conn := &coredata.Connector{
			Provider:    coredata.ConnectorProviderRender,
			RawSettings: raw,
		}

		drv, err := renderSourceDriver(context.Background(), httpclient.DefaultClient(httpclient.WithSSRFProtection()), conn, nil, reg.Endpoints)
		require.NoError(t, err)
		assert.IsType(t, &RenderDriver{}, drv)
	})

	t.Run("errors when owner id is missing", func(t *testing.T) {
		t.Parallel()

		conn := &coredata.Connector{
			Provider:    coredata.ConnectorProviderRender,
			RawSettings: []byte(`{}`),
		}

		_, err := renderSourceDriver(context.Background(), httpclient.DefaultClient(httpclient.WithSSRFProtection()), conn, nil, reg.Endpoints)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "owner_id is required")
	})
}

func TestRenderNewNameResolver(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderRender)
	require.True(t, ok, "render provider must be registered")

	raw, err := json.Marshal(&coredata.RenderConnectorSettings{
		OwnerID: "tea-csp8nlbgbbvc73a8nn9g",
	})
	require.NoError(t, err)

	conn := &coredata.Connector{
		Provider:    coredata.ConnectorProviderRender,
		RawSettings: raw,
	}

	resolver := renderSourceNameResolver(context.Background(), httpclient.DefaultClient(httpclient.WithSSRFProtection()), conn, nil, reg.Endpoints)
	require.NotNil(t, resolver, "render name resolver must be constructed for a valid connector")
}
