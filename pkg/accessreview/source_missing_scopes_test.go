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

package accessreview

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
)

func TestMissingOAuthScopesForConnector(t *testing.T) {
	t.Parallel()

	required := []string{
		"openid",
		"https://graph.microsoft.com/AuditLog.Read.All",
		"https://graph.microsoft.com/User.Read.All",
	}

	t.Run("non oauth protocol returns empty", func(t *testing.T) {
		t.Parallel()

		dbConnector := coredata.Connector{
			Protocol:   coredata.ConnectorProtocolAPIKey,
			Connection: &connector.APIKeyConnection{APIKey: "k"},
		}

		assert.Empty(t, missingOAuthScopesForConnector(dbConnector, required))
	})

	t.Run("empty required returns empty", func(t *testing.T) {
		t.Parallel()

		dbConnector := coredata.Connector{
			Protocol: coredata.ConnectorProtocolOAuth2,
			Connection: &connector.OAuth2Connection{
				Scope: "openid",
			},
		}

		assert.Empty(t, missingOAuthScopesForConnector(dbConnector, nil))
	})

	t.Run("nil connection treats grant as empty", func(t *testing.T) {
		t.Parallel()

		dbConnector := coredata.Connector{
			Protocol:   coredata.ConnectorProtocolOAuth2,
			Connection: nil,
		}

		assert.Equal(
			t,
			[]string{
				"https://graph.microsoft.com/AuditLog.Read.All",
				"https://graph.microsoft.com/User.Read.All",
				"openid",
			},
			missingOAuthScopesForConnector(dbConnector, required),
		)
	})

	t.Run("partial grant returns missing scopes", func(t *testing.T) {
		t.Parallel()

		dbConnector := coredata.Connector{
			Protocol: coredata.ConnectorProtocolOAuth2,
			Connection: &connector.OAuth2Connection{
				Scope: "openid User.Read.All",
			},
		}

		assert.Equal(
			t,
			[]string{"https://graph.microsoft.com/AuditLog.Read.All"},
			missingOAuthScopesForConnector(dbConnector, required),
		)
	})

	t.Run("full grant returns empty", func(t *testing.T) {
		t.Parallel()

		dbConnector := coredata.Connector{
			Protocol: coredata.ConnectorProtocolOAuth2,
			Connection: &connector.OAuth2Connection{
				Scope: "openid AuditLog.Read.All User.Read.All",
			},
		}

		assert.Empty(t, missingOAuthScopesForConnector(dbConnector, required))
	})
}
