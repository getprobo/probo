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

package console_v1

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/server/api/console/v1/types"
)

func TestComputeCrispVerificationCode(t *testing.T) {
	t.Parallel()

	const (
		secret  = "test-token-secret"
		org     = "gid://organization/1"
		website = "e8592878-c0d0-4632-b2f7-7d882f288d43"
	)

	t.Run("is deterministic for the same inputs", func(t *testing.T) {
		t.Parallel()

		a := computeCrispVerificationCode(secret, org, website)
		b := computeCrispVerificationCode(secret, org, website)
		assert.Equal(t, a, b)
	})

	t.Run("is bound to the organization", func(t *testing.T) {
		t.Parallel()

		a := computeCrispVerificationCode(secret, "gid://organization/1", website)
		b := computeCrispVerificationCode(secret, "gid://organization/2", website)
		assert.NotEqual(t, a, b)
	})

	t.Run("is bound to the website", func(t *testing.T) {
		t.Parallel()

		a := computeCrispVerificationCode(secret, org, "website-a")
		b := computeCrispVerificationCode(secret, org, "website-b")
		assert.NotEqual(t, a, b)
	})

	t.Run("depends on the secret", func(t *testing.T) {
		t.Parallel()

		a := computeCrispVerificationCode("secret-a", org, website)
		b := computeCrispVerificationCode("secret-b", org, website)
		assert.NotEqual(t, a, b)
	})

	t.Run("is 12 human-typeable base32 characters", func(t *testing.T) {
		t.Parallel()

		code := computeCrispVerificationCode(secret, org, website)
		assert.Len(t, code, crispVerificationCodeLength)
		assert.Regexp(t, regexp.MustCompile(`^[A-Z2-7]{12}$`), code)
	})

	// Delimiting org and website prevents a boundary collision where a longer
	// org and shorter website (or vice versa) concatenate to the same bytes.
	t.Run("delimiter prevents org/website boundary collisions", func(t *testing.T) {
		t.Parallel()

		a := computeCrispVerificationCode(secret, "ab", "c")
		b := computeCrispVerificationCode(secret, "a", "bc")
		assert.NotEqual(t, a, b)
	})
}

func TestVerifyCrispVerificationCode(t *testing.T) {
	t.Parallel()

	const (
		secret  = "test-token-secret"
		org     = "gid://organization/1"
		website = "e8592878-c0d0-4632-b2f7-7d882f288d43"
	)

	valid := computeCrispVerificationCode(secret, org, website)
	require.NotEmpty(t, valid)

	t.Run("accepts the exact code", func(t *testing.T) {
		t.Parallel()
		assert.True(t, verifyCrispVerificationCode(secret, org, website, valid))
	})

	t.Run("tolerates surrounding whitespace and lowercase", func(t *testing.T) {
		t.Parallel()
		assert.True(t, verifyCrispVerificationCode(secret, org, website, "  "+valid+"  "))
		assert.True(t, verifyCrispVerificationCode(secret, org, website, strings.ToLower(valid)))
	})

	t.Run("rejects a mismatched or empty code", func(t *testing.T) {
		t.Parallel()
		assert.False(t, verifyCrispVerificationCode(secret, org, website, ""))
		assert.False(t, verifyCrispVerificationCode(secret, org, website, "AAAAAAAAAAAA"))
	})

	t.Run("rejects a code minted for another organization", func(t *testing.T) {
		t.Parallel()

		other := computeCrispVerificationCode(secret, "gid://organization/2", website)
		assert.False(t, verifyCrispVerificationCode(secret, org, website, other))
	})

	t.Run("rejects a code minted for another website", func(t *testing.T) {
		t.Parallel()

		other := computeCrispVerificationCode(secret, org, "another-website")
		assert.False(t, verifyCrispVerificationCode(secret, org, website, other))
	})
}

// The stored Crisp Website ID must be the SAME trimmed value that
// verifyCrispOwnership and the CrispVerificationCode query derive the code
// against, or a padded ID would verify then break the driver's request URL.
func TestApiKeyConnectorSettings_CrispTrimsWebsiteID(t *testing.T) {
	t.Parallel()

	website := "  e8592878-c0d0-4632-b2f7-7d882f288d43  "

	raw, err := apiKeyConnectorSettings(types.CreateAPIKeyConnectorInput{
		Provider:       coredata.ConnectorProviderCrisp,
		CrispWebsiteID: &website,
	})
	require.NoError(t, err)

	var settings coredata.CrispConnectorSettings
	require.NoError(t, json.Unmarshal(raw, &settings))
	assert.Equal(t, "e8592878-c0d0-4632-b2f7-7d882f288d43", settings.WebsiteID)
}

func TestApiKeyConnectorSettings_CrispRejectsWhitespaceOnly(t *testing.T) {
	t.Parallel()

	website := "   "

	_, err := apiKeyConnectorSettings(types.CreateAPIKeyConnectorInput{
		Provider:       coredata.ConnectorProviderCrisp,
		CrispWebsiteID: &website,
	})
	require.Error(t, err)
}

// resolveAPIKeyConnectorCredential is the sole gate keeping normal API-key
// providers requiring a customer key while ManagedAPIKey providers (Model B,
// e.g. Crisp) persist none. A regression either drops the required-key check for
// every provider or persists the managed key on the row.
func TestResolveAPIKeyConnectorCredential(t *testing.T) {
	t.Parallel()

	// NewBuiltinRegistry registers Crisp as a ManagedAPIKey provider; the
	// managed key must then be set for it to count as configured.
	configuredReg := provider.NewBuiltinRegistry()
	configuredReg.SetManagedAPIKey(coredata.ConnectorProviderCrisp, "identifier:secret")
	configured := &Resolver{providerRegistry: configuredReg}

	unconfigured := &Resolver{providerRegistry: provider.NewBuiltinRegistry()}

	clientKey := "customer-key"
	empty := ""

	t.Run("managed and configured persists no key", func(t *testing.T) {
		t.Parallel()

		key, err := configured.resolveAPIKeyConnectorCredential(coredata.ConnectorProviderCrisp, nil)
		require.NoError(t, err)
		assert.Equal(t, "", key)
	})

	t.Run("managed and configured ignores a client-supplied key", func(t *testing.T) {
		t.Parallel()

		key, err := configured.resolveAPIKeyConnectorCredential(coredata.ConnectorProviderCrisp, &clientKey)
		require.NoError(t, err)
		assert.Equal(t, "", key)
	})

	t.Run("managed but unconfigured is rejected", func(t *testing.T) {
		t.Parallel()

		_, err := unconfigured.resolveAPIKeyConnectorCredential(coredata.ConnectorProviderCrisp, nil)
		require.Error(t, err)
	})

	t.Run("non-managed requires a key", func(t *testing.T) {
		t.Parallel()

		_, errNil := configured.resolveAPIKeyConnectorCredential(coredata.ConnectorProviderTally, nil)
		require.EqualError(t, errNil, "apiKey is required")

		_, errEmpty := configured.resolveAPIKeyConnectorCredential(coredata.ConnectorProviderTally, &empty)
		require.EqualError(t, errEmpty, "apiKey is required")
	})

	t.Run("non-managed returns the client key", func(t *testing.T) {
		t.Parallel()

		key, err := configured.resolveAPIKeyConnectorCredential(coredata.ConnectorProviderTally, &clientKey)
		require.NoError(t, err)
		assert.Equal(t, "customer-key", key)
	})
}

// verifyCrispOwnershipWith is the #1b create-time ownership gate. This pins its
// branch wiring: the security polarity (only a matching code passes) and the
// Invalid-versus-Internal error mapping. The settings fetch is faked so no live
// Crisp API is reached.
func TestVerifyCrispOwnershipWith(t *testing.T) {
	t.Parallel()

	const (
		secret   = "test-token-secret"
		pluginID = "plugin-id"
	)

	orgID := gid.New(gid.NewTenantID(), coredata.OrganizationEntityType)
	websiteID := "e8592878-c0d0-4632-b2f7-7d882f288d43"
	validCode := computeCrispVerificationCode(secret, orgID.String(), websiteID)

	// newResolver builds a Resolver whose registry has Crisp configured as a
	// managed provider; the plugin ID is set only when withPluginID is true.
	newResolver := func(withPluginID bool) *Resolver {
		reg := provider.NewBuiltinRegistry()
		reg.SetManagedAPIKey(coredata.ConnectorProviderCrisp, "identifier:secret")

		if withPluginID {
			reg.SetManagedResourceID(coredata.ConnectorProviderCrisp, pluginID)
		}

		return &Resolver{
			providerRegistry: reg,
			tokenSecret:      secret,
			logger:           log.NewLogger(log.WithOutput(io.Discard)),
		}
	}

	newInput := func(website *string) types.CreateAPIKeyConnectorInput {
		return types.CreateAPIKeyConnectorInput{
			OrganizationID: orgID,
			Provider:       coredata.ConnectorProviderCrisp,
			CrispWebsiteID: website,
		}
	}

	fetchCode := func(code string) crispSettingsFetcher {
		return func(context.Context, *http.Client, string, string) (*drivers.CrispSubscriptionSettings, error) {
			return &drivers.CrispSubscriptionSettings{ProboVerificationCode: code}, nil
		}
	}

	assertCode := func(t *testing.T, err error, code string) {
		t.Helper()

		require.Error(t, err)

		gqlErr, ok := err.(*gqlerror.Error)
		require.True(t, ok, "expected *gqlerror.Error, got %T", err)
		assert.Equal(t, code, gqlErr.Extensions["code"])
	}

	t.Run("missing website id is invalid", func(t *testing.T) {
		t.Parallel()

		err := newResolver(true).verifyCrispOwnershipWith(context.Background(), newInput(nil), fetchCode(validCode))
		assertCode(t, err, "INVALID")
	})

	t.Run("whitespace website id is invalid", func(t *testing.T) {
		t.Parallel()

		blank := "   "
		err := newResolver(true).verifyCrispOwnershipWith(context.Background(), newInput(&blank), fetchCode(validCode))
		assertCode(t, err, "INVALID")
	})

	t.Run("missing plugin id is internal", func(t *testing.T) {
		t.Parallel()

		err := newResolver(false).verifyCrispOwnershipWith(context.Background(), newInput(&websiteID), fetchCode(validCode))
		assertCode(t, err, "INTERNAL")
	})

	t.Run("plugin not subscribed is invalid", func(t *testing.T) {
		t.Parallel()

		fetch := func(context.Context, *http.Client, string, string) (*drivers.CrispSubscriptionSettings, error) {
			return nil, drivers.ErrCrispPluginNotSubscribed
		}
		err := newResolver(true).verifyCrispOwnershipWith(context.Background(), newInput(&websiteID), fetch)
		assertCode(t, err, "INVALID")
	})

	t.Run("other fetch error is internal and leaks nothing", func(t *testing.T) {
		t.Parallel()

		fetch := func(context.Context, *http.Client, string, string) (*drivers.CrispSubscriptionSettings, error) {
			return nil, errors.New("boom: plugin token 12345 rejected")
		}
		err := newResolver(true).verifyCrispOwnershipWith(context.Background(), newInput(&websiteID), fetch)
		assertCode(t, err, "INTERNAL")
		assert.NotContains(t, err.Error(), "boom")
		assert.NotContains(t, err.Error(), "12345")
	})

	t.Run("mismatched code is invalid", func(t *testing.T) {
		t.Parallel()

		err := newResolver(true).verifyCrispOwnershipWith(context.Background(), newInput(&websiteID), fetchCode("WRONGCODE000"))
		assertCode(t, err, "INVALID")
	})

	t.Run("matching code passes", func(t *testing.T) {
		t.Parallel()

		err := newResolver(true).verifyCrispOwnershipWith(context.Background(), newInput(&websiteID), fetchCode(validCode))
		require.NoError(t, err)
	})
}
