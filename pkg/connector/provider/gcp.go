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
	"fmt"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/cloud"
	cloudgcp "go.probo.inc/probo/pkg/cloud/gcp"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/identityfederation"
)

// gcpRegistration declares GCP as a workload identity provider: Probo holds no
// GCP credential and mints an assertion the customer's Workload Identity
// Federation pool exchanges, then impersonates the named service account. It
// therefore declares no OAuth2, API-key or client-credentials path — there is
// no credential for a customer to paste or an operator to configure.
//
// Isolation is the per-organization issuer; a successful impersonation is the
// whole check, so there is no grant readback beside Probe.
func gcpRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderGCP,
		DisplayName:      "Google Cloud",
		DocumentationURL: accessReviewDocsURL("gcp"),
		// See Registration.EndpointOverrideUnsupported: Google API clients
		// resolve every host they dial themselves, so there is no host in
		// Endpoints for an override to move.
		EndpointOverrideUnsupported: "the GCP APIs resolve their own hosts, not values in Endpoints",
		WorkloadIdentity: &WorkloadIdentityConfig{
			NewSession:      newGCPSession,
			NewDriver:       newGCPDriver,
			Probe:           probeGCP,
			NewNameResolver: newGCPNameResolver,
			ExtraSettings: []ExtraSetting{
				{Key: "workloadIdentityProvider", Label: "Workload identity provider", Required: true},
				{Key: "serviceAccountEmail", Label: "Service account email", Required: true},
			},
		},
	}
}

// newGCPSession opens a session on the project the connector names, by
// exchanging an assertion and impersonating the service account the customer
// created for Probo there.
//
// The organization comes from the connector row, never from its settings: it
// selects whose assertion is minted, and so whose cloud project the resulting
// credentials can reach.
func newGCPSession(
	_ context.Context,
	issuer *identityfederation.Issuer,
	conn *coredata.Connector,
) (cloud.Session, error) {
	settings, err := coredata.ConnectorSettings[coredata.GCPConnectorSettings](conn)
	if err != nil {
		return nil, fmt.Errorf("cannot read gcp connector settings: %w", err)
	}

	session, err := cloudgcp.NewSession(
		issuer,
		conn.OrganizationID,
		settings.WorkloadIdentityProvider,
		settings.ServiceAccountEmail,
	)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func newGCPNameResolver(
	ctx context.Context,
	session cloud.Session,
	conn *coredata.Connector,
	logger *log.Logger,
) drivers.NameResolver {
	gcpSession, ok := session.(*cloudgcp.Session)
	if !ok {
		logger.ErrorCtx(ctx, "cannot create gcp name resolver", log.String("cloud", session.Cloud()))
		return nil
	}

	settings, err := coredata.ConnectorSettings[coredata.GCPConnectorSettings](conn)
	if err != nil {
		logger.ErrorCtx(ctx, "cannot read gcp connector settings", log.Error(err))
		return nil
	}

	return drivers.NewGCPNameResolver(gcpSession, settings.ServiceAccountEmail, logger)
}

// newGCPDriver builds the access review driver over the session already
// impersonated on the connected project.
func newGCPDriver(
	_ context.Context,
	session cloud.Session,
	_ *coredata.Connector,
	logger *log.Logger,
) (drivers.Driver, error) {
	gcpSession, ok := session.(*cloudgcp.Session)
	if !ok {
		return nil, fmt.Errorf("cannot create gcp driver: session is for %s", session.Cloud())
	}

	return drivers.NewGCPDriver(gcpSession, logger), nil
}

// probeGCP checks the connection by completing the STS exchange and
// impersonation. It reaches for the concrete session because a cloud.Session
// deliberately exposes only which cloud and which account it names.
func probeGCP(ctx context.Context, session cloud.Session, _ *coredata.Connector) error {
	gcpSession, ok := session.(*cloudgcp.Session)
	if !ok {
		return fmt.Errorf("cannot probe gcp connector: session is for %s", session.Cloud())
	}

	return gcpSession.CheckAccess(ctx)
}
