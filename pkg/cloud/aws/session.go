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
	"net/http"
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
	// DefaultRegion is where STS is called when the connector names no region.
	// It only selects the STS endpoint; IAM and Organizations are global.
	DefaultRegion = "us-east-1"

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
	}

	// Option configures a Session.
	Option func(*options)

	options struct {
		region      string
		stsEndpoint string
		httpClient  *http.Client
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

// WithRegion selects the region the STS exchange targets.
func WithRegion(region string) Option {
	return func(o *options) { o.region = region }
}

// WithSTSEndpoint overrides the STS endpoint URL. Production leaves it unset
// and lets the SDK resolve the regional endpoint.
func WithSTSEndpoint(endpoint string) Option {
	return func(o *options) { o.stsEndpoint = endpoint }
}

// WithHTTPClient replaces the SSRF-protected client the STS exchange uses.
func WithHTTPClient(c *http.Client) Option {
	return func(o *options) { o.httpClient = c }
}

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
	opts ...Option,
) (*Session, error) {
	parsedARN, err := arn.Parse(roleARN)
	if err != nil {
		return nil, fmt.Errorf("cannot open aws session: cannot parse role ARN: %w", err)
	}

	if parsedARN.AccountID == "" {
		return nil, fmt.Errorf("cannot open aws session: role ARN carries no account ID")
	}

	var o options
	for _, opt := range opts {
		opt(&o)
	}

	if o.region == "" {
		o.region = DefaultRegion
	}

	if o.httpClient == nil {
		o.httpClient = httpclient.DefaultPooledClient(httpclient.WithSSRFProtection())
	}

	// AssumeRoleWithWebIdentity is the one STS call that takes no credential —
	// the assertion is the credential. Signing it anonymously also keeps any
	// ambient credentials in Probo's own environment out of the exchange.
	stsClient := sts.NewFromConfig(
		awssdk.Config{
			Region:      o.region,
			Credentials: awssdk.AnonymousCredentials{},
			HTTPClient:  o.httpClient,
		},
		func(so *sts.Options) {
			if o.stsEndpoint != "" {
				so.BaseEndpoint = awssdk.String(o.stsEndpoint)
			}
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
			Region:      o.region,
			Credentials: awssdk.NewCredentialsCache(provider),
			HTTPClient:  o.httpClient,
		},
		accountID: parsedARN.AccountID,
	}, nil
}

// Cloud implements cloud.Session.
func (s *Session) Cloud() string {
	return cloud.AWS
}

// AccountID is the AWS account the assumed role lives in.
func (s *Session) AccountID() string {
	return s.accountID
}

// Config returns the SDK config to build service clients from.
func (s *Session) Config() awssdk.Config {
	return s.cfg
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
