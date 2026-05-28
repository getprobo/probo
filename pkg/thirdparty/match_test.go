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

package thirdparty

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

//go:fix inline
func ptr[T any](v T) *T { return new(v) }

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
