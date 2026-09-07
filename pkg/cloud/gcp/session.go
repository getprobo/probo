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

// Package gcp vends a one-hour authenticated session on one customer GCP
// project by exchanging a Probo-minted OIDC token and impersonating the
// customer's audit service account.
//
// Probo holds no GCP credential for any customer. It mints a short-lived
// assertion (pkg/identityfederation), exchanges it at STS in the project's
// universe, and impersonates the customer's service account. The customer
// revokes by deleting the workload identity binding.
package gcp

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"go.gearno.de/kit/httpclient"
	"go.probo.inc/probo/pkg/cloud"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/identityfederation"
	"golang.org/x/oauth2"
	"google.golang.org/api/iamcredentials/v1"
	"google.golang.org/api/option"
	"google.golang.org/api/sts/v1"
)

const (
	tokenExchangeGrantType = "urn:ietf:params:oauth:grant-type:token-exchange"
	idTokenType            = "urn:ietf:params:oauth:token-type:id_token"
	accessTokenType        = "urn:ietf:params:oauth:token-type:access_token"
	cloudPlatformScope     = "https://www.googleapis.com/auth/cloud-platform"
	// Admin Directory is a Workspace API; cloud-platform does not cover it.
	adminDirectoryUserReadonlyScope = "https://www.googleapis.com/auth/admin.directory.user.readonly"
	serviceAccountLifetime          = "3600s"
)

type (
	// Session is authenticated access to one GCP project, held as an HTTP
	// client whose token refreshes itself through WIF and impersonation.
	// Build service clients with HTTPClient; do not copy the token out.
	Session struct {
		httpClient             *http.Client
		authorizedClient       *http.Client
		mu                     sync.Mutex
		token                  *oauth2.Token
		issuer                 *identityfederation.Issuer
		organizationID         gid.GID
		provider               providerResource
		serviceAccountEmail    string
		accountID              string
		universeDomain         string
		stsEndpoint            string
		iamCredentialsEndpoint string
	}

	sessionTransport struct {
		session *Session
		base    http.RoundTripper
	}

	// SessionOption configures a session built by NewSessionFromToken.
	SessionOption func(*Session)
)

var (
	_ cloud.Session     = (*Session)(nil)
	_ http.RoundTripper = (*sessionTransport)(nil)

	serviceAccountScopes = []string{
		cloudPlatformScope,
		adminDirectoryUserReadonlyScope,
	}
)

// NewSession opens a session on the project named by providerResource, by
// exchanging an assertion minted for organizationID and impersonating
// serviceAccountEmail.
//
// organizationID must come from the connector row being serviced, never from
// user input: it selects whose cloud accounts the resulting credentials reach.
//
// No credential is fetched here, which is why there is no context to pass. The
// first API call performs the exchange and impersonation and owns caching and
// refresh from then on, so a session is cheap to build and no token is ever
// written to disk.
func NewSession(
	issuer *identityfederation.Issuer,
	organizationID gid.GID,
	providerResource string,
	serviceAccountEmail string,
) (*Session, error) {
	parsed, err := parseProviderResource(providerResource)
	if err != nil {
		return nil, fmt.Errorf("cannot open gcp session: %w", err)
	}

	email, err := parseServiceAccountEmail(serviceAccountEmail)
	if err != nil {
		return nil, fmt.Errorf("cannot open gcp session: %w", err)
	}

	universe, err := inferUniverse(parsed, email)
	if err != nil {
		return nil, fmt.Errorf("cannot open gcp session: %w", err)
	}

	httpClient := httpclient.DefaultPooledClient(httpclient.WithSSRFProtection())

	session := &Session{
		httpClient:          httpClient,
		issuer:              issuer,
		organizationID:      organizationID,
		provider:            parsed,
		serviceAccountEmail: email,
		accountID:           parsed.projectNumber,
		universeDomain:      universe,
	}
	session.authorizedClient = authorizeSession(httpClient, session)

	return session, nil
}

// WithHTTPClient sets the transport the authorized client wraps. Tests pass a
// VCR client. Omit it to use the default SSRF-protected pool.
func WithHTTPClient(httpClient *http.Client) SessionOption {
	return func(s *Session) {
		s.httpClient = httpClient
	}
}

// NewSessionFromToken builds a session from an already-issued access token.
// Production uses NewSession, which obtains credentials through WIF.
func NewSessionFromToken(projectNumber, accessToken string, opts ...SessionOption) *Session {
	session := &Session{
		httpClient:     httpclient.DefaultPooledClient(httpclient.WithSSRFProtection()),
		token:          &oauth2.Token{AccessToken: accessToken},
		accountID:      projectNumber,
		universeDomain: CommercialUniverse,
	}

	for _, opt := range opts {
		opt(session)
	}

	session.authorizedClient = authorizeSession(session.httpClient, session)

	return session
}

// Cloud implements cloud.Session.
func (s *Session) Cloud() string {
	return cloud.GCP
}

// AccountID is the GCP project number the workload identity provider lives in.
func (s *Session) AccountID() string {
	return s.accountID
}

// HTTPClient is the SSRF-protected client with the impersonated token already
// attached. Pass it to option.WithHTTPClient. Token refresh uses the request
// context.
func (s *Session) HTTPClient() *http.Client {
	return s.authorizedClient
}

// UniverseDomain is the Google Cloud universe this session dials:
// googleapis.com (commercial) or s3nsapis.fr (S3NS). JWT audiences stay on
// iam.googleapis.com in every universe; only HTTP hosts change.
func (s *Session) UniverseDomain() string {
	if s.universeDomain == "" {
		return CommercialUniverse
	}

	return s.universeDomain
}

// ServiceOptions are the client options every Google API client on this
// session must use: the impersonated HTTP client, no ADC, and the universe
// domain so hosts resolve to this universe rather than always public GCP.
func (s *Session) ServiceOptions() []option.ClientOption {
	return s.serviceOptions(s.HTTPClient())
}

func (s *Session) serviceOptions(httpClient *http.Client) []option.ClientOption {
	return []option.ClientOption{
		option.WithHTTPClient(httpClient),
		option.WithoutAuthentication(),
		option.WithUniverseDomain(s.UniverseDomain()),
	}
}

// CheckAccess reports whether this session can actually reach its project.
//
// It completes the STS exchange and service-account impersonation. Every
// federated principal granted workloadIdentityUser on the service account can
// do that, so a failure means the exchange itself was refused — a missing
// pool, an attribute condition that does not name this organization, an
// unpublished signing key — and never a missing IAM role on the service
// account. That is what makes it a connection check rather than a capability
// check.
//
// It is also the first call to force the exchange: NewSession fetches no
// credential, so until something talks to GCP there is nothing to be wrong.
func (s *Session) CheckAccess(ctx context.Context) error {
	_, err := s.accessToken(ctx)
	if err != nil {
		return fmt.Errorf("cannot reach gcp project: %w", err)
	}

	return nil
}

func (s *Session) accessToken(ctx context.Context) (*oauth2.Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.token.Valid() {
		return s.token, nil
	}

	token, err := s.fetchServiceAccountToken(ctx)
	if err != nil {
		return nil, err
	}

	s.token = token

	return token, nil
}

func (s *Session) fetchServiceAccountToken(ctx context.Context) (*oauth2.Token, error) {
	assertion, err := s.mintAssertion(ctx)
	if err != nil {
		return nil, err
	}

	federated, err := s.exchangeAssertion(ctx, assertion)
	if err != nil {
		return nil, err
	}

	return s.impersonate(ctx, federated)
}

func (s *Session) mintAssertion(ctx context.Context) (string, error) {
	audience, err := jwtAudience(s.provider)
	if err != nil {
		return "", fmt.Errorf("cannot build gcp identity federation audience: %w", err)
	}

	token, err := s.issuer.Token(ctx, s.organizationID, audience)
	if err != nil {
		return "", fmt.Errorf("cannot mint gcp identity federation token: %w", err)
	}

	return token, nil
}

func (s *Session) exchangeAssertion(ctx context.Context, assertion string) (*oauth2.Token, error) {
	opts := s.serviceOptions(s.httpClient)
	if s.stsEndpoint != "" {
		opts = append(opts, option.WithEndpoint(s.stsEndpoint))
	}

	svc, err := sts.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot create gcp sts client: %w", err)
	}

	audience, err := stsAudience(s.provider)
	if err != nil {
		return nil, fmt.Errorf("cannot build gcp sts audience: %w", err)
	}

	resp, err := svc.V1.Token(
		&sts.GoogleIdentityStsV1ExchangeTokenRequest{
			Audience:           audience,
			GrantType:          tokenExchangeGrantType,
			RequestedTokenType: accessTokenType,
			Scope:              cloudPlatformScope,
			SubjectToken:       assertion,
			SubjectTokenType:   idTokenType,
		},
	).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("cannot exchange gcp identity federation token: %w", err)
	}

	return &oauth2.Token{
		AccessToken: resp.AccessToken,
		TokenType:   resp.TokenType,
		Expiry:      time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second),
	}, nil
}

func (s *Session) impersonate(ctx context.Context, federated *oauth2.Token) (*oauth2.Token, error) {
	name, err := serviceAccountResourceName(s.serviceAccountEmail)
	if err != nil {
		return nil, err
	}

	opts := s.serviceOptions(authorizeClient(s.httpClient, oauth2.StaticTokenSource(federated)))
	if s.iamCredentialsEndpoint != "" {
		opts = append(opts, option.WithEndpoint(s.iamCredentialsEndpoint))
	}

	svc, err := iamcredentials.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot create gcp iam credentials client: %w", err)
	}

	resp, err := svc.Projects.ServiceAccounts.GenerateAccessToken(
		name,
		&iamcredentials.GenerateAccessTokenRequest{
			Scope:    serviceAccountScopes,
			Lifetime: serviceAccountLifetime,
		},
	).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("cannot impersonate gcp service account: %w", err)
	}

	expiry, err := time.Parse(time.RFC3339, resp.ExpireTime)
	if err != nil {
		return nil, fmt.Errorf("cannot parse gcp service account token expiry: %w", err)
	}

	return &oauth2.Token{
		AccessToken: resp.AccessToken,
		TokenType:   "Bearer",
		Expiry:      expiry,
	}, nil
}

func jwtAudience(p providerResource) (string, error) {
	path, err := iamResourcePath(p)
	if err != nil {
		return "", err
	}

	return new(url.URL{
		Scheme: "https",
		Host:   iamHost,
		Path:   path,
	}).String(), nil
}

func stsAudience(p providerResource) (string, error) {
	path, err := iamResourcePath(p)
	if err != nil {
		return "", err
	}

	return new(url.URL{
		Host: iamHost,
		Path: path,
	}).String(), nil
}

func iamResourcePath(p providerResource) (string, error) {
	path, err := url.JoinPath(
		"/",
		"projects",
		url.PathEscape(p.projectNumber),
		"locations",
		"global",
		"workloadIdentityPools",
		url.PathEscape(p.poolID),
		"providers",
		url.PathEscape(p.providerID),
	)
	if err != nil {
		return "", fmt.Errorf("cannot build gcp iam resource path: %w", err)
	}

	return path, nil
}

func serviceAccountResourceName(email string) (string, error) {
	name, err := url.JoinPath("projects/-/serviceAccounts", url.PathEscape(email))
	if err != nil {
		return "", fmt.Errorf("cannot build service account resource name: %w", err)
	}

	return name, nil
}

func authorizeSession(base *http.Client, session *Session) *http.Client {
	client := *base
	client.Transport = &sessionTransport{
		session: session,
		base:    base.Transport,
	}

	return &client
}

func (t *sessionTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	token, err := t.session.accessToken(req.Context())
	if err != nil {
		if req.Body != nil {
			_ = req.Body.Close()
		}

		return nil, err
	}

	authorized := req.Clone(req.Context())
	token.SetAuthHeader(authorized)

	return t.base.RoundTrip(authorized)
}

func authorizeClient(base *http.Client, src oauth2.TokenSource) *http.Client {
	client := *base
	client.Transport = &oauth2.Transport{
		Source: src,
		Base:   base.Transport,
	}

	return &client
}
