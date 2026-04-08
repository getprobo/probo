// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
