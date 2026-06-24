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

package coredata

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/connector"
)

func TestConnectorScopeCount(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   *Connector
		want int
	}{
		{
			name: "nil connector",
			in:   nil,
			want: 0,
		},
		{
			name: "nil connection",
			in:   &Connector{},
			want: 0,
		},
		{
			name: "oauth2 empty scope",
			in: &Connector{
				Connection: &connector.OAuth2Connection{Scope: ""},
			},
			want: 0,
		},
		{
			name: "oauth2 single scope",
			in: &Connector{
				Connection: &connector.OAuth2Connection{Scope: "read:user"},
			},
			want: 1,
		},
		{
			name: "oauth2 multiple scopes",
			in: &Connector{
				Connection: &connector.OAuth2Connection{Scope: "read:user write:user admin:org"},
			},
			want: 3,
		},
		{
			name: "oauth2 github comma scopes",
			in: &Connector{
				Connection: &connector.OAuth2Connection{Scope: "repo,gist,user"},
			},
			want: 3,
		},
		{
			name: "slack multi scope",
			in: &Connector{
				Connection: &connector.SlackConnection{
					OAuth2Connection: connector.OAuth2Connection{Scope: "chat:write channels:join incoming-webhook"},
				},
			},
			want: 3,
		},
		{
			name: "unknown connection type",
			in: &Connector{
				Connection: &connector.APIKeyConnection{},
			},
			want: 0,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, c.want, connectorScopeCount(c.in))
		})
	}
}
