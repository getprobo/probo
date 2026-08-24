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

package authn

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/uri"
)

func TestOAuth2AudiencePolicyAllows(t *testing.T) {
	t.Parallel()

	root := uri.URI("https://app.example.com")
	mcp := uri.URI("https://app.example.com/api/mcp/v1")
	other := uri.URI("https://other.example.com")
	clientID := gid.GID{}

	tests := []struct {
		name   string
		policy OAuth2AudiencePolicy
		token  *coredata.OAuth2AccessToken
		want   bool
	}{
		{
			name:   "manual token remains unbound",
			policy: OAuth2AudiencePolicy{Resource: mcp},
			token:  &coredata.OAuth2AccessToken{Resource: &other},
			want:   true,
		},
		{
			name: "legacy client token allowed explicitly",
			policy: OAuth2AudiencePolicy{
				Resource:     root,
				AllowUnbound: true,
			},
			token: &coredata.OAuth2AccessToken{ClientID: &clientID},
			want:  true,
		},
		{
			name:   "unbound client token rejected by strict resource",
			policy: OAuth2AudiencePolicy{Resource: mcp},
			token:  &coredata.OAuth2AccessToken{ClientID: &clientID},
			want:   false,
		},
		{
			name:   "matching resource allowed",
			policy: OAuth2AudiencePolicy{Resource: mcp},
			token:  &coredata.OAuth2AccessToken{ClientID: &clientID, Resource: &mcp},
			want:   true,
		},
		{
			name:   "unknown resource rejected",
			policy: OAuth2AudiencePolicy{Resource: root},
			token:  &coredata.OAuth2AccessToken{ClientID: &clientID, Resource: &other},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tt.want, tt.policy.Allows(tt.token))
			},
		)
	}
}
