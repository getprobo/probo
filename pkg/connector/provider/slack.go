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

package provider

import (
	"context"
	"errors"
	"net/http"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
)

func slackRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderSlack,
		DisplayName:      "Slack",
		DocumentationURL: accessReviewDocsURL("slack"),
		Endpoints: Endpoints{
			Auth:    "https://slack.com/oauth/v2/authorize",
			Token:   "https://slack.com/api/oauth.v2.access",
			APIBase: "https://slack.com/api",
		},
		// has_2fa is only returned to an admin or owner caller, which a bot
		// never is, so the connector asks for a user token alone. Exclusive
		// so a reconnect never replays a bot-only scope as a user scope.
		OAuth2: &OAuth2Config{
			Scopes:          []string{"users:read", "users:read.email"},
			ScopeParam:      "user_scope",
			ScopeSeparator:  ",",
			ExclusiveScopes: true,
		},
		Probe: func(ctx context.Context, c *http.Client, dbConnector *coredata.Connector, ep Endpoints) error {
			// A bot token is never an admin; NeedsReconnect reports it.
			if slackBotToken(dbConnector) {
				return nil
			}

			return slackProbeVerdict(drivers.CheckSlackInstallerIsAdmin(ctx, c, ep.APIBase))
		},
		NeedsReconnect: slackBotToken,
		NewDriver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger, ep Endpoints) (drivers.Driver, error) {
			return drivers.NewSlackDriver(c, ep.APIBase), nil
		},
		NewNameResolver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger, ep Endpoints) drivers.NameResolver {
			return drivers.NewSlackNameResolver(c, ep.APIBase)
		},
		ValidateInstall: func(ctx context.Context, c *http.Client, ep Endpoints) error {
			return drivers.CheckSlackInstallerIsAdmin(ctx, c, ep.APIBase)
		},
	}
}

// slackBotToken reports a connection made before the connector asked for a
// user token.
func slackBotToken(dbConnector *coredata.Connector) bool {
	conn, ok := dbConnector.Connection.(*connector.SlackConnection)

	return ok && !conn.IsUserToken()
}

// slackProbeVerdict maps the installer check onto the probe's verdicts. Slack
// answers a dead token with HTTP 200 and ok=false, so a plain GET probe
// would report it connected; an installer who lost the admin role keeps a
// working token that can no longer see has_2fa.
func slackProbeVerdict(err error) error {
	if err == nil {
		return nil
	}

	if _, ok := errors.AsType[*drivers.InstallRejectedError](err); ok {
		return newCredentialRejected(http.StatusForbidden)
	}

	if apiErr, ok := errors.AsType[*drivers.SlackAPIError](err); ok {
		switch apiErr.Code {
		case "invalid_auth", "not_authed", "token_revoked", "token_expired", "account_inactive":
			return newCredentialRejected(http.StatusUnauthorized)
		case "missing_scope", "no_permission":
			return newCredentialRejected(http.StatusForbidden)
		}
	}

	return err
}
