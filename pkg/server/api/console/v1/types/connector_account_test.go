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

package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/server/api/console/v1/types"
)

func TestNewDiscoveredConnectorAccounts_Enabled(t *testing.T) {
	t.Parallel()

	stored := map[string]struct{}{
		"111111111111": {},
	}
	got := types.NewDiscoveredConnectorAccounts(
		[]provider.DiscoveredAccount{
			{ExternalAccountID: "111111111111", Name: "stored"},
			{ExternalAccountID: "222222222222", Name: "new"},
		},
		stored,
	)

	require.Len(t, got, 2)
	assert.Equal(t, "111111111111", got[0].ExternalAccountID)
	assert.Equal(t, "stored", got[0].Name)
	assert.True(t, got[0].Enabled)
	assert.Equal(t, "222222222222", got[1].ExternalAccountID)
	assert.Equal(t, "new", got[1].Name)
	assert.False(t, got[1].Enabled)
}
