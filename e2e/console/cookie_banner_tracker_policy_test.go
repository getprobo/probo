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
