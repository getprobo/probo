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
			// This token posts legacy messages and may lack users:read, so
			// only a dead token, which cannot post either, is reported.
			if slackLegacyMessaging(dbConnector) {
				return slackProbeVerdict(drivers.CheckSlackTokenAlive(ctx, c, ep.APIBase))
			}

			// A bot token is never an admin, so only check that it still works.
			if slackBotToken(dbConnector) {
				return slackProbeVerdict(drivers.CheckSlackToken(ctx, c, ep.APIBase))
			}

			return slackProbeVerdict(drivers.CheckSlackInstallerIsAdmin(ctx, c, ep.APIBase))
		},
		NeedsReconnect: slackNeedsReconnect,
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

// slackLegacyMessaging reports a bot-token connection that still posts
// legacy Slack messages: a user token has no chat:write, so it must keep its
// token.
func slackLegacyMessaging(dbConnector *coredata.Connector) bool {
	conn, ok := dbConnector.Connection.(*connector.SlackConnection)

	return ok && !conn.IsUserToken() && conn.Settings.ChannelID != ""
}

// slackNeedsReconnect asks bot-token connections to reconnect, except one
// that still posts legacy Slack messages, whatever scopes it misses.
func slackNeedsReconnect(dbConnector *coredata.Connector, missingScopes []string) bool {
	if slackLegacyMessaging(dbConnector) {
		return false
	}

	if slackBotToken(dbConnector) {
		return true
	}

	return len(missingScopes) > 0
}

// slackProbeVerdict maps a Slack check onto the probe's verdicts. Slack
// answers a dead token with HTTP 200 and ok=false. Any other failure, such as
// throttling, an outage or a workspace migration, stays inconclusive, as it
// is for the plain GET probe.
func slackProbeVerdict(err error) error {
	if err == nil {
		return nil
	}

	if _, ok := errors.AsType[*drivers.InstallRejectedError](err); ok {
		return newCredentialRejected(http.StatusForbidden)
	}

	if statusErr, ok := errors.AsType[*drivers.SlackStatusError](err); ok {
		switch statusErr.StatusCode {
		case http.StatusUnauthorized, http.StatusForbidden:
			return newCredentialRejected(statusErr.StatusCode)
		default:
			return nil
		}
	}

	if apiErr, ok := errors.AsType[*drivers.SlackAPIError](err); ok {
		switch apiErr.Code {
		case "invalid_auth", "not_authed", "token_revoked", "token_expired", "account_inactive":
			return newCredentialRejected(http.StatusUnauthorized)
		case "missing_scope", "no_permission", "two_factor_setup_required":
			return newCredentialRejected(http.StatusForbidden)
		default:
			return nil
		}
	}

	return err
}
