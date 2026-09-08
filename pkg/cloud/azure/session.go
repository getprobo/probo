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

// Package azure vends authenticated access to one customer Azure subscription
// by presenting a Probo-minted OIDC token as an Entra client assertion.
//
// Probo holds no Azure credential for any customer. It mints a short-lived
// assertion (pkg/identityfederation), presents it at the environment's
// authority, and azidentity caches the resulting tokens per scope. The
// customer revokes by deleting the federated identity credential.
package azure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"go.gearno.de/kit/httpclient"
	"go.probo.inc/probo/pkg/cloud"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/identityfederation"
)

const entraFederatedCredentialNotFound = 70021

type (
	// Session is authenticated access to one Azure subscription, held as a
	// client-assertion credential whose tokens refresh themselves through WIF.
	// Build ARM clients from ARMClientOptions; talk to Graph through GraphClient.
	Session struct {
		httpClient     *http.Client
		graphClient    *GraphClient
		credential     azcore.TokenCredential
		clientOptions  azcore.ClientOptions
		tenantID       string
		subscriptionID string
		environment    Environment
		graphBaseURL   string
		graphScope     string
		armScope       string
		after          func(time.Duration) <-chan time.Time
	}

	// GraphClient attaches a Graph access token on Do. Redirects are
	// followed by net/http, which strips Authorization when the host
	// changes.
	GraphClient struct {
		client  *http.Client
		session *Session
	}

	staticTokenCredential struct {
		token string
	}

	// SessionOption configures a session built by NewSessionFromToken.
	SessionOption func(*Session)
)

var (
	_ cloud.Session          = (*Session)(nil)
	_ azcore.TokenCredential = staticTokenCredential{}

	federatedCredentialBackoffs = []time.Duration{3 * time.Second, 7 * time.Second}
)

// NewSession opens a session on subscriptionID in the given environment, by
// exchanging an assertion minted for organizationID.
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
	tenantID string,
	clientID string,
	subscriptionID string,
	environment Environment,
) (*Session, error) {
	tenantID, err := parseGUID(tenantID, errInvalidTenantID)
	if err != nil {
		return nil, fmt.Errorf("cannot open azure session: %w", err)
	}

	clientID, err = parseGUID(clientID, errInvalidClientID)
	if err != nil {
		return nil, fmt.Errorf("cannot open azure session: %w", err)
	}

	subscriptionID, err = parseGUID(subscriptionID, errInvalidSubscriptionID)
	if err != nil {
		return nil, fmt.Errorf("cannot open azure session: %w", err)
	}

	if environment == "" {
		environment = EnvironmentPublic
	}

	ep, err := environment.endpoints()
	if err != nil {
		return nil, fmt.Errorf("cannot open azure session: %w", err)
	}

	httpClient := httpclient.DefaultPooledClient(httpclient.WithSSRFProtection())
	opts := azcore.ClientOptions{
		Cloud:     ep.configuration,
		Transport: httpClient,
	}

	getAssertion := func(ctx context.Context) (string, error) {
		token, err := issuer.Token(ctx, organizationID, identityfederation.AudienceAzure)
		if err != nil {
			return "", fmt.Errorf("cannot mint azure identity federation token: %w", err)
		}

		return token, nil
	}

	cred, err := azidentity.NewClientAssertionCredential(
		tenantID,
		clientID,
		getAssertion,
		&azidentity.ClientAssertionCredentialOptions{ClientOptions: opts},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot open azure session: %w", err)
	}

	session := &Session{
		httpClient:     httpClient,
		credential:     cred,
		clientOptions:  opts,
		tenantID:       tenantID,
		subscriptionID: subscriptionID,
		environment:    environment,
		graphBaseURL:   ep.graphBaseURL,
		graphScope:     ep.graphScope,
		armScope:       ep.armScope,
	}
	session.graphClient = authorizeGraph(httpClient, session)

	return session, nil
}

// WithHTTPClient sets the transport the Graph client wraps and azidentity
// uses for the token exchange. Tests pass a VCR client. Omit it to use the
// default SSRF-protected pool.
func WithHTTPClient(httpClient *http.Client) SessionOption {
	return func(s *Session) {
		s.httpClient = httpClient
	}
}

// WithEnvironment selects the Azure cloud the session dials. Omit it to use
// the public cloud.
func WithEnvironment(environment Environment) SessionOption {
	return func(s *Session) {
		s.environment = environment
	}
}

// NewSessionFromToken builds a session from an already-issued access token.
// Production uses NewSession, which obtains credentials through WIF.
func NewSessionFromToken(subscriptionID, accessToken string, opts ...SessionOption) *Session {
	session := &Session{
		httpClient:     httpclient.DefaultPooledClient(httpclient.WithSSRFProtection()),
		credential:     staticTokenCredential{token: accessToken},
		subscriptionID: subscriptionID,
		environment:    EnvironmentPublic,
	}

	for _, opt := range opts {
		opt(session)
	}

	if err := session.applyEnvironment(); err != nil {
		session.environment = EnvironmentPublic
		_ = session.applyEnvironment()
	}

	session.graphClient = authorizeGraph(session.httpClient, session)

	return session
}

func (s *Session) applyEnvironment() error {
	ep, err := s.environment.endpoints()
	if err != nil {
		return err
	}

	s.graphBaseURL = ep.graphBaseURL
	s.graphScope = ep.graphScope
	s.armScope = ep.armScope
	s.clientOptions = azcore.ClientOptions{
		Cloud:     ep.configuration,
		Transport: s.httpClient,
	}

	return nil
}

// Cloud implements cloud.Session.
func (s *Session) Cloud() string {
	return cloud.Azure
}

// AccountID is the Azure subscription this session reviews.
func (s *Session) AccountID() string {
	return s.subscriptionID
}

// TenantID is the Entra tenant that issued the federated credential.
func (s *Session) TenantID() string {
	return s.tenantID
}

// Environment is the Azure cloud this session dials.
func (s *Session) Environment() Environment {
	return s.environment
}

// ARMClientOptions are the client options every ARM SDK client on this
// session must use. azcore injects the Resource Manager host from Cloud;
// it does not know about Graph.
func (s *Session) ARMClientOptions() *arm.ClientOptions {
	return &arm.ClientOptions{ClientOptions: s.clientOptions}
}

// GraphBaseURL is the Microsoft Graph host for this session's environment.
// Every Graph URL must be built from this value with url.JoinPath.
func (s *Session) GraphBaseURL() string {
	return s.graphBaseURL
}

// GraphClient is the SSRF-protected client that attaches a Graph bearer
// token on each Do. Token refresh uses the request context.
func (s *Session) GraphClient() *GraphClient {
	return s.graphClient
}

// CheckAccess reports whether this session can actually reach its subscription.
//
// It acquires an ARM token and stops. Every federated principal that Entra
// accepts can do that, so a failure means the exchange itself was refused — a
// missing federated credential, a subject that does not name this
// organization, an unpublished signing key — and never a missing Reader
// assignment. That is what makes it a connection check rather than a
// capability check.
//
// Entra documents that a token request made minutes after creating a
// federated identity credential can fail with AADSTS70021. CheckAccess
// retries that code a small number of times over roughly ten seconds. Any
// other Entra error fails immediately.
//
// It is also the first call to force the exchange: NewSession fetches no
// credential, so until something talks to Entra there is nothing to be wrong.
func (s *Session) CheckAccess(ctx context.Context) error {
	var err error
	for attempt := 0; ; attempt++ {
		_, err = s.credential.GetToken(
			ctx,
			policy.TokenRequestOptions{Scopes: []string{s.armScope}},
		)
		if err == nil {
			return nil
		}

		if ctx.Err() != nil {
			return fmt.Errorf("cannot reach azure subscription: %w", ctx.Err())
		}

		if !isFederatedCredentialNotFound(err) || attempt >= len(federatedCredentialBackoffs) {
			return fmt.Errorf("cannot reach azure subscription: %w", err)
		}

		if err := s.wait(ctx, federatedCredentialBackoffs[attempt]); err != nil {
			return fmt.Errorf("cannot reach azure subscription: %w", err)
		}
	}
}

func (s *Session) wait(ctx context.Context, d time.Duration) error {
	var timerC <-chan time.Time
	stop := func() {}
	if s.after != nil {
		timerC = s.after(d)
	} else {
		timer := time.NewTimer(d)
		timerC = timer.C
		stop = func() { timer.Stop() }
	}

	select {
	case <-ctx.Done():
		stop()

		return ctx.Err()
	case <-timerC:
		return nil
	}
}

func isFederatedCredentialNotFound(err error) bool {
	authErr, ok := errors.AsType[*azidentity.AuthenticationFailedError](err)
	if !ok || authErr.RawResponse == nil {
		return false
	}

	body, readErr := runtime.Payload(authErr.RawResponse)
	if readErr != nil {
		return false
	}

	var payload struct {
		ErrorCodes []int `json:"error_codes"`
	}
	if json.Unmarshal(body, &payload) != nil {
		return false
	}

	return slices.Contains(payload.ErrorCodes, entraFederatedCredentialNotFound)
}

func authorizeGraph(base *http.Client, session *Session) *GraphClient {
	client := *base

	return &GraphClient{
		client:  &client,
		session: session,
	}
}

// Do sends req with a Graph bearer token. Token refresh uses req.Context.
func (c *GraphClient) Do(req *http.Request) (*http.Response, error) {
	token, err := c.session.credential.GetToken(
		req.Context(),
		policy.TokenRequestOptions{Scopes: []string{c.session.graphScope}},
	)
	if err != nil {
		if req.Body != nil {
			_ = req.Body.Close()
		}

		return nil, err
	}

	authorized := req.Clone(req.Context())
	authorized.Header.Set("Authorization", "Bearer "+token.Token)

	return c.client.Do(authorized)
}

func (c staticTokenCredential) GetToken(
	_ context.Context,
	_ policy.TokenRequestOptions,
) (azcore.AccessToken, error) {
	return azcore.AccessToken{
		Token:     c.token,
		ExpiresOn: time.Now().Add(time.Hour),
	}, nil
}
