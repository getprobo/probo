// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

	// InitiateOptions holds per-call options passed by the caller initiating
	// a connector flow. Different callers may need different configurations
	// for the same provider — for OAuth2, the most common case is requesting
	// a different set of scopes (e.g. SCIM bridge vs access review).
	InitiateOptions struct {
		Scopes []string
		// IncludeGrantedScopes is honored only when the provider has
		// SupportsIncrementalAuth=true.
		IncludeGrantedScopes bool
		// ConnectorID, when set, marks this flow as a reconnect of an
		// existing connector: the callback updates the row in place
		// instead of creating a new one.
		ConnectorID string
		// Site selects a per-customer region/site for multi-site
		// providers (e.g. Datadog). Consumed by the connector's
		// Registration.BuildAuthURLForSite. Empty for single-site
		// providers.
		Site string
	}

	Connector interface {
		Initiate(ctx context.Context, provider string, organizationID gid.GID, opts InitiateOptions, r *http.Request) (string, error)
		Complete(ctx context.Context, r *http.Request) (Connection, *gid.GID, string, error) // returns: connection, organizationID, continueURL, error
	}

	Connection interface {
		Type() ProtocolType
		Client(ctx context.Context) (*http.Client, error)
		Scopes() []string

		json.Unmarshaler
		json.Marshaler
	}
)

const (
	ProtocolOAuth2 ProtocolType = "OAUTH2"
	ProtocolAPIKey ProtocolType = "API_KEY"
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

	case string(ProtocolAPIKey):
		var conn APIKeyConnection
		if err := json.Unmarshal(data, &conn); err != nil {
			return nil, fmt.Errorf("cannot unmarshal api key connection: %w", err)
		}

		return &conn, nil
	}

	return nil, fmt.Errorf("unknown connection protocol: %s", protocol)
}
