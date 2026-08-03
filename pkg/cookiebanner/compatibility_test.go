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

package cookiebanner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemapTextsForPresentation(t *testing.T) {
	t.Parallel()

	baseTexts := func() map[string]string {
		return map[string]string{
			"banner_title":       "We use cookies",
			"banner_description": "This site uses cookies.",
			"button_accept_all":  "Accept All",
			"button_reject_all":  "Reject All",
			"button_customize":   "Customize",
		}
	}

	t.Run("opt-in presentation keeps all buttons", func(t *testing.T) {
		t.Parallel()

		texts := baseTexts()
		remapTextsForPresentation(texts, PresentationOptIn)

		assert.Equal(t, "Accept All", texts["button_accept_all"])
		assert.Equal(t, "Reject All", texts["button_reject_all"])
		assert.Equal(t, "Customize", texts["button_customize"])
	})

	t.Run("opt-out presentation maps opt out to reject and clears customize", func(t *testing.T) {
		t.Parallel()

		texts := baseTexts()
		texts["button_opt_out"] = "Do Not Sell"
		remapTextsForPresentation(texts, PresentationOptOut)

		assert.Equal(t, "Do Not Sell", texts["button_reject_all"])
		assert.Empty(t, texts["button_customize"])
	})
}

func TestSupportsLayout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		version string
		want    bool
	}{
		{"0.10.1", false},
		{"0.10.9", false},
		{"0.11.0", true},
		{"0.12.0", true},
		{"1.0.0", true},
		{"v0.11.0", true},
		{"0.2.0", false},
		{"", true},
		{"invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, supportsLayout(tt.version))
		})
	}
}

func TestSupportsTextRemap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		version string
		want    bool
	}{
		{"0.1.0", false},
		{"0.2.0", false},
		{"0.2.5", false},
		{"0.3.0", true},
		{"1.0.0", true},
		{"", true},
		{"invalid", true},
		{"v0.2.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, supportsTextRemap(tt.version))
		})
	}
}
