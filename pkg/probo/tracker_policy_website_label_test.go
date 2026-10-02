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

package probo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/coredata"
)

func TestTrackerPolicyWebsiteLabel_UsesBannerNameWhenCorsless(t *testing.T) {
	t.Parallel()

	t.Run(
		"uses origin when the banner is not corsless",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(
				t,
				"https://example.com",
				trackerPolicyWebsiteLabel(
					&coredata.CookieBanner{
						Name:   "Product telemetry",
						Origin: "https://example.com",
					},
				),
			)
		},
	)

	t.Run(
		"uses the banner name when the banner is corsless",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(
				t,
				"Product telemetry",
				trackerPolicyWebsiteLabel(
					&coredata.CookieBanner{
						Name:   "Product telemetry",
						Origin: "https://placeholder.example",
						Capabilities: coredata.CookieBannerCapabilities{
							Corsless: true,
						},
					},
				),
			)
		},
	)
}
