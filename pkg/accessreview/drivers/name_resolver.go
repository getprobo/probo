// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
	"fmt"
	"net/http"

	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"

	"go.probo.inc/probo/pkg/coredata"
)

// NameResolver fetches the human-readable instance name from a provider
// (e.g. Slack workspace name, Google Workspace domain).
type NameResolver interface {
	ResolveInstanceName(ctx context.Context) (string, error)
}

var providerDisplayNames = map[coredata.ConnectorProvider]string{
	coredata.ConnectorProviderSlack:           "Slack",
	coredata.ConnectorProviderGoogleWorkspace: "Google Workspace",
	coredata.ConnectorProviderLinear:          "Linear",
	coredata.ConnectorProviderOnePassword:     "1Password",
	coredata.ConnectorProviderHubSpot:         "HubSpot",
	coredata.ConnectorProviderDocuSign:        "DocuSign",
	coredata.ConnectorProviderNotion:          "Notion",
	coredata.ConnectorProviderBrex:            "Brex",
	coredata.ConnectorProviderTally:           "Tally",
	coredata.ConnectorProviderCloudflare:      "Cloudflare",
	coredata.ConnectorProviderOpenAI:          "OpenAI",
	coredata.ConnectorProviderSentry:          "Sentry",
	coredata.ConnectorProviderSupabase:        "Supabase",
	coredata.ConnectorProviderGitHub:          "GitHub",
	coredata.ConnectorProviderIntercom:        "Intercom",
	coredata.ConnectorProviderResend:          "Resend",
}

// ProviderDisplayName returns the human-readable label for a connector provider.
func ProviderDisplayName(provider coredata.ConnectorProvider) string {
	if name, ok := providerDisplayNames[provider]; ok {
		return name
	}
	return string(provider)
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
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://api.cloudflare.com/client/v4/accounts?page=1&per_page=1",
		nil,
	)
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
	url := fmt.Sprintf("https://api.tally.so/organizations/%s", r.organizationID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

// docusignNameResolver resolves the DocuSign account name from userinfo.
type docusignNameResolver struct {
	httpClient *http.Client
}

func NewDocuSignNameResolver(httpClient *http.Client) NameResolver {
	return &docusignNameResolver{httpClient: httpClient}
}

func (r *docusignNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, docusignUserInfoEndpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create docusign userinfo request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute docusign userinfo request: %w", err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", fmt.Errorf("cannot fetch docusign userinfo: unexpected status %d", httpResp.StatusCode)
	}

	var resp struct {
		Accounts []struct {
			AccountName string `json:"account_name"`
			IsDefault   bool   `json:"is_default"`
		} `json:"accounts"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode docusign userinfo response: %w", err)
	}

	for _, account := range resp.Accounts {
		if account.IsDefault {
			return account.AccountName, nil
		}
	}

	if len(resp.Accounts) > 0 {
		return resp.Accounts[0].AccountName, nil
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

	url := fmt.Sprintf("https://sentry.io/api/0/organizations/%s/", r.orgSlug)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create sentry organization request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute sentry organization request: %w", err)
	}
	defer func() { _ = httpResp.Body.Close() }()

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
	url := fmt.Sprintf("https://api.github.com/orgs/%s", r.org)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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
