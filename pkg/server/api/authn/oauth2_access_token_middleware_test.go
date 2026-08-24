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
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/iam/oauth2"
	"go.probo.inc/probo/pkg/uri"
)

func TestOAuth2AccessTokenMatchesIssuer(t *testing.T) {
	t.Parallel()

	root := uri.URI("https://app.example.com")
	mcp := uri.URI("https://app.example.com/api/mcp/v1")
	other := uri.URI("https://other.example.com")
	clientID := gid.GID{}
	svc := &iam.Service{
		OAuth2ServerService: oauth2.NewService(
			nil,
			nil,
			root,
			log.NewLogger(),
		),
	}

	tests := []struct {
		name  string
		token *coredata.OAuth2AccessToken
		want  bool
	}{
		{
			name:  "manual token remains unbound",
			token: &coredata.OAuth2AccessToken{Resources: []uri.URI{other}},
			want:  true,
		},
		{
			name:  "unmigrated client token rejected",
			token: &coredata.OAuth2AccessToken{ClientID: &clientID},
			want:  false,
		},
		{
			name: "matching issuer allowed",
			token: &coredata.OAuth2AccessToken{
				ClientID:  &clientID,
				Resources: []uri.URI{root},
			},
			want: true,
		},
		{
			name: "issuer among audiences allowed",
			token: &coredata.OAuth2AccessToken{
				ClientID:  &clientID,
				Resources: []uri.URI{other, root},
			},
			want: true,
		},
		{
			name: "route resource rejected",
			token: &coredata.OAuth2AccessToken{
				ClientID:  &clientID,
				Resources: []uri.URI{mcp},
			},
			want: false,
		},
		{
			name: "unknown resource rejected",
			token: &coredata.OAuth2AccessToken{
				ClientID:  &clientID,
				Resources: []uri.URI{other},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tt.want, OAuth2AccessTokenMatchesIssuer(svc, tt.token))
			},
		)
	}
}
