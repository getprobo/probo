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

package coredata_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/page"
)

func TestRisk_CursorKey(t *testing.T) {
	t.Parallel()

	for _, field := range coredata.RiskOrderFields() {
		t.Run(
			string(field),
			func(t *testing.T) {
				t.Parallel()

				risk := &coredata.Risk{}

				assert.NotPanics(t, func() { risk.CursorKey(field) })
			},
		)
	}

	t.Run(
		"updated at uses the update time",
		func(t *testing.T) {
			t.Parallel()

			updatedAt := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
			risk := &coredata.Risk{UpdatedAt: updatedAt}

			var key page.CursorKey

			require.NotPanics(t, func() { key = risk.CursorKey(coredata.RiskOrderFieldUpdatedAt) })
			assert.Equal(t, updatedAt, key.Value)
		},
	)
}
