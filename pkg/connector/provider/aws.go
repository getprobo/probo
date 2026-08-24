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
// IMPLIED, INCLUDING WITHOUT LIMITATION THE WARRANTIES OF MERCHANTABILITY,
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
	cloudaws "go.probo.inc/probo/pkg/cloud/aws"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/identityfederation"
)

// awsRegistration declares AWS as a workload identity provider: Probo holds no
// AWS credential and mints an assertion the customer's STS exchanges for
// temporary ones. It therefore declares no OAuth2, API-key or
// client-credentials path — there is no credential for a customer to paste or
// an operator to configure. It also registers no InspectGrant: each
// organization has its own issuer, so STS rejects a foreign token before any
// trust-policy condition runs.
func awsRegistration() *Registration {
	return &Registration{
		Provider:                    coredata.ConnectorProviderAWS,
		DisplayName:                 "Amazon Web Services",
		EndpointOverrideUnsupported: "the AWS SDK resolves its own endpoints from the session region, not from values in Endpoints",
		WorkloadIdentity: &WorkloadIdentityConfig{
			NewSession: newAWSSession,
			NewDriver:  newAWSDriver,
			Probe:      probeAWS,
		},
	}
}

// newAWSSession opens a session on the account the connector names, by
// assuming the role the customer created for Probo there.
//
// The organization comes from the connector row, never from its settings: it
// selects whose assertion is minted, and so whose cloud account the resulting
// credentials can reach.
func newAWSSession(
	_ context.Context,
	issuer *identityfederation.Issuer,
	conn *coredata.Connector,
) (cloud.Session, error) {
	settings, err := coredata.ConnectorSettings[coredata.AWSConnectorSettings](conn)
	if err != nil {
		return nil, fmt.Errorf("cannot read aws connector settings: %w", err)
	}

	session, err := cloudaws.NewSession(issuer, conn.OrganizationID, settings.RoleARN)
	if err != nil {
		return nil, err
	}

	return session, nil
}

// newAWSDriver builds the access review driver. Listing IAM users in the
// connected account is not implemented yet.
func newAWSDriver(
	_ context.Context,
	_ cloud.Session,
	_ *coredata.Connector,
	_ *log.Logger,
) (drivers.Driver, error) {
	return nil, fmt.Errorf("cannot create aws driver: access review for AWS is not implemented")
}

// probeAWS checks the connection by asking AWS who we are. It reaches for the
// concrete session because a cloud.Session deliberately exposes only which
// cloud and which account it names.
func probeAWS(ctx context.Context, session cloud.Session, _ *coredata.Connector) error {
	awsSession, ok := session.(*cloudaws.Session)
	if !ok {
		return fmt.Errorf("cannot probe aws connector: session is for %s", session.Cloud())
	}

	return awsSession.CheckAccess(ctx)
}
