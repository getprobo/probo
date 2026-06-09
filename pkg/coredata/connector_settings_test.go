// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
