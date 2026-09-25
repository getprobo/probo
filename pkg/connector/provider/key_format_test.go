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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
)

// keyTail stands in for the random part of a key: long, mixed case, and using
// every separator providers put in one. A pattern that rejects it is asserting
// something about the random part, which is the way this check locks out a
// customer holding a perfectly good key.
const keyTail = "AbCdEf0123456789-_.+/=abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func TestAPIKeyFormatInvariants(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()

	for _, reg := range r.All() {
		if reg.APIKey == nil || reg.APIKey.KeyFormat == nil {
			continue
		}

		t.Run(string(reg.Provider), func(t *testing.T) {
			t.Parallel()

			format := reg.APIKey.KeyFormat

			// Register already refuses an example its own pattern rejects.
			// Asserting it through ValidateAPIKey covers the path the console
			// actually takes, where the example is the placeholder.
			assert.NoError(
				t,
				r.ValidateAPIKey(reg.Provider, format.Example),
				"the example shown to the customer must pass its own check",
			)

			assert.Error(
				t,
				r.ValidateAPIKey(reg.Provider, ""),
				"an empty field must not read as a well-formed key",
			)
		})
	}
}

func TestAPIKeyFormatsAcceptAndReject(t *testing.T) {
	t.Parallel()

	// One row per provider that declares a shape. accepted holds a key whose
	// random part is deliberately awkward, because the pattern must assert the
	// prefix and the separator and nothing else. rejected holds the prefix on
	// its own wherever the shape has one, which is a key the customer stopped
	// copying too early, and what the provider's own documentation says cannot
	// work here, so the check earns its keep by naming the problem at paste
	// time instead of returning a generic 401 a campaign later.
	cases := []struct {
		provider coredata.ConnectorProvider
		accepted []string
		rejected []string
	}{
		{
			provider: coredata.ConnectorProviderAnthropic,
			// The Admin API takes an unscoped personal or service account
			// key as well as an admin key, so sk-ant- is the whole of what
			// every credential that can work here shares. The version
			// segment sits outside the prefix either way.
			accepted: []string{"sk-ant-admin01-" + keyTail, "sk-ant-admin99-" + keyTail, "sk-ant-api03-" + keyTail},
			rejected: []string{"sk-ant-", "sk-admin-" + keyTail, keyTail},
		},
		{
			provider: coredata.ConnectorProviderBrevo,
			accepted: []string{"xkeysib-" + keyTail},
			rejected: []string{"xkeysib-", keyTail},
		},
		{
			provider: coredata.ConnectorProviderBrex,
			accepted: []string{"bxt_" + keyTail},
			rejected: []string{"bxt_", keyTail},
		},
		{
			provider: coredata.ConnectorProviderCalCom,
			// Live and test keys part company only after the prefix they
			// share.
			accepted: []string{"cal_live_" + keyTail, "cal_" + keyTail},
			rejected: []string{"cal_", keyTail},
		},
		{
			provider: coredata.ConnectorProviderClickHouse,
			// Neither half carries a prefix, so the separator is all there is
			// to assert; a secret holding a colon of its own still passes.
			accepted: []string{"keyid:" + keyTail, "keyid:secret:with:colons", "keyid:sec\rret"},
			rejected: []string{"keyid", keyTail, ":secret", "keyid:"},
		},
		{
			provider: coredata.ConnectorProviderDotfile,
			accepted: []string{"dotkey." + keyTail + "." + keyTail},
			// The dot is a literal, not a wildcard.
			rejected: []string{"dotkey.", "dotkeyx" + keyTail, keyTail},
		},
		{
			provider: coredata.ConnectorProviderLangfuse,
			// Langfuse mints organization and project keys with the same two
			// prefixes, so both pairs are well formed here and only one of
			// them can list an organization's members. The pattern must not
			// pretend to know which; the probe finds out.
			accepted: []string{"pk-lf-" + keyTail + ":sk-lf-" + keyTail, "pk-lf-1111:sk-lf-2222"},
			rejected: []string{
				"pk-lf-1111",
				"sk-lf-2222",
				"sk-lf-2222:pk-lf-1111",
				"pk-lf-1111 sk-lf-2222",
				"pk-lf-1111:",
				":sk-lf-2222",
				"pk-lf-:sk-lf-",
				"pk-lf-1111:sk-lf-2222:extra",
				"sk-ant-api03-abcdef",
			},
		},
		{
			provider: coredata.ConnectorProviderMetabase,
			accepted: []string{"mb_" + keyTail},
			rejected: []string{"mb_", keyTail},
		},
		{
			provider: coredata.ConnectorProviderOpenAI,
			accepted: []string{"sk-admin-" + keyTail},
			// A project key cannot reach the organization endpoints.
			rejected: []string{"sk-admin-", "sk-proj-" + keyTail, "sk-ant-admin01-" + keyTail},
		},
		{
			provider: coredata.ConnectorProviderQovery,
			// Both organization token classes ride the same Token scheme.
			accepted: []string{"qov_" + keyTail, "sk-qov-01-" + keyTail},
			// The Console's own credential is a JWT sent as Bearer.
			rejected: []string{"qov_", "sk-qov-", "eyJhbGciOiJSUzI1NiJ9.e30.sig", keyTail},
		},
		{
			provider: coredata.ConnectorProviderResend,
			// apiKeyPrefix builds thirteen of these patterns the same way, so
			// this one case pins the helper: what follows the prefix is not
			// the check's business, line terminators included.
			accepted: []string{"re_" + keyTail, "re_\r" + keyTail},
			rejected: []string{"re_", keyTail},
		},
		{
			provider: coredata.ConnectorProviderSendGrid,
			accepted: []string{"SG." + keyTail},
			// The dot is a literal, not a wildcard.
			rejected: []string{"SG.", "SGx" + keyTail, keyTail},
		},
		{
			provider: coredata.ConnectorProviderSupabase,
			accepted: []string{"sbp_" + keyTail},
			// Project keys authenticate one project, not the Management API.
			rejected: []string{"sbp_", "sb_secret_" + keyTail, "sb_publishable_" + keyTail},
		},
		{
			provider: coredata.ConnectorProviderTailscale,
			accepted: []string{"tskey-api-" + keyTail},
			// The other tskey- credentials onboard a device or need an OAuth
			// exchange first.
			rejected: []string{"tskey-api-", "tskey-auth-" + keyTail, "tskey-client-" + keyTail},
		},
		{
			provider: coredata.ConnectorProviderTally,
			accepted: []string{"tly-" + keyTail},
			rejected: []string{"tly-", keyTail},
		},
		{
			provider: coredata.ConnectorProviderUpCloud,
			accepted: []string{"ucat_" + keyTail},
			rejected: []string{"ucat_", keyTail},
		},
	}

	r := provider.NewBuiltinRegistry()

	covered := make(map[coredata.ConnectorProvider]bool, len(cases))
	for _, tc := range cases {
		covered[tc.provider] = true
	}

	// A pattern is a hard block on a paying customer's paste, so the next
	// provider to declare one arrives here rather than shipping unexercised.
	for _, reg := range r.All() {
		if reg.APIKey != nil && reg.APIKey.KeyFormat != nil {
			assert.True(t, covered[reg.Provider], "provider %s declares a key shape with no case here", reg.Provider)
		}
	}

	for _, tc := range cases {
		t.Run(string(tc.provider), func(t *testing.T) {
			t.Parallel()

			// ValidateAPIKey passes anything for a provider that declares no
			// shape, so a stale row here would sail through the accepted loop
			// below rather than fail. Check the row is real first.
			reg, ok := r.Get(tc.provider)
			require.True(t, ok, "provider must be registered")
			require.NotNil(t, reg.APIKey, "provider must offer the API-key path")
			require.NotNil(t, reg.APIKey.KeyFormat, "provider must declare a key shape")

			for _, key := range tc.accepted {
				assert.NoError(t, r.ValidateAPIKey(tc.provider, key), "key %q must be accepted", key)
			}

			for _, key := range tc.rejected {
				err := r.ValidateAPIKey(tc.provider, key)
				if !assert.Error(t, err, "key %q must be rejected", key) {
					continue
				}

				// The rejection has to say what to paste instead, and must
				// never repeat the credential back: it travels to the console
				// as a GraphQL error.
				assert.Contains(t, err.Error(), reg.APIKey.KeyFormat.Example)

				// A rejected key that is part of the example is no secret, and
				// the message names the example on purpose.
				if key != "" && !strings.Contains(reg.APIKey.KeyFormat.Example, key) {
					assert.NotContains(t, err.Error(), key)
				}
			}
		})
	}
}
