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
	cloudazure "go.probo.inc/probo/pkg/cloud/azure"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/identityfederation"
)

// azureRegistration declares Azure as a workload identity provider: Probo
// holds no Azure credential and mints an assertion the customer's Entra
// federated identity credential exchanges. It therefore declares no OAuth2,
// API-key or client-credentials path — there is no credential for a customer
// to paste or an operator to configure.
//
// Isolation is the per-organization issuer. A successful token acquisition is
// the whole check, so there is no grant readback beside Probe.
func azureRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderAzure,
		DisplayName:      "Microsoft Azure",
		DocumentationURL: accessReviewDocsURL("azure"),
		// See Registration.EndpointOverrideUnsupported: the Azure SDK resolves
		// every host from the cloud configuration on the session, so there is
		// no host in Endpoints for an override to move.
		EndpointOverrideUnsupported: "the Azure SDK resolves its hosts from the session cloud configuration, not values in Endpoints",
		WorkloadIdentity: &WorkloadIdentityConfig{
			NewSession:      newAzureSession,
			NewDriver:       newAzureDriver,
			Probe:           probeAzure,
			NewNameResolver: newAzureNameResolver,
			ExtraSettings: []ExtraSetting{
				{Key: "tenantId", Label: "Directory (tenant) ID", Required: true},
				{Key: "clientId", Label: "Application (client) ID", Required: true},
				{Key: "subscriptionId", Label: "Subscription ID", Required: true},
				{Key: "environment", Label: "Azure environment", Required: true},
			},
		},
	}
}

// newAzureSession opens a session on the subscription the connector names, by
// exchanging an assertion the customer's federated identity credential trusts.
//
// The organization comes from the connector row, never from its settings: it
// selects whose assertion is minted, and so whose Azure subscription the
// resulting credentials can reach.
func newAzureSession(
	_ context.Context,
	issuer *identityfederation.Issuer,
	conn *coredata.Connector,
) (cloud.Session, error) {
	settings, err := coredata.ConnectorSettings[coredata.AzureConnectorSettings](conn)
	if err != nil {
		return nil, fmt.Errorf("cannot read azure connector settings: %w", err)
	}

	session, err := cloudazure.NewSession(
		issuer,
		conn.OrganizationID,
		settings.TenantID,
		settings.ClientID,
		settings.SubscriptionID,
		cloudazure.Environment(settings.Environment),
	)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func newAzureNameResolver(
	ctx context.Context,
	session cloud.Session,
	conn *coredata.Connector,
	logger *log.Logger,
) drivers.NameResolver {
	azureSession, ok := session.(*cloudazure.Session)
	if !ok {
		logger.ErrorCtx(ctx, "cannot create azure name resolver", log.String("cloud", session.Cloud()))
		return nil
	}

	return drivers.NewAzureNameResolver(
		azureSession,
		logger.With(log.String("connector_id", conn.ID.String())),
	)
}

func newAzureDriver(
	_ context.Context,
	session cloud.Session,
	conn *coredata.Connector,
	logger *log.Logger,
) (drivers.Driver, error) {
	azureSession, ok := session.(*cloudazure.Session)
	if !ok {
		return nil, fmt.Errorf("cannot create azure driver: session is for %s", session.Cloud())
	}

	return drivers.NewAzureDriver(
		azureSession,
		logger.With(log.String("connector_id", conn.ID.String())),
	), nil
}

// probeAzure checks the connection by acquiring an ARM token. It reaches for
// the concrete session because a cloud.Session deliberately exposes only which
// cloud and which account it names.
func probeAzure(ctx context.Context, session cloud.Session, _ *coredata.Connector) error {
	azureSession, ok := session.(*cloudazure.Session)
	if !ok {
		return fmt.Errorf("cannot probe azure connector: session is for %s", session.Cloud())
	}

	return azureSession.CheckAccess(ctx)
}
