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

package coredata_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

// TestConnectorSettings_RoundTrip exercises the ConnectorSettings[T] generic accessor
// against every per-provider settings struct that ships extra
// fields. Each case writes a typed value via SetSettings, reads it
// back via ConnectorSettings[T], and asserts equality. The empty-RawSettings
// case asserts that a connector with no settings yields the zero
// value with no error.
func TestConnectorSettings_RoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("SentryConnectorSettings", func(t *testing.T) {
		t.Parallel()

		want := coredata.SentryConnectorSettings{OrganizationSlug: "acme"}
		c := &coredata.Connector{}
		require.NoError(t, c.SetSettings(&want))

		got, err := coredata.ConnectorSettings[coredata.SentryConnectorSettings](c)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("OnePasswordConnectorSettings", func(t *testing.T) {
		t.Parallel()

		want := coredata.OnePasswordConnectorSettings{SCIMBridgeURL: "https://scim.example.test"}
		c := &coredata.Connector{}
		require.NoError(t, c.SetSettings(&want))

		got, err := coredata.ConnectorSettings[coredata.OnePasswordConnectorSettings](c)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("TallyConnectorSettings", func(t *testing.T) {
		t.Parallel()

		want := coredata.TallyConnectorSettings{OrganizationID: "org_123"}
		c := &coredata.Connector{}
		require.NoError(t, c.SetSettings(&want))

		got, err := coredata.ConnectorSettings[coredata.TallyConnectorSettings](c)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("empty RawSettings returns zero value", func(t *testing.T) {
		t.Parallel()

		c := &coredata.Connector{}
		got, err := coredata.ConnectorSettings[coredata.SentryConnectorSettings](c)
		require.NoError(t, err)
		assert.Equal(t, coredata.SentryConnectorSettings{}, got)
	})

	t.Run("null RawSettings returns zero value", func(t *testing.T) {
		t.Parallel()

		c := &coredata.Connector{RawSettings: []byte("null")}
		got, err := coredata.ConnectorSettings[coredata.TallyConnectorSettings](c)
		require.NoError(t, err)
		assert.Equal(t, coredata.TallyConnectorSettings{}, got)
	})

	t.Run("invalid JSON returns wrapped error", func(t *testing.T) {
		t.Parallel()

		c := &coredata.Connector{RawSettings: []byte("{not-valid")}
		_, err := coredata.ConnectorSettings[coredata.TallyConnectorSettings](c)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot unmarshal connector settings")
	})
}
