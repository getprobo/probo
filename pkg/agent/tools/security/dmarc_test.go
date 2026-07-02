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

func TestParseDMARCTag(t *testing.T) {
	t.Parallel()

	t.Run(
		"extracts policy tag",
		func(t *testing.T) {
			t.Parallel()

			record := "v=DMARC1; p=reject; rua=mailto:dmarc@example.com"
			assert.Equal(t, "reject", parseDMARCTag(record, "p"))
		},
	)

	t.Run(
		"extracts rua tag",
		func(t *testing.T) {
			t.Parallel()

			record := "v=DMARC1; p=none; rua=mailto:reports@example.com"
			assert.Equal(t, "mailto:reports@example.com", parseDMARCTag(record, "rua"))
		},
	)

	t.Run(
		"returns empty string for missing tag",
		func(t *testing.T) {
			t.Parallel()

			record := "v=DMARC1; p=quarantine"
			assert.Equal(t, "", parseDMARCTag(record, "ruf"))
		},
	)

	t.Run(
		"extracts pct tag",
		func(t *testing.T) {
			t.Parallel()

			record := "v=DMARC1; p=reject; pct=50; rua=mailto:d@example.com"
			assert.Equal(t, "50", parseDMARCTag(record, "pct"))
		},
	)
}
