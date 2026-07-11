// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package drivers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"go.probo.inc/probo/pkg/connector"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

// NameResolver fetches the human-readable instance name from a provider
// (e.g. Slack workspace name, Google Workspace domain).
type NameResolver interface {
	ResolveInstanceName(ctx context.Context) (string, error)
}

// slackNameResolver resolves the Slack workspace name via auth.test.
type slackNameResolver struct {
	httpClient *http.Client
}

func NewSlackNameResolver(httpClient *http.Client) NameResolver {
	return &slackNameResolver{httpClient: httpClient}
}

func (r *slackNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://slack.com/api/auth.test", nil)
	if err != nil {
		return "", fmt.Errorf("cannot create slack auth.test request: %w", err)
	}

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute slack auth.test request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	var resp struct {
		OK   bool   `json:"ok"`
		Team string `json:"team"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode slack auth.test response: %w", err)
	}

	if !resp.OK {
		return "", fmt.Errorf("slack auth.test returned ok=false")
	}

	return resp.Team, nil
}

// googleWorkspaceNameResolver resolves the Google Workspace primary domain.
type googleWorkspaceNameResolver struct {
	httpClient *http.Client
}

func NewGoogleWorkspaceNameResolver(httpClient *http.Client) NameResolver {
	return &googleWorkspaceNameResolver{httpClient: httpClient}
}

func (r *googleWorkspaceNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	adminService, err := admin.NewService(ctx, option.WithHTTPClient(r.httpClient))
	if err != nil {
		return "", fmt.Errorf("cannot create google admin service: %w", err)
	}

	customer, err := adminService.Customers.Get("my_customer").Context(ctx).Do()
	if err != nil {
		var gerr *googleapi.Error
		if errors.As(err, &gerr) && gerr.Code == http.StatusForbidden {
			return "", nil
		}

		return "", fmt.Errorf("cannot fetch google workspace customer: %w", err)
	}

	return customer.CustomerDomain, nil
}

// linearNameResolver resolves the Linear organization name via GraphQL.
type linearNameResolver struct {
	httpClient *http.Client
}

func NewLinearNameResolver(httpClient *http.Client) NameResolver {
	return &linearNameResolver{httpClient: httpClient}
}

func (r *linearNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	body := struct {
		Query string `json:"query"`
	}{
		Query: `{ organization { name } }`,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("cannot marshal linear organization query: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, linearGraphQLEndpoint, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("cannot create linear organization request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute linear organization request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", fmt.Errorf("cannot fetch linear organization: unexpected status %d", httpResp.StatusCode)
	}

	var resp struct {
		Data struct {
			Organization struct {
				Name string `json:"name"`
			} `json:"organization"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode linear organization response: %w", err)
	}

	if len(resp.Errors) > 0 {
		return "", fmt.Errorf("linear graphql error: %s", resp.Errors[0].Message)
	}

	return resp.Data.Organization.Name, nil
}

// cloudflareNameResolver resolves the Cloudflare account name.
type cloudflareNameResolver struct {
	httpClient *http.Client
}

func NewCloudflareNameResolver(httpClient *http.Client) NameResolver {
	return &cloudflareNameResolver{httpClient: httpClient}
}

func (r *cloudflareNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	cfURL, err := url.Parse("https://api.cloudflare.com/client/v4/accounts")
	if err != nil {
		return "", fmt.Errorf("cannot parse cloudflare accounts URL: %w", err)
	}

	q := cfURL.Query()
	q.Set("page", "1")
	q.Set("per_page", "1")
	cfURL.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfURL.String(), nil)
	if err != nil {
		return "", fmt.Errorf("cannot create cloudflare accounts request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute cloudflare accounts request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", fmt.Errorf("cannot fetch cloudflare accounts: unexpected status %d", httpResp.StatusCode)
	}

	var resp struct {
		Result []struct {
			Name string `json:"name"`
		} `json:"result"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode cloudflare accounts response: %w", err)
	}

	if len(resp.Result) == 0 {
		return "", fmt.Errorf("no cloudflare accounts found")
	}

	return resp.Result[0].Name, nil
}

// brexNameResolver resolves the Brex company name.
type brexNameResolver struct {
	httpClient *http.Client
}

func NewBrexNameResolver(httpClient *http.Client) NameResolver {
	return &brexNameResolver{httpClient: httpClient}
}

func (r *brexNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://platform.brexapis.com/v2/company",
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("cannot create brex company request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute brex company request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", fmt.Errorf("cannot fetch brex company: unexpected status %d", httpResp.StatusCode)
	}

	var resp struct {
		LegalName string `json:"legal_name"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode brex company response: %w", err)
	}

	return resp.LegalName, nil
}

// tallyNameResolver resolves the Tally organization name.
type tallyNameResolver struct {
	httpClient     *http.Client
	organizationID string
}

func NewTallyNameResolver(httpClient *http.Client, organizationID string) NameResolver {
	return &tallyNameResolver{
		httpClient:     httpClient,
		organizationID: organizationID,
	}
}

func (r *tallyNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	endpoint, err := url.JoinPath("https://api.tally.so", "organizations", url.PathEscape(r.organizationID))
	if err != nil {
		return "", fmt.Errorf("cannot build tally organization URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create tally organization request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute tally organization request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", fmt.Errorf("cannot fetch tally organization: unexpected status %d", httpResp.StatusCode)
	}

	var resp struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode tally organization response: %w", err)
	}

	return resp.Name, nil
}

// qoveryNameResolver resolves the Qovery organization name.
type qoveryNameResolver struct {
	httpClient     *http.Client
	organizationID string
}

func NewQoveryNameResolver(httpClient *http.Client, organizationID string) NameResolver {
	return &qoveryNameResolver{
		httpClient:     httpClient,
		organizationID: organizationID,
	}
}

func (r *qoveryNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	if r.organizationID == "" {
		return "", nil
	}

	endpoint, err := url.JoinPath(qoveryAPIBaseURL, "organization", url.PathEscape(r.organizationID))
	if err != nil {
		return "", fmt.Errorf("cannot build qovery organization URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create qovery organization request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute qovery organization request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	// Best-effort: a non-2xx (revoked token, deleted org, stale ID) must not
	// make the source-name worker retry forever. Give up gracefully and keep
	// the generic source name; a dead token surfaces on the next ListAccounts.
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", nil
	}

	var resp struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode qovery organization response: %w", err)
	}

	return resp.Name, nil
}

// renderNameResolver resolves the Render workspace (owner) name from
// GET /v1/owners/{ownerId}, used to title the AccessReviewSource "Render <name>".
type renderNameResolver struct {
	httpClient *http.Client
	ownerID    string
}

func NewRenderNameResolver(httpClient *http.Client, ownerID string) NameResolver {
	return &renderNameResolver{
		httpClient: httpClient,
		ownerID:    ownerID,
	}
}

func (r *renderNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	if r.ownerID == "" {
		return "", nil
	}

	endpoint, err := url.JoinPath(renderAPIBaseURL, "owners", url.PathEscape(r.ownerID))
	if err != nil {
		return "", fmt.Errorf("cannot build render owner URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create render owner request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute render owner request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	// Best-effort: a non-2xx (revoked token, deleted workspace, stale ID) must
	// not make the source-name worker retry forever. Give up gracefully and
	// keep the generic source name; a dead token surfaces on the next
	// ListAccounts.
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", nil
	}

	var resp struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode render owner response: %w", err)
	}

	return resp.Name, nil
}

// neonNameResolver resolves the Neon organization name.
type neonNameResolver struct {
	httpClient     *http.Client
	organizationID string
}

func NewNeonNameResolver(httpClient *http.Client, organizationID string) NameResolver {
	return &neonNameResolver{
		httpClient:     httpClient,
		organizationID: organizationID,
	}
}

func (r *neonNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	if r.organizationID == "" {
		return "", nil
	}

	endpoint, err := url.JoinPath(neonAPIBaseURL, "organizations", url.PathEscape(r.organizationID))
	if err != nil {
		return "", fmt.Errorf("cannot build neon organization URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create neon organization request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute neon organization request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	// Best-effort: a non-2xx (revoked key, deleted org, stale ID) must not
	// make the source-name worker retry forever. Give up gracefully and keep
	// the generic source name; a dead key surfaces on the next ListAccounts.
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", nil
	}

	var resp struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode neon organization response: %w", err)
	}

	return resp.Name, nil
}

// hubspotNameResolver resolves the HubSpot account name.
type hubspotNameResolver struct {
	httpClient *http.Client
}

func NewHubSpotNameResolver(httpClient *http.Client) NameResolver {
	return &hubspotNameResolver{httpClient: httpClient}
}

func (r *hubspotNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://api.hubapi.com/account-info/v3/details",
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("cannot create hubspot account-info request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute hubspot account-info request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", fmt.Errorf("cannot fetch hubspot account info: unexpected status %d", httpResp.StatusCode)
	}

	var resp struct {
		PortalID    int    `json:"portalId"`
		AccountName string `json:"accountName"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode hubspot account-info response: %w", err)
	}

	return resp.AccountName, nil
}

// docusignNameResolver resolves the configured DocuSign account's name from
// the OAuth2 userinfo endpoint, for the AccessReviewSource title.
type docusignNameResolver struct {
	httpClient *http.Client
	accountID  string
}

func NewDocuSignNameResolver(httpClient *http.Client, accountID string) NameResolver {
	return &docusignNameResolver{httpClient: httpClient, accountID: accountID}
}

func (r *docusignNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	accounts, err := fetchDocuSignAccounts(ctx, r.httpClient)
	if err != nil {
		// A dead token (non-2xx) is terminal: the source-name worker marks
		// the source synced on ("", nil), so it stops retrying. Transient and
		// decode failures stay errors so the worker retries.
		if errors.Is(err, errDocuSignUserInfoStatus) {
			return "", nil
		}

		return "", err
	}

	for _, account := range accounts {
		if account.AccountID == r.accountID {
			return account.AccountName, nil
		}
	}

	return "", nil
}

// openaiNameResolver resolves the OpenAI organization name.
type openaiNameResolver struct {
	httpClient *http.Client
}

func NewOpenAINameResolver(httpClient *http.Client) NameResolver {
	return &openaiNameResolver{httpClient: httpClient}
}

func (r *openaiNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://api.openai.com/v1/organization",
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("cannot create openai organization request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute openai organization request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		// OpenAI may not support this endpoint for all token types.
		return "", nil
	}

	var resp struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode openai organization response: %w", err)
	}

	return resp.Name, nil
}

// anthropicNameResolver resolves the Anthropic organization name via the
// Admin API /v1/organizations/me endpoint, which returns the org an
// admin key belongs to.
type anthropicNameResolver struct {
	httpClient *http.Client
}

func NewAnthropicNameResolver(httpClient *http.Client) NameResolver {
	return &anthropicNameResolver{httpClient: httpClient}
}

func (r *anthropicNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://api.anthropic.com/v1/organizations/me",
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("cannot create anthropic organization request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("anthropic-version", anthropicAPIVersion)

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute anthropic organization request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	// Best-effort: a non-2xx (e.g. a revoked admin key) must not make the
	// source-name worker retry forever. Give up gracefully and keep the
	// generic source name; a dead key surfaces on the next ListAccounts.
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", nil
	}

	var resp struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode anthropic organization response: %w", err)
	}

	return resp.Name, nil
}

// sendGridNameResolver resolves the SendGrid account's company name from
// the user profile endpoint, used as the AccessReviewSource instance label.
type sendGridNameResolver struct {
	httpClient *http.Client
}

func NewSendGridNameResolver(httpClient *http.Client) NameResolver {
	return &sendGridNameResolver{httpClient: httpClient}
}

func (r *sendGridNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://api.sendgrid.com/v3/user/profile",
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("cannot create sendgrid profile request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute sendgrid profile request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	// Best-effort: a non-2xx (revoked key, or a key without the
	// user.profile.read scope) must not make the source-name worker retry
	// forever. Give up gracefully and keep the generic source name; a dead
	// key surfaces on the next ListAccounts.
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", nil
	}

	var resp struct {
		Company string `json:"company"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode sendgrid profile response: %w", err)
	}

	return resp.Company, nil
}

// sentryNameResolver resolves the Sentry organization name.
type sentryNameResolver struct {
	httpClient *http.Client
	orgSlug    string
}

func NewSentryNameResolver(httpClient *http.Client, orgSlug string) NameResolver {
	return &sentryNameResolver{httpClient: httpClient, orgSlug: orgSlug}
}

func (r *sentryNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	if r.orgSlug == "" {
		return "", nil
	}

	endpoint, err := url.JoinPath("https://sentry.io", "api", "0", "organizations", url.PathEscape(r.orgSlug))
	if err != nil {
		return "", fmt.Errorf("cannot build sentry organization URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create sentry organization request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute sentry organization request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	// 404 means the stored slug is no longer visible to this token.
	// Treat as terminal so the worker stops looping; other non-2xx
	// stay retryable for token refresh / transient outages.
	if httpResp.StatusCode == http.StatusNotFound {
		return "", nil
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", fmt.Errorf("cannot fetch sentry organization: unexpected status %d", httpResp.StatusCode)
	}

	var resp struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode sentry organization response: %w", err)
	}

	return resp.Name, nil
}

// githubNameResolver resolves the GitHub organization name.
type githubNameResolver struct {
	httpClient *http.Client
	org        string
}

func NewGitHubNameResolver(httpClient *http.Client, org string) NameResolver {
	return &githubNameResolver{httpClient: httpClient, org: org}
}

func (r *githubNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	if r.org == "" {
		return "", nil
	}

	endpoint, err := url.JoinPath("https://api.github.com", "orgs", url.PathEscape(r.org))
	if err != nil {
		return "", fmt.Errorf("cannot build github organization URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create github organization request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute github organization request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", fmt.Errorf("cannot fetch github organization: unexpected status %d", httpResp.StatusCode)
	}

	var resp struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode github organization response: %w", err)
	}

	if resp.Name == "" {
		return r.org, nil
	}

	return resp.Name, nil
}

// supabaseNameResolver returns the Supabase organization slug as the name.
type supabaseNameResolver struct {
	orgSlug string
}

func NewSupabaseNameResolver(orgSlug string) NameResolver {
	return &supabaseNameResolver{orgSlug: orgSlug}
}

func (r *supabaseNameResolver) ResolveInstanceName(_ context.Context) (string, error) {
	return r.orgSlug, nil
}

// intercomNameResolver resolves the Intercom app name.
type intercomNameResolver struct {
	httpClient *http.Client
}

func NewIntercomNameResolver(httpClient *http.Client) NameResolver {
	return &intercomNameResolver{httpClient: httpClient}
}

func (r *intercomNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.intercom.io/me", nil)
	if err != nil {
		return "", fmt.Errorf("cannot create intercom me request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Intercom-Version", "2.11")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute intercom me request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", nil
	}

	var resp struct {
		App struct {
			Name string `json:"name"`
		} `json:"app"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode intercom me response: %w", err)
	}

	return resp.App.Name, nil
}

// resendNameResolver returns a static name for Resend.
type resendNameResolver struct{}

func NewResendNameResolver() NameResolver {
	return &resendNameResolver{}
}

func (r *resendNameResolver) ResolveInstanceName(_ context.Context) (string, error) {
	return "Resend", nil
}

// betterStackNameResolver returns the Better Stack team name captured when
// the API-key connector was created. The team name is the human-readable
// instance identifier, so no HTTP call is required.
type betterStackNameResolver struct {
	teamName string
}

func NewBetterStackNameResolver(teamName string) NameResolver {
	return &betterStackNameResolver{teamName: teamName}
}

func (r *betterStackNameResolver) ResolveInstanceName(_ context.Context) (string, error) {
	return r.teamName, nil
}

// gitlabNameResolver resolves the GitLab group name.
type gitlabNameResolver struct {
	httpClient *http.Client
	groupID    string
}

func NewGitLabNameResolver(httpClient *http.Client, groupID string) NameResolver {
	return &gitlabNameResolver{httpClient: httpClient, groupID: groupID}
}

func (r *gitlabNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	if r.groupID == "" {
		return "", nil
	}

	endpoint, err := url.JoinPath("https://gitlab.com", "api", "v4", "groups", url.PathEscape(r.groupID))
	if err != nil {
		return "", fmt.Errorf("cannot build gitlab group URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create gitlab group request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute gitlab group request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", fmt.Errorf("cannot fetch gitlab group: unexpected status %d", httpResp.StatusCode)
	}

	var resp struct {
		Name     string `json:"name"`
		FullPath string `json:"full_path"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode gitlab group response: %w", err)
	}

	if resp.Name != "" {
		return resp.Name, nil
	}

	return resp.FullPath, nil
}

// bitbucketNameResolver resolves the Bitbucket workspace name.
type bitbucketNameResolver struct {
	httpClient *http.Client
	workspace  string
}

func NewBitbucketNameResolver(httpClient *http.Client, workspace string) NameResolver {
	return &bitbucketNameResolver{httpClient: httpClient, workspace: workspace}
}

func (r *bitbucketNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	if r.workspace == "" {
		return "", nil
	}

	endpoint, err := url.JoinPath("https://api.bitbucket.org", "2.0", "workspaces", url.PathEscape(r.workspace))
	if err != nil {
		return "", fmt.Errorf("cannot build bitbucket workspace URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create bitbucket workspace request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute bitbucket workspace request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", fmt.Errorf("cannot fetch bitbucket workspace: unexpected status %d", httpResp.StatusCode)
	}

	var resp struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode bitbucket workspace response: %w", err)
	}

	if resp.Name != "" {
		return resp.Name, nil
	}

	return resp.Slug, nil
}

// herokuNameResolver resolves the Heroku team name.
type herokuNameResolver struct {
	httpClient *http.Client
	teamID     string
}

func NewHerokuNameResolver(httpClient *http.Client, teamID string) NameResolver {
	return &herokuNameResolver{httpClient: httpClient, teamID: teamID}
}

func (r *herokuNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	if r.teamID == "" {
		return "", nil
	}

	// A personal account has no Team to name; short-circuit before hitting
	// GET /teams/@personal, which 404s and would loop the source-name worker.
	if r.teamID == herokuPersonalAccountSlug {
		return herokuPersonalAccountDisplayName, nil
	}

	endpoint, err := url.JoinPath("https://api.heroku.com", "teams", url.PathEscape(r.teamID))
	if err != nil {
		return "", fmt.Errorf("cannot build heroku team URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create heroku team request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.heroku+json; version=3")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute heroku team request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", fmt.Errorf("cannot fetch heroku team: unexpected status %d", httpResp.StatusCode)
	}

	var resp struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode heroku team response: %w", err)
	}

	return resp.Name, nil
}

// pagerdutyNameResolver returns the PagerDuty subdomain stored in connector
// settings. The subdomain is captured during the OAuth callback (see
// handleConnectorComplete) so no HTTP call is required.
type pagerdutyNameResolver struct {
	subdomain string
}

func NewPagerDutyNameResolver(subdomain string) NameResolver {
	return &pagerdutyNameResolver{subdomain: subdomain}
}

func (r *pagerdutyNameResolver) ResolveInstanceName(_ context.Context) (string, error) {
	return r.subdomain, nil
}

// datadogNameResolver returns the Datadog site/region label stored in
// connector settings (e.g. "US3"), captured during the OAuth callback. No
// HTTP call is required; the AccessReviewSource title becomes "Datadog <region>".
// Org-name resolution is intentionally omitted to keep scopes to
// user_access_read (the org name endpoint needs org_management).
type datadogNameResolver struct {
	region string
}

func NewDatadogNameResolver(region string) NameResolver {
	return &datadogNameResolver{region: region}
}

func (r *datadogNameResolver) ResolveInstanceName(_ context.Context) (string, error) {
	return r.region, nil
}

// oktaNameResolver resolves the Okta org name via GET /api/v1/org on the
// configured org host. A non-2xx is terminal — a read-only API token may
// lack org-settings read, so it returns ("", nil) to keep the generic
// source name rather than make the source-name worker retry forever.
type oktaNameResolver struct {
	httpClient *http.Client
	domain     string
}

func NewOktaNameResolver(httpClient *http.Client, domain string) NameResolver {
	return &oktaNameResolver{httpClient: httpClient, domain: domain}
}

func (r *oktaNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	if r.domain == "" {
		return "", nil
	}

	endpoint := url.URL{Scheme: "https", Host: r.domain, Path: "/api/v1/org"}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return "", fmt.Errorf("cannot create okta org request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute okta org request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", nil
	}

	var resp struct {
		CompanyName string `json:"companyName"`
		Subdomain   string `json:"subdomain"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode okta org response: %w", err)
	}

	if resp.CompanyName != "" {
		return resp.CompanyName, nil
	}

	return resp.Subdomain, nil
}

// zendeskNameResolver returns the Zendesk subdomain stored in connector
// settings (e.g. "acme" for acme.zendesk.com), captured at connect time. No
// HTTP call is required; the AccessReviewSource title becomes "Zendesk <subdomain>".
// Account-name resolution is intentionally omitted to keep the scope to
// users:read (Zendesk exposes no human account name on that scope).
type zendeskNameResolver struct {
	subdomain string
}

func NewZendeskNameResolver(subdomain string) NameResolver {
	return &zendeskNameResolver{subdomain: subdomain}
}

func (r *zendeskNameResolver) ResolveInstanceName(_ context.Context) (string, error) {
	return r.subdomain, nil
}

// asanaNameResolver resolves the Asana workspace name.
type asanaNameResolver struct {
	httpClient   *http.Client
	workspaceGID string
}

func NewAsanaNameResolver(httpClient *http.Client, workspaceGID string) NameResolver {
	return &asanaNameResolver{httpClient: httpClient, workspaceGID: workspaceGID}
}

func (r *asanaNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	if r.workspaceGID == "" {
		return "", nil
	}

	endpoint, err := url.JoinPath("https://app.asana.com", "api", "1.0", "workspaces", url.PathEscape(r.workspaceGID))
	if err != nil {
		return "", fmt.Errorf("cannot build asana workspace URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create asana workspace request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute asana workspace request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", fmt.Errorf("cannot fetch asana workspace: unexpected status %d", httpResp.StatusCode)
	}

	var resp struct {
		Data struct {
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode asana workspace response: %w", err)
	}

	return resp.Data.Name, nil
}

// netlifyNameResolver resolves the Netlify account name.
type netlifyNameResolver struct {
	httpClient  *http.Client
	accountSlug string
}

func NewNetlifyNameResolver(httpClient *http.Client, accountSlug string) NameResolver {
	return &netlifyNameResolver{httpClient: httpClient, accountSlug: accountSlug}
}

func (r *netlifyNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	if r.accountSlug == "" {
		return "", nil
	}

	endpoint, err := url.JoinPath("https://api.netlify.com", "api", "v1", "accounts", url.PathEscape(r.accountSlug))
	if err != nil {
		return "", fmt.Errorf("cannot build netlify account URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create netlify account request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute netlify account request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", fmt.Errorf("cannot fetch netlify account: unexpected status %d", httpResp.StatusCode)
	}

	var resp struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode netlify account response: %w", err)
	}

	return resp.Name, nil
}

// clickupNameResolver resolves the ClickUp team name.
type clickupNameResolver struct {
	httpClient *http.Client
	teamID     string
}

func NewClickUpNameResolver(httpClient *http.Client, teamID string) NameResolver {
	return &clickupNameResolver{httpClient: httpClient, teamID: teamID}
}

func (r *clickupNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	if r.teamID == "" {
		return "", nil
	}

	endpoint, err := url.JoinPath("https://api.clickup.com", "api", "v2", "team", url.PathEscape(r.teamID))
	if err != nil {
		return "", fmt.Errorf("cannot build clickup team URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create clickup team request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute clickup team request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", fmt.Errorf("cannot fetch clickup team: unexpected status %d", httpResp.StatusCode)
	}

	var resp struct {
		Team struct {
			Name string `json:"name"`
		} `json:"team"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode clickup team response: %w", err)
	}

	return resp.Team.Name, nil
}

// vercelNameResolver resolves the Vercel team name. When the captured
// TeamID is a personal-account UID, the v2 teams endpoint returns 404;
// the resolver falls back to /v2/user and uses `username` (or `name`)
// as the display name.
type vercelNameResolver struct {
	httpClient *http.Client
	teamID     string
}

func NewVercelNameResolver(httpClient *http.Client, teamID string) NameResolver {
	return &vercelNameResolver{httpClient: httpClient, teamID: teamID}
}

func (r *vercelNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	if r.teamID == "" {
		return "", nil
	}

	teamURL, err := url.JoinPath("https://api.vercel.com", "v2", "teams", url.PathEscape(r.teamID))
	if err != nil {
		return "", fmt.Errorf("cannot build vercel team URL: %w", err)
	}

	teamReq, err := http.NewRequestWithContext(ctx, http.MethodGet, teamURL, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create vercel team request: %w", err)
	}

	teamReq.Header.Set("Accept", "application/json")

	teamResp, err := r.httpClient.Do(teamReq)
	if err != nil {
		return "", fmt.Errorf("cannot execute vercel team request: %w", err)
	}

	defer func() { _ = teamResp.Body.Close() }()

	if teamResp.StatusCode == http.StatusOK {
		var body struct {
			Name string `json:"name"`
			Slug string `json:"slug"`
		}
		if err := json.NewDecoder(teamResp.Body).Decode(&body); err != nil {
			return "", fmt.Errorf("cannot decode vercel team response: %w", err)
		}

		if body.Name != "" {
			return body.Name, nil
		}

		return body.Slug, nil
	}

	if teamResp.StatusCode != http.StatusNotFound {
		return "", fmt.Errorf("cannot fetch vercel team: unexpected status %d", teamResp.StatusCode)
	}

	// Personal-account fallback: /v2/teams/<uid> returns 404, but
	// /v2/user works with the same Bearer token.
	user, err := connector.FetchVercelUser(ctx, r.httpClient)
	if err != nil {
		return "", err
	}

	if user.Username != "" {
		return user.Username, nil
	}

	return user.Name, nil
}

// mondayNameResolver resolves the Monday.com account name via GraphQL.
type mondayNameResolver struct {
	httpClient *http.Client
}

func NewMondayNameResolver(httpClient *http.Client) NameResolver {
	return &mondayNameResolver{httpClient: httpClient}
}

func (r *mondayNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	body := struct {
		Query string `json:"query"`
	}{
		Query: `query { account { id name slug tier } }`,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("cannot marshal monday account query: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, mondayGraphQLEndpoint, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("cannot create monday account request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute monday account request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", fmt.Errorf("cannot fetch monday account: unexpected status %d", httpResp.StatusCode)
	}

	var resp struct {
		Data struct {
			Account struct {
				Name string `json:"name"`
			} `json:"account"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode monday account response: %w", err)
	}

	if len(resp.Errors) > 0 {
		// Provider-supplied messages may carry tenant identifiers or
		// query fragments — never embed them. Driver scrubs the same
		// field; keep both call sites aligned.
		return "", fmt.Errorf("cannot fetch monday account: graphql error")
	}

	return resp.Data.Account.Name, nil
}

// notionNameResolver resolves the Notion workspace name via /v1/users/me.
type notionNameResolver struct {
	httpClient *http.Client
}

func NewNotionNameResolver(httpClient *http.Client) NameResolver {
	return &notionNameResolver{httpClient: httpClient}
}

func (r *notionNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.notion.com/v1/users/me", nil)
	if err != nil {
		return "", fmt.Errorf("cannot create notion users/me request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Notion-Version", notionAPIVersion)

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute notion users/me request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", fmt.Errorf("cannot fetch notion users/me: unexpected status %d", httpResp.StatusCode)
	}

	var resp struct {
		Bot struct {
			WorkspaceName string `json:"workspace_name"`
		} `json:"bot"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode notion users/me response: %w", err)
	}

	return resp.Bot.WorkspaceName, nil
}

// microsoft365NameResolver resolves the Microsoft 365 tenant display name
// via the Microsoft Graph organization endpoint.
type microsoft365NameResolver struct {
	httpClient *http.Client
}

func NewMicrosoft365NameResolver(httpClient *http.Client) NameResolver {
	return &microsoft365NameResolver{httpClient: httpClient}
}

func (r *microsoft365NameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	msURL, err := url.Parse("https://graph.microsoft.com/v1.0/organization")
	if err != nil {
		return "", fmt.Errorf("cannot parse microsoft 365 organization URL: %w", err)
	}

	q := msURL.Query()
	q.Set("$select", "displayName,verifiedDomains")
	msURL.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, msURL.String(), nil)
	if err != nil {
		return "", fmt.Errorf("cannot create microsoft 365 organization request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute microsoft 365 organization request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", fmt.Errorf("cannot fetch microsoft 365 organization: unexpected status %d", httpResp.StatusCode)
	}

	var resp struct {
		Value []struct {
			DisplayName     string `json:"displayName"`
			VerifiedDomains []struct {
				Name      string `json:"name"`
				IsDefault bool   `json:"isDefault"`
			} `json:"verifiedDomains"`
		} `json:"value"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode microsoft 365 organization response: %w", err)
	}

	if len(resp.Value) == 0 {
		return "", nil
	}

	org := resp.Value[0]
	if org.DisplayName != "" {
		return org.DisplayName, nil
	}

	for _, d := range org.VerifiedDomains {
		if d.IsDefault {
			return d.Name, nil
		}
	}

	if len(org.VerifiedDomains) > 0 {
		return org.VerifiedDomains[0].Name, nil
	}

	return "", nil
}

// railwayNameResolver resolves the Railway workspace name via GraphQL, for the
// AccessReviewSource title. With a single workspace it uses that workspace's
// name; with several it falls back to the account holder's name, since the
// source spans all of the account's workspaces.
type railwayNameResolver struct {
	httpClient *http.Client
}

func NewRailwayNameResolver(httpClient *http.Client) NameResolver {
	return &railwayNameResolver{httpClient: httpClient}
}

func (r *railwayNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	httpResp, err := railwayPost(ctx, r.httpClient, "account", `query { me { name workspaces { id name } } }`)
	if err != nil {
		return "", err
	}

	defer func() { _ = httpResp.Body.Close() }()

	// Best-effort: a non-2xx must not make the source-name worker retry forever
	// — keep the generic name. (Railway also signals auth failure with a 200 +
	// errors body, handled below.)
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", nil
	}

	var resp struct {
		Data struct {
			Me *struct {
				Name       string `json:"name"`
				Workspaces []struct {
					Name string `json:"name"`
				} `json:"workspaces"`
			} `json:"me"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode railway account response: %w", err)
	}

	if len(resp.Errors) > 0 || resp.Data.Me == nil {
		return "", nil
	}

	// A single workspace names the source directly; with several (or none) fall
	// back to the account holder's display name. Never the email — a terminal
	// empty result keeps the generic source name, which the worker tolerates.
	me := resp.Data.Me
	if len(me.Workspaces) == 1 {
		return me.Workspaces[0].Name, nil
	}

	return me.Name, nil
}

// crispNameResolver resolves the Crisp website name via GET /v1/website/{id},
// for the AccessReviewSource title. Like the driver it sends the X-Crisp-Tier
// header; the Basic credential is supplied by the connection transport.
type crispNameResolver struct {
	httpClient *http.Client
	websiteID  string
}

func NewCrispNameResolver(httpClient *http.Client, websiteID string) NameResolver {
	return &crispNameResolver{httpClient: httpClient, websiteID: websiteID}
}

func (r *crispNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	if r.websiteID == "" {
		return "", nil
	}

	httpResp, err := crispGet(ctx, r.httpClient, "website", "website", url.PathEscape(r.websiteID))
	if err != nil {
		return "", err
	}

	defer func() { _ = httpResp.Body.Close() }()

	// Best-effort: a non-2xx (revoked token, stale website id) must not make the
	// source-name worker retry forever — keep the generic name.
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", nil
	}

	var resp struct {
		Data struct {
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode crisp website response: %w", err)
	}

	return resp.Data.Name, nil
}

// squareNameResolver resolves the Square merchant's business name via
// GET /v2/merchants/me. A Square token — OAuth or PAT — is scoped to a single
// merchant, so "me" resolves it for both connection kinds.
type squareNameResolver struct {
	httpClient *http.Client
}

func NewSquareNameResolver(httpClient *http.Client) NameResolver {
	return &squareNameResolver{httpClient: httpClient}
}

func (r *squareNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://connect.squareup.com/v2/merchants/me", nil)
	if err != nil {
		return "", fmt.Errorf("cannot create square merchant request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Square-Version", squareAPIVersion)

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute square merchant request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	// A non-2xx (revoked token, missing scope) is terminal: keep the generic
	// source name rather than make the source-name worker retry forever.
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", nil
	}

	var resp struct {
		Merchant struct {
			BusinessName string `json:"business_name"`
		} `json:"merchant"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode square merchant response: %w", err)
	}

	return resp.Merchant.BusinessName, nil
}

// googleAnalyticsNameResolver resolves a GA4 account's display name.
type googleAnalyticsNameResolver struct {
	httpClient *http.Client
	accountID  string
}

func NewGoogleAnalyticsNameResolver(httpClient *http.Client, accountID string) NameResolver {
	return &googleAnalyticsNameResolver{httpClient: httpClient, accountID: accountID}
}

func (r *googleAnalyticsNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	if r.accountID == "" {
		return "", nil
	}

	endpoint, err := url.JoinPath("https://"+googleAnalyticsAPIHost, "v1alpha", "accounts", url.PathEscape(r.accountID))
	if err != nil {
		return "", fmt.Errorf("cannot build google analytics account URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create google analytics account request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute google analytics account request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	// A non-2xx (revoked token, renamed/deleted account) is terminal: keep the
	// generic source name rather than retry forever.
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", nil
	}

	var resp struct {
		DisplayName string `json:"displayName"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode google analytics account response: %w", err)
	}

	return resp.DisplayName, nil
}
