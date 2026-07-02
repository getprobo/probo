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

package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSPFPolicy(t *testing.T) {
	t.Parallel()

	t.Run(
		"detects hard fail",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "fail", parseSPFPolicy("v=spf1 include:_spf.google.com -all"))
		},
	)

	t.Run(
		"detects soft fail",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "softfail", parseSPFPolicy("v=spf1 include:spf.example.com ~all"))
		},
	)

	t.Run(
		"detects neutral",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "neutral", parseSPFPolicy("v=spf1 ?all"))
		},
	)

	t.Run(
		"detects pass all",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "pass", parseSPFPolicy("v=spf1 +all"))
		},
	)

	t.Run(
		"returns empty for no all qualifier",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "", parseSPFPolicy("v=spf1 include:_spf.google.com"))
		},
	)
}
