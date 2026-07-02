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

package thirdparty

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestStripCorporateSuffixes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "llc suffix", in: "google llc", want: "google"},
		{name: "comma inc", in: "stripe, inc", want: "stripe"},
		{name: "inc dot", in: "meta inc.", want: "meta"},
		{name: "ltd", in: "deepmind ltd", want: "deepmind"},
		{name: "gmbh", in: "n8n gmbh", want: "n8n"},
		{name: "no suffix", in: "cloudflare", want: "cloudflare"},
		{name: "trailing space", in: "github  inc", want: "github"},
		{name: "only one suffix stripped", in: "foo inc llc", want: "foo inc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, stripCorporateSuffixes(tt.in))
		})
	}
}

func TestRankCandidates(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()

	mkTP := func(name, website string) *coredata.ThirdParty {
		tp := &coredata.ThirdParty{
			ID:   gid.New(tenantID, coredata.ThirdPartyEntityType),
			Name: name,
		}

		if website != "" {
			tp.WebsiteURL = new(website)
		}

		return tp
	}

	t.Run("exact name match scores 1.0", func(t *testing.T) {
		t.Parallel()

		common := coredata.CommonThirdParty{Name: "Google", Slug: "google"}
		got := RankCandidates(common, nil, coredata.ThirdParties{
			mkTP("Google", ""),
			mkTP("Stripe", ""),
		})

		require.Len(t, got, 1)
		assert.Equal(t, 1.0, got[0].Score)
		assert.Equal(t, "Google", got[0].ThirdParty.Name)
	})

	t.Run("suffix-stripped name scores 0.9", func(t *testing.T) {
		t.Parallel()

		common := coredata.CommonThirdParty{Name: "Google", Slug: "google"}
		got := RankCandidates(common, nil, coredata.ThirdParties{
			mkTP("Google LLC", ""),
		})

		require.Len(t, got, 1)
		assert.Equal(t, 0.9, got[0].Score)
	})

	t.Run("slug equality scores 0.85", func(t *testing.T) {
		t.Parallel()

		common := coredata.CommonThirdParty{Name: "Google", Slug: "google"}
		got := RankCandidates(common, nil, coredata.ThirdParties{
			mkTP("google!", ""),
		})

		require.Len(t, got, 1)
		assert.Equal(t, 0.85, got[0].Score)
	})

	t.Run("website host overlap scores 0.8 when name does not match", func(t *testing.T) {
		t.Parallel()

		common := coredata.CommonThirdParty{
			Name:       "Google Analytics",
			Slug:       "google-analytics",
			WebsiteURL: new("https://google.com"),
		}

		got := RankCandidates(common, nil, coredata.ThirdParties{
			mkTP("Sundar's Search Co", "https://www.google.com/about"),
		})

		require.Len(t, got, 1)
		assert.Equal(t, 0.8, got[0].Score)
	})

	t.Run("domain set overlap scores 0.8", func(t *testing.T) {
		t.Parallel()

		common := coredata.CommonThirdParty{Name: "Stripe", Slug: "stripe"}
		domains := coredata.CommonThirdPartyDomains{
			{Domain: "stripe.com"},
			{Domain: "stripe.network"},
		}

		got := RankCandidates(common, domains, coredata.ThirdParties{
			mkTP("Payment Processor", "https://api.stripe.com/v1"),
		})

		require.Len(t, got, 1)
		assert.Equal(t, 0.8, got[0].Score)
	})

	t.Run("no match returns empty", func(t *testing.T) {
		t.Parallel()

		common := coredata.CommonThirdParty{Name: "Stripe", Slug: "stripe"}
		got := RankCandidates(common, nil, coredata.ThirdParties{
			mkTP("Acme", "https://acme.example"),
			mkTP("Widgets Inc", "https://widgets.example"),
		})

		assert.Empty(t, got)
	})

	t.Run("ranks descending by score", func(t *testing.T) {
		t.Parallel()

		common := coredata.CommonThirdParty{
			Name:       "Google",
			Slug:       "google",
			WebsiteURL: new("https://google.com"),
		}

		got := RankCandidates(common, nil, coredata.ThirdParties{
			mkTP("Random", "https://google.com"),
			mkTP("Google", ""),
			mkTP("Google LLC", ""),
		})

		require.Len(t, got, 3)
		assert.Equal(t, "Google", got[0].ThirdParty.Name)
		assert.Equal(t, 1.0, got[0].Score)
		assert.Equal(t, "Google LLC", got[1].ThirdParty.Name)
		assert.Equal(t, 0.9, got[1].Score)
		assert.Equal(t, "Random", got[2].ThirdParty.Name)
		assert.Equal(t, 0.8, got[2].Score)
	})
}
