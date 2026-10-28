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

package linear

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/connectorprotocol"
)

func TestLinearConnectionGenerationKey(t *testing.T) {
	t.Parallel()

	conn := &connectorprotocol.OAuth2Connection{
		AccessToken:  "access-1",
		RefreshToken: "refresh-1",
		TokenURL:     "https://api.linear.app/oauth/token",
		ClientID:     "client-1",
	}

	key := linearConnectionGenerationKey(conn)
	conn.AccessToken = "access-2"

	assert.Equal(t, key, linearConnectionGenerationKey(conn))
	assert.Equal(
		t,
		linearAssetGeneration(key),
		linearAssetGeneration(linearConnectionGenerationKey(conn)),
	)

	conn.RefreshToken = "refresh-2"
	assert.NotEqual(t, key, linearConnectionGenerationKey(conn))

	withoutRefresh := &connectorprotocol.OAuth2Connection{
		AccessToken: "access-only",
		TokenURL:    "https://api.linear.app/oauth/token",
		ClientID:    "client-1",
	}
	assert.Equal(
		t,
		"https://api.linear.app/oauth/token\x00client-1",
		linearConnectionGenerationKey(withoutRefresh),
	)
	assert.NotContains(t, linearConnectionGenerationKey(withoutRefresh), "access-only")

	assert.Empty(t, linearConnectionGenerationKey(nil))
	assert.Empty(t, linearConnectionGenerationKey(&connectorprotocol.OAuth2Connection{
		AccessToken: "access-only",
	}))
}
