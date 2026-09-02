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

// Package aws vends read-only AWS credentials for a customer account by OIDC
// web identity federation.
//
// Probo holds no AWS credential for any customer. It mints a short-lived
// assertion (pkg/identityfederation), calls sts:AssumeRoleWithWebIdentity, and
// STS verifies the assertion against Probo's published JWKS before evaluating
// the customer's own trust policy. The customer revokes by deleting their role.
package aws

import (
	"context"
	"fmt"
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"go.gearno.de/kit/httpclient"
	"go.probo.inc/probo/pkg/cloud"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/identityfederation"
)

const (
	// CommercialPartition is the worldwide AWS partition.
	CommercialPartition = "aws"

	// GovPartition is AWS GovCloud (US).
	GovPartition = "aws-us-gov"

	// ChinaPartition is AWS China.
	ChinaPartition = "aws-cn"

	// DefaultCommercialRegion is the STS endpoint for CommercialPartition.
	// It only selects the STS host; IAM is global within the partition.
	DefaultCommercialRegion = "us-east-1"

	// DefaultGovRegion is the STS endpoint for GovPartition.
	DefaultGovRegion = "us-gov-west-1"

	// DefaultChinaRegion is the STS endpoint for ChinaPartition.
	DefaultChinaRegion = "cn-north-1"

	// roleSessionNamePrefix labels the assumed-role session in the customer's
	// CloudTrail, so they can attribute every call to the Probo organization
	// that made it. GID characters are all within the session-name charset.
	roleSessionNamePrefix = "probo-"
)

type (
	// Session is authenticated access to one AWS account, held as an
	// aws.Config whose credentials refresh themselves through web identity.
	// Build service clients from Config; do not copy the credentials out.
	Session struct {
		cfg       awssdk.Config
		accountID string
		partition string
	}

	// issuerTokenRetriever adapts the issuer to the AWS SDK's
	// IdentityTokenRetriever, which the SDK calls on every credential refresh.
	issuerTokenRetriever struct {
		issuer         *identityfederation.Issuer
		organizationID gid.GID
	}
)

var (
	_ cloud.Session                   = (*Session)(nil)
	_ stscreds.IdentityTokenRetriever = (*issuerTokenRetriever)(nil)
)

// NewSession opens a session on the account owning roleARN, by exchanging an
// assertion minted for organizationID.
//
// organizationID must come from the connector row being serviced, never from
// user input: it selects whose cloud accounts the resulting credentials reach.
//
// No credential is fetched here, which is why there is no context to pass. The
// SDK performs the exchange lazily on the first API call and owns caching and
// refresh from then on, so a session is cheap to build and no token is ever
// written to disk.
func NewSession(
	issuer *identityfederation.Issuer,
	organizationID gid.GID,
	roleARN string,
) (*Session, error) {
	parsedARN, err := arn.Parse(roleARN)
	if err != nil {
		return nil, fmt.Errorf("cannot open aws session: cannot parse role ARN: %w", err)
	}

	if parsedARN.AccountID == "" {
		return nil, fmt.Errorf("cannot open aws session: role ARN carries no account ID")
	}

	region := regionForPartition(parsedARN.Partition)

	httpClient := httpclient.DefaultPooledClient(httpclient.WithSSRFProtection())

	// AssumeRoleWithWebIdentity is the one STS call that takes no credential —
	// the assertion is the credential. Signing it anonymously also keeps any
	// ambient credentials in Probo's own environment out of the exchange.
	stsClient := sts.NewFromConfig(
		awssdk.Config{
			Region:      region,
			Credentials: awssdk.AnonymousCredentials{},
			HTTPClient:  httpClient,
		},
	)

	provider := stscreds.NewWebIdentityRoleProvider(
		stsClient,
		roleARN,
		&issuerTokenRetriever{
			issuer:         issuer,
			organizationID: organizationID,
		},
		func(wo *stscreds.WebIdentityRoleOptions) {
			wo.RoleSessionName = roleSessionNamePrefix + organizationID.String()
			wo.Duration = time.Hour
		},
	)

	return &Session{
		cfg: awssdk.Config{
			Region:      region,
			Credentials: awssdk.NewCredentialsCache(provider),
			HTTPClient:  httpClient,
		},
		accountID: parsedARN.AccountID,
		partition: parsedARN.Partition,
	}, nil
}

// regionForPartition is any STS region in the partition the role ARN names.
// IAM roles are global; the region only selects the STS host, which cannot
// cross partitions.
func regionForPartition(partition string) string {
	switch partition {
	case GovPartition:
		return DefaultGovRegion
	case ChinaPartition:
		return DefaultChinaRegion
	default:
		return DefaultCommercialRegion
	}
}

// Cloud implements cloud.Session.
func (s *Session) Cloud() string {
	return cloud.AWS
}

// AccountID is the AWS account the assumed role lives in.
func (s *Session) AccountID() string {
	return s.accountID
}

// Partition is the AWS partition the assumed role lives in: commercial,
// GovCloud, or China. IAM ARNs are partition-scoped, so a constructed
// identity must use this rather than assuming the commercial form.
func (s *Session) Partition() string {
	return s.partition
}

// Config returns the SDK config to build service clients from.
func (s *Session) Config() awssdk.Config {
	return s.cfg
}

// NewSessionFromConfig builds a session from an already-resolved SDK config.
// Production uses NewSession, which obtains credentials through web identity.
func NewSessionFromConfig(accountID, partition string, cfg awssdk.Config) *Session {
	return &Session{cfg: cfg, accountID: accountID, partition: partition}
}

// CheckAccess reports whether this session can actually reach its account.
//
// It calls sts:GetCallerIdentity, which every principal may call regardless of
// its policies, so a failure means the web identity exchange itself was
// refused — a missing role, a trust policy that does not name this
// organization, an unpublished signing key — and never a missing permission on
// the role. That is what makes it a connection check rather than a capability
// check.
//
// It is also the first call to force the exchange: NewSession fetches no
// credential, so until something calls AWS there is nothing to be wrong.
func (s *Session) CheckAccess(ctx context.Context) error {
	client := sts.NewFromConfig(s.cfg)

	_, err := client.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return fmt.Errorf("cannot reach aws account: %w", err)
	}

	return nil
}

// GetIdentityToken mints the assertion STS exchanges for credentials.
//
// The SDK's interface passes no context. That costs nothing here because
// minting is entirely in-process — an RSA signature over claims already in
// memory, with no I/O to cancel.
//
// The returned token is a bearer credential for the customer's cloud account:
// it must never reach a log or an error message.
func (r *issuerTokenRetriever) GetIdentityToken() ([]byte, error) {
	token, err := r.issuer.Token(
		context.Background(),
		r.organizationID,
		identityfederation.AudienceAWS,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot mint aws identity federation token: %w", err)
	}

	return []byte(token), nil
}
