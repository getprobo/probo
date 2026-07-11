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

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
)

const (
	anthropicAPIVersion       = "2023-06-01"
	anthropicUsersProbeURL    = "https://api.anthropic.com/v1/organizations/users?limit=1"
	herokuAccountProbeURL     = "https://api.heroku.com/account"
	openRouterMembersProbeURL = "https://openrouter.ai/api/v1/organization/members?limit=1"
	linearGraphQLEndpoint     = "https://api.linear.app/graphql"
	mondayGraphQLEndpoint     = "https://api.monday.com/v2"
	railwayGraphQLEndpoint    = "https://backboard.railway.com/graphql/v2"
	posthogOrganizationPath   = "/api/organizations/@current/"
	posthogUSBaseURL          = "https://us.posthog.com"
	posthogEUBaseURL          = "https://eu.posthog.com"
	crispAPIBaseURL           = "https://api.crisp.chat/v1"
	crispTierHeader           = "X-Crisp-Tier"
	crispTierValue            = "plugin"
)

// ProbeConnection verifies that the connector credential is accepted by the
// provider. It dispatches to a provider-specific Probe closure when
// registered, otherwise issues a lightweight GET against ProbeURL or
// BuildProbeURL. An empty probe URL means the check is skipped.
func (r *Registry) ProbeConnection(
	ctx context.Context,
	httpClient *http.Client,
	conn *coredata.Connector,
) error {
	reg, ok := r.Get(conn.Provider)
	if !ok {
		return nil
	}

	if reg.Probe != nil {
		return reg.Probe(ctx, httpClient, conn)
	}

	probeURL := reg.ProbeURL
	if reg.BuildProbeURL != nil {
		built, err := reg.BuildProbeURL(conn)
		if err != nil {
			return fmt.Errorf("cannot build probe URL: %w", err)
		}

		probeURL = built
	}

	return probeGET(ctx, httpClient, probeURL)
}

func probeGET(ctx context.Context, httpClient *http.Client, probeURL string) error {
	if probeURL == "" {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, probeURL, nil)
	if err != nil {
		return fmt.Errorf("cannot create probe request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	return doProbeRequest(httpClient, req)
}

func probePOSTJSON(
	ctx context.Context,
	httpClient *http.Client,
	probeURL string,
	payload any,
	extraHeaders map[string]string,
) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("cannot marshal probe request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, probeURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("cannot create probe request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	for key, value := range extraHeaders {
		req.Header.Set(key, value)
	}

	return doProbeRequest(httpClient, req)
}

// doProbeRequest executes a probe request and maps the status to a verdict:
// 401/403 always mean the credential is rejected, any 2xx/other status means
// connected. extraReject lets a provider add statuses that also mean a hard
// rejection (e.g. OpenRouter's 404 for a non-organization key); pass none for
// the default 401/403-only contract.
func doProbeRequest(httpClient *http.Client, req *http.Request, extraReject ...int) error {
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("probe request failed: %w", err)
	}

	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusUnauthorized ||
		resp.StatusCode == http.StatusForbidden ||
		slices.Contains(extraReject, resp.StatusCode) {
		return fmt.Errorf("credential rejected: status %d", resp.StatusCode)
	}

	return nil
}

func buildDatadogProbeURL(conn *coredata.Connector) (string, error) {
	s, err := coredata.ConnectorSettings[coredata.DatadogConnectorSettings](conn)
	if err != nil {
		return "", fmt.Errorf("cannot read datadog connector settings: %w", err)
	}

	if !connector.IsValidDatadogDomain(s.Domain) {
		return "", fmt.Errorf("invalid or missing datadog domain")
	}

	q := url.Values{}
	q.Set("page[size]", "1")
	q.Set("page[number]", "0")

	endpoint := url.URL{
		Scheme:   "https",
		Host:     "api." + s.Domain,
		Path:     "/api/v2/users",
		RawQuery: q.Encode(),
	}

	return endpoint.String(), nil
}

func buildZendeskProbeURL(conn *coredata.Connector) (string, error) {
	s, err := coredata.ConnectorSettings[coredata.ZendeskConnectorSettings](conn)
	if err != nil {
		return "", fmt.Errorf("cannot read zendesk connector settings: %w", err)
	}

	if !connector.IsValidZendeskSubdomain(s.Subdomain) {
		return "", fmt.Errorf("invalid or missing zendesk subdomain")
	}

	q := url.Values{}
	q.Set("page[size]", "1")
	q.Add("role[]", "agent")
	q.Add("role[]", "admin")

	endpoint := url.URL{
		Scheme:   "https",
		Host:     s.Subdomain + ".zendesk.com",
		Path:     "/api/v2/users.json",
		RawQuery: q.Encode(),
	}

	return endpoint.String(), nil
}

func buildOktaProbeURL(conn *coredata.Connector) (string, error) {
	s, err := coredata.ConnectorSettings[coredata.OktaConnectorSettings](conn)
	if err != nil {
		return "", fmt.Errorf("cannot read okta connector settings: %w", err)
	}

	if !connector.IsValidOktaDomain(s.Domain) {
		return "", fmt.Errorf("invalid or missing okta domain")
	}

	endpoint := url.URL{
		Scheme:   "https",
		Host:     s.Domain,
		Path:     "/api/v1/users",
		RawQuery: url.Values{"limit": {"1"}}.Encode(),
	}

	return endpoint.String(), nil
}

func buildNeonProbeURL(conn *coredata.Connector) (string, error) {
	s, err := coredata.ConnectorSettings[coredata.NeonConnectorSettings](conn)
	if err != nil {
		return "", fmt.Errorf("cannot read neon connector settings: %w", err)
	}

	if s.OrganizationID == "" {
		return "", fmt.Errorf("missing neon organization_id")
	}

	endpoint, err := url.JoinPath(
		"https://console.neon.tech/api/v2",
		"organizations",
		url.PathEscape(s.OrganizationID),
		"members",
	)
	if err != nil {
		return "", fmt.Errorf("cannot build neon probe URL: %w", err)
	}

	q := url.Values{"limit": {"1"}}

	return endpoint + "?" + q.Encode(), nil
}

func buildScalewayProbeURL(conn *coredata.Connector) (string, error) {
	s, err := coredata.ConnectorSettings[coredata.ScalewayConnectorSettings](conn)
	if err != nil {
		return "", fmt.Errorf("cannot read scaleway connector settings: %w", err)
	}

	if s.OrganizationID == "" {
		return "", fmt.Errorf("missing scaleway organization_id")
	}

	endpoint := url.URL{
		Scheme: "https",
		Host:   "api.scaleway.com",
		Path:   "/iam/v1alpha1/users",
		RawQuery: url.Values{
			"organization_id": {s.OrganizationID},
			"page_size":       {"1"},
		}.Encode(),
	}

	return endpoint.String(), nil
}

func buildRenderProbeURL(conn *coredata.Connector) (string, error) {
	s, err := coredata.ConnectorSettings[coredata.RenderConnectorSettings](conn)
	if err != nil {
		return "", fmt.Errorf("cannot read render connector settings: %w", err)
	}

	if s.OwnerID == "" {
		return "", fmt.Errorf("missing render owner_id")
	}

	return url.JoinPath(
		"https://api.render.com/v1",
		"owners",
		url.PathEscape(s.OwnerID),
		"members",
	)
}

func buildQoveryProbeURL(conn *coredata.Connector) (string, error) {
	s, err := coredata.ConnectorSettings[coredata.QoveryConnectorSettings](conn)
	if err != nil {
		return "", fmt.Errorf("cannot read qovery connector settings: %w", err)
	}

	if s.OrganizationID == "" {
		return "", fmt.Errorf("missing qovery organization_id")
	}

	return url.JoinPath(
		"https://api.qovery.com",
		"organization",
		url.PathEscape(s.OrganizationID),
		"member",
	)
}

func buildGrafanaProbeURL(conn *coredata.Connector) (string, error) {
	s, err := coredata.ConnectorSettings[coredata.GrafanaConnectorSettings](conn)
	if err != nil {
		return "", fmt.Errorf("cannot read grafana connector settings: %w", err)
	}

	baseURL, err := normalizeGrafanaBaseURL(s.BaseURL)
	if err != nil {
		return "", err
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("cannot parse grafana base URL: %w", err)
	}

	u = u.JoinPath("api", "org", "users")
	q := u.Query()
	q.Set("perpage", "1")
	q.Set("page", "1")
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func buildMetabaseProbeURL(conn *coredata.Connector) (string, error) {
	s, err := coredata.ConnectorSettings[coredata.MetabaseConnectorSettings](conn)
	if err != nil {
		return "", fmt.Errorf("cannot read metabase connector settings: %w", err)
	}

	instanceURL := strings.TrimSpace(s.InstanceURL)
	if instanceURL == "" {
		return "", fmt.Errorf("missing metabase instance_url")
	}

	if err := validateMetabaseInstanceURL(instanceURL); err != nil {
		return "", err
	}

	u, err := url.Parse(instanceURL)
	if err != nil {
		return "", fmt.Errorf("cannot parse metabase instance URL: %w", err)
	}

	endpoint := u.JoinPath("api", "user")
	q := endpoint.Query()
	q.Set("status", "all")
	q.Set("limit", "1")
	q.Set("offset", "0")
	endpoint.RawQuery = q.Encode()

	return endpoint.String(), nil
}

func buildLangfuseProbeURL(conn *coredata.Connector) (string, error) {
	s, err := coredata.ConnectorSettings[coredata.LangfuseConnectorSettings](conn)
	if err != nil {
		return "", fmt.Errorf("cannot read langfuse connector settings: %w", err)
	}

	baseURL, err := normalizeLangfuseBaseURL(s.BaseURL)
	if err != nil {
		return "", err
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("cannot parse langfuse base URL: %w", err)
	}

	return u.JoinPath("api", "public", "organizations", "memberships").String(), nil
}

func buildSigNozProbeURL(conn *coredata.Connector) (string, error) {
	s, err := coredata.ConnectorSettings[coredata.SigNozConnectorSettings](conn)
	if err != nil {
		return "", fmt.Errorf("cannot read signoz connector settings: %w", err)
	}

	baseURL, err := normalizeSigNozBaseURL(s.BaseURL)
	if err != nil {
		return "", err
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("cannot parse signoz base URL: %w", err)
	}

	return u.JoinPath("api", "v1", "user").String(), nil
}

func buildPostHogProbeURL(conn *coredata.Connector) (string, error) {
	s, err := coredata.ConnectorSettings[coredata.PostHogConnectorSettings](conn)
	if err != nil {
		return "", fmt.Errorf("cannot read posthog connector settings: %w", err)
	}

	baseURL := strings.TrimSpace(s.BaseURL)
	if baseURL == "" {
		return "", nil
	}

	return url.JoinPath(baseURL, posthogOrganizationPath)
}

func probeLinear(
	ctx context.Context,
	httpClient *http.Client,
	_ *coredata.Connector,
) error {
	return probePOSTJSON(
		ctx,
		httpClient,
		linearGraphQLEndpoint,
		map[string]string{"query": "{ viewer { id } }"},
		nil,
	)
}

func probeMonday(
	ctx context.Context,
	httpClient *http.Client,
	_ *coredata.Connector,
) error {
	return probePOSTJSON(
		ctx,
		httpClient,
		mondayGraphQLEndpoint,
		map[string]string{"query": "query { users(limit: 1) { id } }"},
		nil,
	)
}

// probeRailway verifies a Railway account token. Railway returns HTTP 200 with
// a populated errors array (and data.me null) for a rejected token rather than
// 401/403, so the generic probe would falsely pass — this closure inspects the
// response body instead.
func probeRailway(
	ctx context.Context,
	httpClient *http.Client,
	_ *coredata.Connector,
) error {
	body, err := json.Marshal(map[string]string{"query": "query { me { id } }"})
	if err != nil {
		return fmt.Errorf("cannot marshal railway probe request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, railwayGraphQLEndpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("cannot create railway probe request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("railway probe request failed: %w", err)
	}

	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("credential rejected: status %d", resp.StatusCode)
	}

	var parsed struct {
		Data struct {
			Me *struct {
				ID string `json:"id"`
			} `json:"me"`
		} `json:"data"`
		Errors []json.RawMessage `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return fmt.Errorf("cannot decode railway probe response: %w", err)
	}

	if len(parsed.Errors) > 0 || parsed.Data.Me == nil {
		return fmt.Errorf("credential rejected: railway returned no authenticated account")
	}

	return nil
}

// probeCrisp verifies a Crisp plugin token against the configured website.
// Every Crisp request needs the non-auth X-Crisp-Tier header, which the default
// probeGET does not set, so this closure builds the request itself; the Basic
// credential is attached by the connection transport. Beyond the usual 401/403,
// it treats 404 as a rejection too: a valid token whose website_id is wrong or
// unbound returns 404 on operators/list — a permanent misconfiguration that
// would otherwise pass the probe and fail every later access review, so it
// surfaces at connection time instead.
func probeCrisp(
	ctx context.Context,
	httpClient *http.Client,
	conn *coredata.Connector,
) error {
	s, err := coredata.ConnectorSettings[coredata.CrispConnectorSettings](conn)
	if err != nil {
		return fmt.Errorf("cannot read crisp connector settings: %w", err)
	}

	if s.WebsiteID == "" {
		return fmt.Errorf("missing crisp website_id")
	}

	endpoint, err := url.JoinPath(crispAPIBaseURL, "website", url.PathEscape(s.WebsiteID), "operators", "list")
	if err != nil {
		return fmt.Errorf("cannot build crisp probe URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("cannot create crisp probe request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set(crispTierHeader, crispTierValue)

	return doProbeRequest(httpClient, req, http.StatusNotFound)
}

func probeAnthropic(
	ctx context.Context,
	httpClient *http.Client,
	_ *coredata.Connector,
) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, anthropicUsersProbeURL, nil)
	if err != nil {
		return fmt.Errorf("cannot create probe request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("anthropic-version", anthropicAPIVersion)

	return doProbeRequest(httpClient, req)
}

func probeHeroku(
	ctx context.Context,
	httpClient *http.Client,
	_ *coredata.Connector,
) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, herokuAccountProbeURL, nil)
	if err != nil {
		return fmt.Errorf("cannot create probe request: %w", err)
	}

	// Heroku negotiates the API version through the Accept media type; the
	// generic "application/json" the default probe sends yields 400 (not
	// 401/403), which doProbeRequest would read as "connected" and mask a
	// dead token. Send the versioned Accept so a revoked token surfaces as
	// 401 (verified live: 400 with application/json, 401 with this header).
	req.Header.Set("Accept", "application/vnd.heroku+json; version=3")

	return doProbeRequest(httpClient, req)
}

// probeOpenRouter verifies an OpenRouter management key. Beyond the usual
// 401/403, it treats 404 as a rejection too: a personal (non-organization)
// key authenticates but the members endpoint returns 404 "This endpoint is
// only available for organization accounts" (verified live) — a permanent,
// not transient, signal that the connector can never list anyone, so it
// surfaces at connection time instead of failing a campaign later.
func probeOpenRouter(
	ctx context.Context,
	httpClient *http.Client,
	_ *coredata.Connector,
) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, openRouterMembersProbeURL, nil)
	if err != nil {
		return fmt.Errorf("cannot create probe request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	return doProbeRequest(httpClient, req, http.StatusNotFound)
}

func probePostHog(
	ctx context.Context,
	httpClient *http.Client,
	conn *coredata.Connector,
) error {
	probeURL, err := buildPostHogProbeURL(conn)
	if err != nil {
		return err
	}

	if probeURL != "" {
		return probeGET(ctx, httpClient, probeURL)
	}

	for _, host := range []string{posthogUSBaseURL, posthogEUBaseURL} {
		endpoint, err := url.JoinPath(host, posthogOrganizationPath)
		if err != nil {
			continue
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			continue
		}

		req.Header.Set("Accept", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return fmt.Errorf("cannot probe posthog region: %w", ctx.Err())
			}

			continue
		}

		status := resp.StatusCode
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()

		if status == http.StatusUnauthorized || status == http.StatusForbidden {
			return fmt.Errorf("credential rejected: status %d", status)
		}

		if status >= http.StatusOK && status < http.StatusMultipleChoices {
			return nil
		}
	}

	return fmt.Errorf("credential rejected: no posthog region accepted the connection")
}
