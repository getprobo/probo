// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

package connector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.probo.inc/probo/pkg/gid"
)

type (
	ProtocolType string

	Connector interface {
		Initiate(ctx context.Context, provider string, organizationID gid.GID, r *http.Request) (string, error)
		Complete(ctx context.Context, r *http.Request) (Connection, *gid.GID, string, error) // returns: connection, organizationID, continueURL, error
	}

	Connection interface {
		Type() ProtocolType
		Client(ctx context.Context) (*http.Client, error)

		json.Unmarshaler
		json.Marshaler
	}
)

const (
	ProtocolOAuth2 ProtocolType = "OAUTH2"
)

func UnmarshalConnection(protocol string, provider string, data []byte) (Connection, error) {
	switch protocol {
	case string(ProtocolOAuth2):
		switch provider {
		case SlackProvider:
			var slackConn SlackConnection
			if err := json.Unmarshal(data, &slackConn); err != nil {
				return nil, fmt.Errorf("cannot unmarshal slack connection: %w", err)
			}
			return &slackConn, nil

		default:
			var conn OAuth2Connection
			if err := json.Unmarshal(data, &conn); err != nil {
				return nil, fmt.Errorf("cannot unmarshal oauth2 connection: %w", err)
			}
			return &conn, nil
		}
	}

	return nil, fmt.Errorf("unknown connection protocol: %s", protocol)
}
