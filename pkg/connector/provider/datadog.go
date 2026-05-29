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

package provider

import (
	"context"
	"fmt"
	"net/http"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
)

func datadogRegistration() *Registration {
	// Datadog is multi-site: the customer's region drives the authorize
	// host (built at initiate from the region pick) and the token + API
	// host (built at callback from Datadog's `domain` param). AuthURL,
	// TokenURL, and ProbeURL are therefore empty — the closures build the
	// per-customer hosts, and an empty probe is skipped (a dead token
	// surfaces on the first ListAccounts). Confidential client + PKCE map
	// to the default post-form token-endpoint auth.
	return &Registration{
		Provider:               coredata.ConnectorProviderDatadog,
		DisplayName:            "Datadog",
		OAuth2Scopes:           []string{"user_access_read"},
		RequiresPKCE:           true,
		BuildAuthURLForSite:    connector.DatadogAuthorizeURL,
		BuildTokenURLForDomain: connector.DatadogTokenURL,
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			s, err := coredata.ConnectorSettings[coredata.DatadogConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read datadog connector settings: %w", err)
			}

			if s.Domain == "" {
				return nil, fmt.Errorf("cannot create datadog driver: domain is required")
			}

			return drivers.NewDatadogDriver(c, s.Domain), nil
		},
		NewNameResolver: func(ctx context.Context, _ *http.Client, conn *coredata.Connector, logger *log.Logger) drivers.NameResolver {
			s, err := coredata.ConnectorSettings[coredata.DatadogConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read datadog connector settings", log.Error(err))
				return nil
			}

			return drivers.NewDatadogNameResolver(s.Region)
		},
	}
}
