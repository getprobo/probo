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

package connector

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.gearno.de/kit/httpclient"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/statelesstoken"
)

type (
	GitHubAppConnector struct {
		AppID       string
		Slug        string
		PrivateKey  string
		InstallBase string
		APIBase     string
		HTTPClient  *http.Client
	}

	GitHubAppState struct {
		OrganizationID string `json:"oid"`
		ContinueURL    string `json:"continue,omitempty"`
		ConnectorID    string `json:"cid,omitempty"`
		Organization   string `json:"-"`
	}

	GitHubAppConnection struct {
		AppID          string `json:"app_id"`
		PrivateKey     string `json:"private_key"`
		InstallationID int64  `json:"installation_id"`
		APIBase        string `json:"api_base"`

		mu          sync.Mutex
		accessToken string
		expiresAt   time.Time
	}

	gitHubAppInstallation struct {
		TargetType string `json:"target_type"`
		Account    struct {
			Login string `json:"login"`
		} `json:"account"`
	}

	gitHubAppInstallationToken struct {
		Token     string    `json:"token"`
		ExpiresAt time.Time `json:"expires_at"`
	}

	gitHubAppTransport struct {
		connection *GitHubAppConnection
		underlying http.RoundTripper
	}
)

const (
	ProtocolGitHubApp ProtocolType = "GITHUB_APP"
	GitHubProvider                 = "GITHUB"

	gitHubAppStateType = "probo/connector/github-app"
	gitHubAppStateTTL  = 10 * time.Minute
)

var (
	_ Connector  = (*GitHubAppConnector)(nil)
	_ Connection = (*GitHubAppConnection)(nil)
)

func IsGitHubAppState(token string) bool {
	payload, err := statelesstoken.DecodePayload[json.RawMessage](token)
	return err == nil && payload.Type == gitHubAppStateType
}

func (c *GitHubAppConnector) Initiate(
	ctx context.Context,
	_ string,
	organizationID gid.GID,
	opts InitiateOptions,
	r *http.Request,
) (string, error) {
	stateData := GitHubAppState{
		OrganizationID: organizationID.String(),
		ConnectorID:    opts.ConnectorID,
	}
	if r != nil {
		stateData.ContinueURL = r.URL.Query().Get("continue")
	}

	state, err := statelesstoken.NewToken(c.PrivateKey, gitHubAppStateType, gitHubAppStateTTL, stateData)
	if err != nil {
		return "", fmt.Errorf("cannot create github app state token: %w", err)
	}

	installURL, err := url.JoinPath(
		c.InstallBase,
		url.PathEscape(c.Slug),
		"installations/new",
	)
	if err != nil {
		return "", fmt.Errorf("cannot build github app installation URL: %w", err)
	}

	u, err := url.Parse(installURL)
	if err != nil {
		return "", fmt.Errorf("cannot parse github app installation URL: %w", err)
	}

	q := u.Query()
	q.Set("state", state)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func (c *GitHubAppConnector) Complete(
	ctx context.Context,
	r *http.Request,
) (Connection, *gid.GID, string, error) {
	connection, state, err := c.CompleteWithState(ctx, r)
	if err != nil {
		return nil, nil, "", err
	}

	organizationID, err := gid.ParseGID(state.OrganizationID)
	if err != nil {
		return nil, nil, "", fmt.Errorf("cannot parse organization ID: %w", err)
	}

	return connection, &organizationID, state.ContinueURL, nil
}

func (c *GitHubAppConnector) CompleteWithState(
	ctx context.Context,
	r *http.Request,
) (*GitHubAppConnection, *GitHubAppState, error) {
	stateToken := r.URL.Query().Get("state")
	if stateToken == "" {
		return nil, nil, fmt.Errorf("cannot complete github app installation: missing state")
	}

	payload, err := statelesstoken.ValidateToken[GitHubAppState](
		c.PrivateKey,
		gitHubAppStateType,
		stateToken,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot validate github app state token: %w", err)
	}

	setupAction := r.URL.Query().Get("setup_action")
	if setupAction != "install" && setupAction != "update" {
		return nil, nil, fmt.Errorf("cannot complete github app installation: invalid setup action")
	}

	installationID, err := strconv.ParseInt(r.URL.Query().Get("installation_id"), 10, 64)
	if err != nil || installationID <= 0 {
		return nil, nil, fmt.Errorf("cannot complete github app installation: invalid installation ID")
	}

	apiBase := c.APIBase
	if apiBase == "" {
		return nil, nil, fmt.Errorf("cannot complete github app installation: missing API base URL")
	}

	connection := &GitHubAppConnection{
		AppID:          c.AppID,
		PrivateKey:     c.PrivateKey,
		InstallationID: installationID,
		APIBase:        apiBase,
	}

	organization, err := c.fetchInstallationOrganization(ctx, connection)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot fetch github app installation: %w", err)
	}

	payload.Data.Organization = organization

	return connection, &payload.Data, nil
}

func (c *GitHubAppConnector) fetchInstallationOrganization(
	ctx context.Context,
	connection *GitHubAppConnection,
) (string, error) {
	endpoint, err := url.JoinPath(
		connection.APIBase,
		"app/installations",
		strconv.FormatInt(connection.InstallationID, 10),
	)
	if err != nil {
		return "", fmt.Errorf("cannot build installation URL: %w", err)
	}

	jwt, err := connection.appJWT(time.Now())
	if err != nil {
		return "", fmt.Errorf("cannot create app JWT: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create installation request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Transport: httpclient.DefaultPooledTransport(httpclient.WithSSRFProtection()),
		}
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute installation request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("installation response status: %d", resp.StatusCode)
	}

	var installation gitHubAppInstallation
	if err := json.NewDecoder(resp.Body).Decode(&installation); err != nil {
		return "", fmt.Errorf("cannot decode installation response: %w", err)
	}
	if installation.Account.Login == "" {
		return "", fmt.Errorf("installation response has no account login")
	}
	if installation.TargetType != "Organization" {
		return "", fmt.Errorf("github app must be installed on an organization")
	}

	return installation.Account.Login, nil
}

func (c *GitHubAppConnection) Type() ProtocolType {
	return ProtocolGitHubApp
}

func (c *GitHubAppConnection) Scopes() []string {
	return []string{}
}

func (c *GitHubAppConnection) Client(ctx context.Context) (*http.Client, error) {
	if _, err := c.privateKey(); err != nil {
		return nil, fmt.Errorf("cannot create github app client: %w", err)
	}

	transport := httpclient.DefaultPooledTransport(httpclient.WithSSRFProtection())

	return &http.Client{
		Transport: &gitHubAppTransport{
			connection: c,
			underlying: transport,
		},
	}, nil
}

func (c GitHubAppConnection) MarshalJSON() ([]byte, error) {
	type Alias GitHubAppConnection

	return json.Marshal(
		&struct {
			Type string `json:"type"`
			Alias
		}{
			Type:  string(ProtocolGitHubApp),
			Alias: Alias(c),
		},
	)
}

func (c *GitHubAppConnection) UnmarshalJSON(data []byte) error {
	type Alias GitHubAppConnection
	return json.Unmarshal(data, (*Alias)(c))
}

func (c *GitHubAppConnection) installationToken(ctx context.Context, underlying http.RoundTripper) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.accessToken != "" && time.Until(c.expiresAt) > time.Minute {
		return c.accessToken, nil
	}

	jwt, err := c.appJWT(time.Now())
	if err != nil {
		return "", fmt.Errorf("cannot create app JWT: %w", err)
	}

	endpoint, err := url.JoinPath(
		c.APIBase,
		"app/installations",
		strconv.FormatInt(c.InstallationID, 10),
		"access_tokens",
	)
	if err != nil {
		return "", fmt.Errorf("cannot build installation token URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(nil))
	if err != nil {
		return "", fmt.Errorf("cannot create installation token request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := underlying.RoundTrip(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute installation token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", fmt.Errorf(
			"installation token response status: %d: %s",
			resp.StatusCode,
			strings.TrimSpace(string(body)),
		)
	}

	var token gitHubAppInstallationToken
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return "", fmt.Errorf("cannot decode installation token response: %w", err)
	}
	if token.Token == "" {
		return "", fmt.Errorf("installation token response has no token")
	}

	c.accessToken = token.Token
	c.expiresAt = token.ExpiresAt

	return token.Token, nil
}

func (c *GitHubAppConnection) appJWT(now time.Time) (string, error) {
	privateKey, err := c.privateKey()
	if err != nil {
		return "", err
	}

	header, err := json.Marshal(map[string]string{"alg": "RS256", "typ": "JWT"})
	if err != nil {
		return "", fmt.Errorf("cannot marshal JWT header: %w", err)
	}
	claims, err := json.Marshal(map[string]any{
		"iat": now.Add(-time.Minute).Unix(),
		"exp": now.Add(9 * time.Minute).Unix(),
		"iss": c.AppID,
	})
	if err != nil {
		return "", fmt.Errorf("cannot marshal JWT claims: %w", err)
	}

	unsigned := base64.RawURLEncoding.EncodeToString(header) + "." +
		base64.RawURLEncoding.EncodeToString(claims)
	digest := sha256.Sum256([]byte(unsigned))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest[:])
	if err != nil {
		return "", fmt.Errorf("cannot sign JWT: %w", err)
	}

	return unsigned + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func (c *GitHubAppConnection) privateKey() (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(c.PrivateKey))
	if block == nil {
		return nil, fmt.Errorf("cannot decode private key PEM")
	}

	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("cannot parse private key: %w", err)
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not RSA")
	}

	return rsaKey, nil
}

func (t *gitHubAppTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	token, err := t.connection.installationToken(req.Context(), t.underlying)
	if err != nil {
		return nil, err
	}

	cloned := req.Clone(req.Context())
	cloned.Header.Set("Authorization", "Bearer "+token)
	cloned.Header.Set("Accept", "application/vnd.github+json")
	cloned.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	return t.underlying.RoundTrip(cloned)
}
