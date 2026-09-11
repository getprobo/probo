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
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/coredata"
)

func crispRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderCrisp,
		DisplayName:      "Crisp",
		DocumentationURL: accessReviewDocsURL("crisp"),
		// Model B: the plugin token is Probo's own Crisp Marketplace plugin
		// credential, held server-side in bootstrap config, not pasted by
		// the customer. ManagedAPIKey injects it at connect time; the customer
		// supplies nothing at all, the install ceremony below yielding the
		// website id. SupportsAPIKey stays false so the provider is hidden
		// from the driver catalog until the operator configures
		// PROBOD_CONNECTOR_CRISP_PLUGIN_TOKEN — it ships deactivated until
		// Crisp validates the production plugin and activates with no code
		// change once the token is set.
		APIKey: &APIKeyConfig{
			Auth: APIKeyAuth{Mode: APIKeyAuthBasicUserPass},
			// No ExtraSettings: the install ceremony supplies the website id,
			// so there is no customer-typed value and no dialog to render.
			Managed: &ManagedAPIKey{RequiresResourceID: true},
		},
		// The per-website plugin API also needs the plugin ID (a distinct value
		// from the token identifier), supplied via bootstrap alongside the
		// token. Require it so Crisp stays hidden until both are configured
		// rather than surfacing as connectable and failing at connect time; it
		// is also what the install redirect interpolates.
		// Crisp authenticates with the plugin token presented as HTTP Basic,
		// the credential being the verbatim "identifier:key" pair.
		// APIKeyBasicAuthUserPass base64-encodes it (the empty-password
		// APIKeyBasicAuth cannot carry the key). Every request also needs the
		// non-auth X-Crisp-Tier header (set by the driver/probe/name resolver),
		// so the probe is a custom closure.
		// Two hosts an override cannot move: the install redirect pins
		// app.crisp.chat, and GetCrispSubscription reaches crispDefaultBaseURL
		// directly rather than APIBase.
		EndpointOverrideUnsupported: "its plugin-install flow pins app.crisp.chat and GetCrispSubscription pins a host outside APIBase",
		Endpoints: Endpoints{
			// Every endpoint the driver calls shares the /v1 prefix, so the
			// version segment stays in APIBase.
			APIBase: "https://api.crisp.chat/v1",
			// Crisp forces https and takes one callback URL per plugin, so each
			// Probo deployment registers its own Crisp Marketplace plugin. %s
			// is the plugin id.
			Install: "https://app.crisp.chat/initiate/plugin/%s/",
		},
		// Crisp carries no redirect_uri of its own: the return URL is typed into
		// the Marketplace plugin's own settings, once per plugin. Each Probo
		// deployment therefore needs its own plugin, whose Callback URL must be
		// exactly
		//
		//	https://<deployment host>/api/console/v1/connectors/install/crisp/complete
		//
		// matching the route mounted in console/v1/resolver.go. Crisp rewrites
		// the scheme to https unconditionally, so a local stack has to serve
		// TLS rather than plain http.
		Install: &InstallConfig{
			StateParam:          "payload", // Crisp echoes it byte-identical
			SettingsResourceKey: "website_id",
			Verify: func(ctx context.Context, c *http.Client, pluginID string, q url.Values) (string, error) {
				// Parse AND re-spell. uuid.Parse accepts canonical, uppercase,
				// urn:uuid:, braced and 32-hex forms and returns an equal value
				// for all five, so validating without canonicalizing would let
				// one Crisp website bind five times: five distinct
				// settings->>'website_id' strings, five advisory-lock keys,
				// five connectors. The canonical form is what the vendor call,
				// the persisted setting and the idempotency key all use from
				// here on.
				parsed, err := uuid.Parse(q.Get("website_id"))
				if err != nil {
					return "", fmt.Errorf("crisp callback carries an invalid website id")
				}

				websiteID := parsed.String()

				proof := q.Get("token")
				if proof == "" {
					return "", fmt.Errorf("crisp callback carries no subscription token")
				}

				// Terminal vs retryable, decided on the STATUS and not on the
				// error string. A terminal error wrongly marked transient
				// releases the customer's single-use state into a retry loop
				// that cannot succeed; a retryable one wrongly marked terminal
				// burns their ten-minute window over a vendor blip.
				sub, err := drivers.GetCrispSubscription(ctx, c, websiteID, pluginID)
				if err != nil {
					// Probo's plugin is not subscribed to this website. No
					// retry installs it for them.
					if errors.Is(err, drivers.ErrCrispPluginNotSubscribed) {
						return "", err
					}

					if statusErr, ok := errors.AsType[*drivers.CrispStatusError](err); ok {
						switch {
						case statusErr.Code == http.StatusTooManyRequests,
							statusErr.Code >= 500:
							return "", fmt.Errorf("%w: %w", ErrInstallVerificationTransient, err)
						default:
							// 401/403 mean Probo's own plugin credential is
							// wrong, revoked or cross-wired with the plugin id;
							// every other 4xx is a request this code will keep
							// making identically. Terminal in both cases.
							return "", err
						}
					}

					// No status reached us: a dial/TLS/timeout failure, or a
					// body that would not decode. The first is plainly
					// retryable, the second is not, but the two are
					// indistinguishable here and releasing is the safe
					// direction — the state still expires on its own inside ten
					// minutes.
					return "", fmt.Errorf("%w: %w", ErrInstallVerificationTransient, err)
				}

				if subtle.ConstantTimeCompare([]byte(sub.Token), []byte(proof)) != 1 {
					return "", fmt.Errorf("crisp subscription token mismatch")
				}

				return websiteID, nil
			},
		},
		Probe: probeCrisp,
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger, ep Endpoints) (drivers.Driver, error) {
			s, err := coredata.ConnectorSettings[coredata.CrispConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read crisp connector settings: %w", err)
			}

			if s.WebsiteID == "" {
				return nil, fmt.Errorf("cannot create crisp driver: website_id is required")
			}

			return drivers.NewCrispDriver(c, s.WebsiteID, ep.APIBase), nil
		},
		NewNameResolver: func(ctx context.Context, c *http.Client, conn *coredata.Connector, logger *log.Logger, ep Endpoints) drivers.NameResolver {
			s, err := coredata.ConnectorSettings[coredata.CrispConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read crisp connector settings", log.Error(err))

				return nil
			}

			return drivers.NewCrispNameResolver(c, s.WebsiteID, ep.APIBase)
		},
	}
}
