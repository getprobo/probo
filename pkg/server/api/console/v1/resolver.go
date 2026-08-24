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

//go:generate go tool github.com/99designs/gqlgen generate

package console_v1

import (
	"context"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview"
	"go.probo.inc/probo/pkg/agentexecution"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/certmanager"
	"go.probo.inc/probo/pkg/complianceportal/management"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/cookiebanner"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/esign"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/itam"
	"go.probo.inc/probo/pkg/mailman"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/probot"
	slackchannel "go.probo.inc/probo/pkg/probot/channel/slack"
	"go.probo.inc/probo/pkg/probot/identitybinding"
	"go.probo.inc/probo/pkg/resourcealias"
	"go.probo.inc/probo/pkg/riskmanagement"
	"go.probo.inc/probo/pkg/saferedirect"
	"go.probo.inc/probo/pkg/securecookie"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/api/authz"
	"go.probo.inc/probo/pkg/server/api/console/v1/dataloader"
	"go.probo.inc/probo/pkg/server/api/console/v1/types"
	"go.probo.inc/probo/pkg/server/gqlutils"
	"go.probo.inc/probo/pkg/thirdparty"
)

type (
	BotDeliveryDestinations interface {
		GetDestination(ctx context.Context, scope coredata.Scoper, organizationID gid.GID, target probot.DeliveryTarget) (*coredata.BotDeliveryDestination, error)
		SetDestination(ctx context.Context, scope coredata.Scoper, organizationID gid.GID, target probot.DeliveryTarget, externalDestinationID string) (*coredata.BotDeliveryDestination, error)
		RestoreDestination(ctx context.Context, scope coredata.Scoper, organizationID gid.GID, target probot.DeliveryTarget, previous *coredata.BotDeliveryDestination, expectedExternalDestinationID string) (*coredata.BotDeliveryDestination, error)
		ClearDestination(ctx context.Context, scope coredata.Scoper, organizationID gid.GID, target probot.DeliveryTarget) error
	}

	ComplianceMessages interface {
		QueueWelcome(ctx context.Context, organizationID, compliancePortalID gid.GID) error
	}

	Resolver struct {
		authorize               authz.AuthorizeFunc
		batchAuthorize          authz.BatchAuthorizeFunc
		probo                   *probo.Service
		resourceAlias           *resourcealias.Service
		iam                     *iam.Service
		esign                   *esign.Service
		management              *management.Service
		certManager             *certmanager.Service
		accessReview            *accessreview.Service
		agentExecution          *agentexecution.Service
		mailman                 *mailman.Service
		cookieBanner            *cookiebanner.Service
		connectorRegistry       *connector.Registry
		providerRegistry        *provider.Registry
		riskManagement          *riskmanagement.Service
		thirdParty              *thirdparty.Service
		itam                    *itam.Service
		logger                  *log.Logger
		fileManager             *filemanager.Service
		baseURL                 *baseurl.BaseURL
		customDomainCname       string
		tokenSecret             string
		probotIdentityBindings  *identitybinding.Service
		slackbotInstallations   *slackchannel.InstallationService
		botDeliveryDestinations BotDeliveryDestinations
		complianceMessages      ComplianceMessages
	}
)

func NewMux(
	logger *log.Logger,
	proboSvc *probo.Service,
	resourceAliasSvc *resourcealias.Service,
	iamSvc *iam.Service,
	esignSvc *esign.Service,
	managementSvc *management.Service,
	certManagerSvc *certmanager.Service,
	accessReviewSvc *accessreview.Service,
	agentExecutionSvc *agentexecution.Service,
	mailmanSvc *mailman.Service,
	cookieBannerSvc *cookiebanner.Service,
	cookieConfig securecookie.Config,
	tokenSecret string,
	connectorRegistry *connector.Registry,
	providerRegistry *provider.Registry,
	fileManagerSvc *filemanager.Service,
	baseURL *baseurl.BaseURL,
	customDomainCname string,
	thirdPartySvc *thirdparty.Service,
	riskManagementSvc *riskmanagement.Service,
	probotIdentityBindings *identitybinding.Service,
	slackbotInstallations *slackchannel.InstallationService,
	botDeliveryDestinations BotDeliveryDestinations,
	complianceMessages ComplianceMessages,
	graphqlLimits gqlutils.Limits,
	itamSvc *itam.Service,
) *chi.Mux {
	r := chi.NewMux()

	safeRedirect := saferedirect.New(saferedirect.StaticHosts(baseURL.Host()))

	graphqlHandler := NewGraphQLHandler(
		iamSvc,
		proboSvc,
		resourceAliasSvc,
		esignSvc,
		managementSvc,
		certManagerSvc,
		accessReviewSvc,
		agentExecutionSvc,
		mailmanSvc,
		cookieBannerSvc,
		connectorRegistry,
		providerRegistry,
		customDomainCname,
		tokenSecret,
		logger,
		thirdPartySvc,
		riskManagementSvc,
		fileManagerSvc,
		baseURL,
		graphqlLimits,
		itamSvc,
		probotIdentityBindings,
		slackbotInstallations,
		botDeliveryDestinations,
		complianceMessages,
	)

	r.Group(func(r chi.Router) {
		r.Use(authn.NewSessionMiddleware(iamSvc, cookieConfig))
		r.Use(authn.NewAPIKeyMiddleware(iamSvc, tokenSecret))
		r.Use(authn.NewOAuth2AccessTokenMiddleware(iamSvc))
		r.Use(authn.NewIdentityPresenceMiddleware(baseURL))
		r.Use(newMembershipAccessMiddleware(iamSvc, logger))
		r.Use(dataloader.NewMiddleware(
			proboSvc,
			iamSvc,
			cookieBannerSvc,
			thirdPartySvc,
			managementSvc,
		))

		r.Handle("/graphql", graphqlHandler)

		r.Get(
			"/connectors/initiate",
			handleConnectorInitiate(logger, proboSvc, iamSvc, connectorRegistry),
		)

		r.Get(
			"/connectors/github-app/initiate",
			handleConnectorGitHubAppInitiate(logger, proboSvc, iamSvc, connectorRegistry),
		)

		r.Get(
			"/connectors/complete",
			handleConnectorComplete(
				logger,
				baseURL,
				proboSvc,
				accessReviewSvc,
				connectorRegistry,
				safeRedirect,
			),
		)

		r.Get(
			"/connectors/github-app/complete",
			handleConnectorGitHubAppComplete(
				logger,
				baseURL,
				proboSvc,
				accessReviewSvc,
				connectorRegistry,
				safeRedirect,
			),
		)

		r.Get(
			"/slackbot/install/initiate",
			handleSlackbotInstallInitiate(
				logger,
				iamSvc,
				slackbotInstallations,
			),
		)
	})

	r.Get(
		"/slackbot/install/complete",
		handleSlackbotInstallComplete(
			logger,
			baseURL,
			slackbotInstallations,
			safeRedirect,
		),
	)

	// Public, unauthenticated: the OAuth Client ID Metadata Document (CIMD)
	// is fetched server-to-server by public-client providers (PostHog)
	// during authorization, with no Probo credentials. Mounted outside the
	// auth group above.
	r.Get("/connectors/oauth-client-metadata", handleConnectorOAuth2ClientMetadata(baseURL))

	return r
}

func (r *Resolver) Permission(ctx context.Context, obj types.Node, action string) (bool, error) {
	_, err := r.authorize(ctx, obj.GetID(), action, authz.WithDryRun())

	return err == nil, nil
}
