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

package console_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const regeneratePolicyMutation = `
	mutation RegenerateCookieBannerTrackerPolicy($input: RegenerateCookieBannerTrackerPolicyInput!) {
		regenerateCookieBannerTrackerPolicy(input: $input) {
			cookieBanner { id }
		}
	}
`

func TestRegenerateCookieBannerTrackerPolicy(t *testing.T) {
	t.Parallel()

	t.Run("succeeds for a published banner", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		published := publishBanner(t, owner, bannerID)
		require.Equal(t, "PUBLISHED", published.State)

		var result struct {
			RegenerateCookieBannerTrackerPolicy struct {
				CookieBanner struct {
					ID string `json:"id"`
				} `json:"cookieBanner"`
			} `json:"regenerateCookieBannerTrackerPolicy"`
		}

		err := owner.Execute(regeneratePolicyMutation, map[string]any{
			"input": map[string]any{"cookieBannerId": bannerID},
		}, &result)
		require.NoError(t, err)
		assert.Equal(t, bannerID, result.RegenerateCookieBannerTrackerPolicy.CookieBanner.ID)
	})

	t.Run("conflicts when the banner has no published version", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)

		var result struct{}

		err := owner.Execute(regeneratePolicyMutation, map[string]any{
			"input": map[string]any{"cookieBannerId": bannerID},
		}, &result)
		require.Error(t, err, "regenerating without a published version should fail")
	})
}
