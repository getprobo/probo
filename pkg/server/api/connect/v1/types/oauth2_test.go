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

package types

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/iam/oauth2"
)

func TestParseResources(t *testing.T) {
	t.Parallel()

	t.Run(
		"single resource",
		func(t *testing.T) {
			t.Parallel()

			resources, err := parseResources(
				url.Values{"resource": {"https://auth.example.com/api/mcp/v1"}},
			)
			require.NoError(t, err)
			assert.Equal(t, []string{"https://auth.example.com/api/mcp/v1"}, resources)
		},
	)

	t.Run(
		"multiple resources accepted",
		func(t *testing.T) {
			t.Parallel()

			resources, err := parseResources(
				url.Values{
					"resource": {
						"https://auth.example.com/api/mcp/v1",
						"https://other.example.com/api/mcp/v1",
					},
				},
			)
			require.NoError(t, err)
			assert.Equal(
				t,
				[]string{
					"https://auth.example.com/api/mcp/v1",
					"https://other.example.com/api/mcp/v1",
				},
				resources,
			)
		},
	)

	t.Run(
		"empty resource rejected",
		func(t *testing.T) {
			t.Parallel()

			_, err := parseResources(url.Values{"resource": {""}})
			require.ErrorIs(t, err, oauth2.ErrInvalidTarget)
		},
	)
}

func TestOAuth2RegisterInput_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name string
		body string
	}{
		{
			name: "standard scope field",
			body: `{"scope":"openid offline_access v1:org"}`,
		},
		{
			name: "legacy scopes field",
			body: `{"scopes":"openid offline_access v1:org"}`,
		},
	} {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				var input OAuth2RegisterInput
				err := json.Unmarshal([]byte(tt.body), &input)
				require.NoError(t, err)
				assert.Equal(
					t,
					coredata.OAuth2Scopes{
						oauth2.ScopeOpenID,
						oauth2.ScopeOfflineAccess,
						"v1:org",
					},
					input.Scopes,
				)
			},
		)
	}
}

func TestOAuth2RegisterResponse_MarshalJSON(t *testing.T) {
	t.Parallel()

	data, err := json.Marshal(
		OAuth2RegisterResponse{
			Scopes: coredata.OAuth2Scopes{
				oauth2.ScopeOpenID,
				oauth2.ScopeOfflineAccess,
				"v1:org",
			},
		},
	)
	require.NoError(t, err)

	var response map[string]any
	err = json.Unmarshal(data, &response)
	require.NoError(t, err)
	assert.Equal(t, "openid offline_access v1:org", response["scope"])
	assert.NotContains(t, response, "scopes")
}
