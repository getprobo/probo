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

func TestLangfuseNewDriver(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderLangfuse)
	require.True(t, ok, "langfuse provider must be registered")

	t.Run("creates driver with valid base_url", func(t *testing.T) {
		t.Parallel()

		raw, err := json.Marshal(&coredata.LangfuseConnectorSettings{
			BaseURL: "https://cloud.langfuse.com",
		})
		require.NoError(t, err)

		conn := &coredata.Connector{
			Provider:    coredata.ConnectorProviderLangfuse,
			RawSettings: raw,
		}

		drv, err := langfuseSourceDriver(context.Background(), httpclient.DefaultClient(httpclient.WithSSRFProtection()), conn, nil, reg.Endpoints)
		require.NoError(t, err)
		assert.IsType(t, &LangfuseDriver{}, drv)
	})

	t.Run("errors when base_url is missing", func(t *testing.T) {
		t.Parallel()

		conn := &coredata.Connector{
			Provider:    coredata.ConnectorProviderLangfuse,
			RawSettings: []byte(`{}`),
		}

		_, err := langfuseSourceDriver(context.Background(), httpclient.DefaultClient(httpclient.WithSSRFProtection()), conn, nil, reg.Endpoints)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "base_url is required")
	})

	t.Run("errors when base_url is invalid", func(t *testing.T) {
		t.Parallel()

		raw, err := json.Marshal(&coredata.LangfuseConnectorSettings{
			BaseURL: "ftp://cloud.langfuse.com",
		})
		require.NoError(t, err)

		conn := &coredata.Connector{
			Provider:    coredata.ConnectorProviderLangfuse,
			RawSettings: raw,
		}

		_, err = langfuseSourceDriver(context.Background(), httpclient.DefaultClient(httpclient.WithSSRFProtection()), conn, nil, reg.Endpoints)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "base_url must be an http(s) URL")
	})
}
