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

// Package aws vends read-only AWS credentials for a customer account through
// OIDC web identity federation: Probo mints a short-lived token asserting which
// organization it acts for, and STS exchanges it for temporary credentials
// against a role whose trust policy the customer owns.
//
// No credential is stored and no token touches disk.
package aws

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"go.gearno.de/kit/httpclient"
	"go.probo.inc/probo/pkg/cloud"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/identityfederation"
)

const (
	// DefaultRegion applies when the caller names none. STS is global, but the
	// SDK signs regionally.
	DefaultRegion = "us-east-1"

	// roleSessionName labels Probo's sessions in the customer's CloudTrail. AWS
	// restricts it to [\w+=,.@-]{2,64}, which rules out the module path used
	// elsewhere in this repo.
	roleSessionName = "probo-audit"

	// account ID position in arn:partition:service:region:account-id:resource.
	arnAccountIDField = 4
)

type (
	// Session is authenticated access to one AWS account. Its credentials are
	// temporary and refreshed in memory, so a Session is worth keeping for one
	// job and never worth persisting.
	Session struct {
		cfg       awssdk.Config
		accountID string
	}

	Option func(*options)

	options struct {
		region      string
		stsEndpoint string
		httpClient  *http.Client
	}
)

var _ cloud.Session = (*Session)(nil)

// WithRegion selects the region the credential exchange is signed for.
func WithRegion(region string) Option {
	return func(o *options) {
		if region != "" {
			o.region = region
		}
	}
}

// WithSTSEndpoint redirects the credential exchange. It applies to the STS
// client only, not to the clients a driver later builds from Config, so a test
// can fake the exchange without also faking the APIs under audit.
func WithSTSEndpoint(endpoint string) Option {
	return func(o *options) {
		o.stsEndpoint = endpoint
	}
}

// WithHTTPClient overrides the HTTP client every AWS client built from this
// session uses.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(o *options) {
		o.httpClient = httpClient
	}
}

// NewSession exchanges a freshly minted identity federation token for temporary
// credentials on roleARN, and reports the account they belong to.
//
// The organization must come from the connector row being serviced, never from
// user input: it selects which customer's cloud this session can reach.
//
// The exchange is eager, so a deleted role or a trust policy that no longer
// names this organization fails here rather than midway through a job.
func NewSession(
	ctx context.Context,
	issuer *identityfederation.Issuer,
	organizationID gid.GID,
	roleARN string,
	opts ...Option,
) (*Session, error) {
	if issuer == nil {
		return nil, fmt.Errorf("cannot create aws session: identity federation issuer is required")
	}

	if roleARN == "" {
		return nil, fmt.Errorf("cannot create aws session: role ARN is required")
	}

	o := options{region: DefaultRegion}
	for _, opt := range opts {
		opt(&o)
	}

	if o.httpClient == nil {
		o.httpClient = httpclient.DefaultPooledClient()
	}

	cfg := awssdk.Config{
		Region:     o.region,
		HTTPClient: o.httpClient,
	}

	// AssumeRoleWithWebIdentity is unsigned, so this client needs no credentials
	// — the point of the whole design. Built without config.LoadDefaultConfig,
	// whose ambient environment and profile credentials would otherwise stand in
	// for the federated identity and audit the wrong account.
	stsClient := sts.NewFromConfig(
		cfg,
		func(stsOptions *sts.Options) {
			if o.stsEndpoint != "" {
				stsOptions.BaseEndpoint = new(o.stsEndpoint)
			}
		},
	)

	// The SDK owns caching and refresh from here: the retriever runs again only
	// when the credentials it obtained are about to expire.
	cfg.Credentials = awssdk.NewCredentialsCache(
		stscreds.NewWebIdentityRoleProvider(
			stsClient,
			roleARN,
			&issuerTokenRetriever{issuer: issuer, organizationID: organizationID},
			func(webIdentityOptions *stscreds.WebIdentityRoleOptions) {
				webIdentityOptions.RoleSessionName = roleSessionName
			},
		),
	)

	credentials, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot assume role with web identity: %w", err)
	}

	// STS reports the assumed-role user, whose ARN names the account reached. The
	// role ARN names the same account, so it covers an endpoint that omits it.
	accountID := credentials.AccountID
	if accountID == "" {
		accountID = accountIDFromARN(roleARN)
	}

	return &Session{cfg: cfg, accountID: accountID}, nil
}

func (s *Session) Cloud() string {
	return cloud.AWS
}

func (s *Session) AccountID() string {
	return s.accountID
}

// Config carries the federated credentials, so every client a driver builds from
// it acts as the customer's read-only audit role.
func (s *Session) Config() awssdk.Config {
	return s.cfg
}

// issuerTokenRetriever hands the SDK a token minted in this process, avoiding the
// usual WebIdentityTokenFile path: a token never written cannot be read by
// anything else on the host.
type issuerTokenRetriever struct {
	issuer         *identityfederation.Issuer
	organizationID gid.GID
}

var _ stscreds.IdentityTokenRetriever = (*issuerTokenRetriever)(nil)

// GetIdentityToken mints a token for this retriever's organization. The SDK's
// interface carries no context; minting performs no I/O, so there is nothing to
// cancel. The token is a bearer credential for the customer's account: never log
// it, and keep it out of errors.
func (r *issuerTokenRetriever) GetIdentityToken() ([]byte, error) {
	token, err := r.issuer.Token(context.Background(), r.organizationID, identityfederation.AudienceAWS)
	if err != nil {
		return nil, fmt.Errorf("cannot mint identity federation token: %w", err)
	}

	return []byte(token), nil
}

// accountIDFromARN returns the empty string for a value not shaped like an ARN.
func accountIDFromARN(arn string) string {
	fields := strings.Split(arn, ":")
	if len(fields) <= arnAccountIDField {
		return ""
	}

	return fields[arnAccountIDField]
}
