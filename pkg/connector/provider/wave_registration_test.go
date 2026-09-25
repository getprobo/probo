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

// TestWaveRegistrationMetadata pins what the Add Source dialog renders for the
// four connectors added together: the auth scheme each key goes out under, and
// the settings fields the dialog asks for beside it. A setting whose Key or
// Required changes here without the console's matching case is silently
// dropped on create, so the pairing is worth asserting per provider rather
// than only in the shared invariant.
func TestWaveRegistrationMetadata(t *testing.T) {
	t.Parallel()

	type extra struct {
		key      string
		required bool
	}

	for _, tt := range []struct {
		name        string
		provider    coredata.ConnectorProvider
		displayName string
		mode        provider.APIKeyAuthMode
		header      string
		extras      []extra
		apiBase     string
		nameResolve bool
	}{
		{
			name:        "elevenlabs",
			provider:    coredata.ConnectorProviderElevenLabs,
			displayName: "ElevenLabs",
			// ElevenLabs does not read Authorization at all.
			mode:   provider.APIKeyAuthHeader,
			header: "xi-api-key",
			// The key names the workspace, so the dialog asks for nothing else.
			extras:  nil,
			apiBase: "https://api.elevenlabs.io/v1",
			// Nothing an API key reaches carries the workspace name.
			nameResolve: false,
		},
		{
			name:        "new relic",
			provider:    coredata.ConnectorProviderNewRelic,
			displayName: "New Relic",
			mode:        provider.APIKeyAuthHeader,
			header:      "API-Key",
			// The region cannot be discovered from the key: the other region's
			// endpoint answers 403.
			extras: []extra{{key: "region", required: true}},
			// Per-region host, so there is no compile-time API base.
			apiBase:     "",
			nameResolve: true,
		},
		{
			name:        "retool",
			provider:    coredata.ConnectorProviderRetool,
			displayName: "Retool",
			mode:        provider.APIKeyAuthBearer,
			header:      "",
			// Optional: a cloud token selects its own organization, so only a
			// self-hosted customer has a URL to give.
			extras:  []extra{{key: "baseUrl", required: false}},
			apiBase: "https://api.retool.com/api/v2",
			// /organization/ returns settings, not a name.
			nameResolve: false,
		},
		{
			name:        "twingate",
			provider:    coredata.ConnectorProviderTwingate,
			displayName: "Twingate",
			mode:        provider.APIKeyAuthHeader,
			header:      "X-API-KEY",
			extras:      []extra{{key: "network", required: true}},
			// Per-network host, so there is no compile-time API base.
			apiBase:     "",
			nameResolve: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reg, ok := provider.NewBuiltinRegistry().Get(tt.provider)
			require.Truef(t, ok, "%s must be registered", tt.name)

			assert.Equal(t, tt.displayName, reg.DisplayName)
			assert.True(t, reg.SupportsAPIKey())
			assert.Equal(t, tt.mode, reg.APIKey.Auth.Mode)
			assert.Equal(t, tt.header, reg.APIKey.Auth.Name)
			assert.Equal(t, tt.apiBase, reg.Endpoints.APIBase)

			// None of the four declares a key shape: only New Relic's keys
			// carry a documented prefix, and the keys it migrated from did not.
			assert.Nil(t, reg.APIKey.KeyFormat)

			// A docs page exists for each, so the dialog renders its link.
			assert.NotEmpty(t, reg.DocumentationURL)

			require.Len(t, reg.APIKeyExtraSettings(), len(tt.extras))

			for i, want := range tt.extras {
				assert.Equal(t, want.key, reg.APIKeyExtraSettings()[i].Key)
				assert.Equal(t, want.required, reg.APIKeyExtraSettings()[i].Required)
				assert.NotEmpty(t, reg.APIKeyExtraSettings()[i].Label)
			}

			assert.NotNil(t, reg.NewDriver)
			assert.Equal(t, tt.nameResolve, reg.NewNameResolver != nil)

			// Every one of the four checks its connection with a closure or a
			// builder rather than a static URL, because each needs something a
			// plain authenticated GET cannot express.
			assert.True(t, reg.Probe != nil || reg.BuildProbeURL != nil)
			assert.Empty(t, reg.Endpoints.Probe)
		})
	}
}
