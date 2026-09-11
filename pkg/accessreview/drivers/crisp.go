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

package drivers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"go.probo.inc/probo/pkg/coredata"
)

// ErrCrispPluginNotSubscribed is returned by GetCrispSubscription when Crisp
// answers 404: the plugin is not subscribed to the given website, so no
// subscription exists yet. It is an expected verification state (the customer
// has not installed the plugin on that website) rather than a failure, and
// callers distinguish it with errors.Is.
var ErrCrispPluginNotSubscribed = errors.New("crisp plugin not subscribed to website")

// CrispSubscription is the Probo plugin's subscription to one Crisp website.
// Only the field the install callback verifies against is modeled.
type CrispSubscription struct {
	// Token is the per-(website, plugin) secret Crisp also hands the browser
	// on the install callback. Reading it back server-side with Probo's own
	// plugin credential is what makes that callback verifiable.
	Token string `json:"token"`
}

// CrispStatusError carries the HTTP status of a non-2xx Crisp response so a
// caller can classify it. GetCrispSubscription returns it for every non-2xx
// other than 404 (which stays ErrCrispPluginNotSubscribed), because the install
// callback has to tell a terminal 401/403 from a retryable 429/5xx: the first
// burns the customer's single-use state, the second releases it.
type CrispStatusError struct {
	Operation string
	Code      int
}

func (e *CrispStatusError) Error() string {
	return fmt.Sprintf("cannot fetch crisp %s: unexpected status %d", e.Operation, e.Code)
}

// crispSubscriptionResponse is the envelope of
// GET /v1/plugins/subscription/{website_id}/{plugin_id}/settings. Despite the
// path, data is the subscription itself (ids, secret token, JSONSchema,
// form/callback URLs); the schema-defined per-website configuration sits one
// level down at data.settings, which verification does not need.
type crispSubscriptionResponse struct {
	Error bool              `json:"error"`
	Data  CrispSubscription `json:"data"`
}

const (
	// crispTierHeader selects the token tier on every Crisp request. A Probo
	// connection uses a plugin token, so the value is always "plugin". This is
	// not authentication (the Basic credential is attached by the transport),
	// so the driver, probe and name resolver each set it explicitly.
	crispTierHeader = "X-Crisp-Tier"
	crispTierValue  = "plugin"
)

// crispDefaultBaseURL is the Crisp REST API root. It backs only the exported
// GetCrispSubscription, which the install callback calls with no registration —
// and therefore no Endpoints — in scope. The driver and the name resolver go
// through their injected baseURL instead.
const crispDefaultBaseURL = "https://api.crisp.chat/v1"

// CrispDriver lists the operators (dashboard agents) of a single Crisp website.
// A plugin token can be connected to several websites, so the website is
// captured up front as a connector setting; the Basic credential
// (identifier:key) is applied by the connection transport.
type CrispDriver struct {
	httpClient *http.Client
	websiteID  string
	baseURL    string
}

var _ Driver = (*CrispDriver)(nil)

// crispMemberTypeOperator is the only operators/list member type that is a
// human holding access today. Crisp also returns "invite" (invited but not
// joined, and carrying no user_id to key an entry on) and "sandbox" (a Crisp
// Marketplace plugin developer, not a workspace member) from the same
// endpoint; both would otherwise surface as accounts nobody granted.
const crispMemberTypeOperator = "operator"

type crispOperatorsResponse struct {
	Data []struct {
		Type    string               `json:"type"`
		Details crispOperatorDetails `json:"details"`
	} `json:"data"`
}

type crispOperatorDetails struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
	Title     string `json:"title"`
	// HasToken is Crisp's two-factor flag, documented as "whether operator has
	// Two Factor Authentication enabled or not". A pointer so an absent field
	// stays Unknown instead of reporting every operator as MFA-disabled.
	HasToken *bool `json:"has_token"`
}

// crispMFAStatus maps Crisp's has_token flag onto an MFA status. Crisp
// documents the field for operator members only, so a missing value means the
// endpoint told us nothing rather than that the factor is off.
func crispMFAStatus(hasToken *bool) coredata.MFAStatus {
	if hasToken == nil {
		return coredata.MFAStatusUnknown
	}

	if *hasToken {
		return coredata.MFAStatusEnabled
	}

	return coredata.MFAStatusDisabled
}

// NewCrispDriver builds a driver against baseURL, the versioned Crisp REST
// API origin (e.g. https://api.crisp.chat/v1).
func NewCrispDriver(httpClient *http.Client, websiteID, baseURL string) *CrispDriver {
	return &CrispDriver{
		httpClient: httpClient,
		websiteID:  websiteID,
		baseURL:    baseURL,
	}
}

func (d *CrispDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	httpResp, err := crispGet(ctx, d.httpClient, d.baseURL, "operators", "website", url.PathEscape(d.websiteID), "operators", "list")
	if err != nil {
		return nil, err
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch crisp operators: unexpected status %d", httpResp.StatusCode)
	}

	var resp crispOperatorsResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode crisp operators response: %w", err)
	}

	records := make([]AccountRecord, 0, len(resp.Data))

	for _, op := range resp.Data {
		if !strings.EqualFold(strings.TrimSpace(op.Type), crispMemberTypeOperator) {
			continue
		}

		details := op.Details

		email := strings.TrimSpace(details.Email)
		if email == "" {
			continue
		}

		records = append(records, AccountRecord{
			Email:       email,
			FullName:    crispFullName(details, email),
			Roles:       ownerMemberRoles(details.Role),
			JobTitle:    strings.TrimSpace(details.Title),
			IsAdmin:     new(isOwnerRole(details.Role)),
			MFAStatus:   crispMFAStatus(details.HasToken),
			AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
			ExternalID:  strings.TrimSpace(details.UserID),
		})
	}

	return records, nil
}

// GetCrispSubscription reads the Probo plugin's subscription to a website so
// the install callback can verify that the browser-supplied token is the one
// Crisp issued for that (website, plugin) pair. The httpClient must already
// attach the plugin Basic credential (identifier:key); this helper only sets
// the Accept and X-Crisp-Tier headers, mirroring ListAccounts. A 404 (plugin
// not subscribed to the website) is reported as ErrCrispPluginNotSubscribed;
// every other non-2xx becomes a *CrispStatusError so the caller can tell a
// terminal 401/403 from a retryable 429/5xx.
func GetCrispSubscription(
	ctx context.Context,
	httpClient *http.Client,
	websiteID string,
	pluginID string,
) (*CrispSubscription, error) {
	const operation = "subscription settings"

	httpResp, err := crispGet(
		ctx,
		httpClient,
		crispDefaultBaseURL,
		operation,
		"plugins", "subscription",
		url.PathEscape(websiteID),
		url.PathEscape(pluginID),
		"settings",
	)
	if err != nil {
		return nil, err
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode == http.StatusNotFound {
		return nil, ErrCrispPluginNotSubscribed
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, &CrispStatusError{Operation: operation, Code: httpResp.StatusCode}
	}

	var resp crispSubscriptionResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode crisp subscription settings response: %w", err)
	}

	if resp.Error {
		return nil, fmt.Errorf("cannot fetch crisp subscription settings: crisp reported an error")
	}

	return &resp.Data, nil
}

// crispGet issues an authenticated GET against the Crisp API for the given path
// segments (joined onto baseURL), setting the Accept and X-Crisp-Tier headers
// every Crisp request needs; the Basic plugin credential is attached by the
// connection transport. The caller owns status-code handling and must close the
// returned response body. label names the request in wrapped errors.
func crispGet(ctx context.Context, httpClient *http.Client, baseURL, label string, path ...string) (*http.Response, error) {
	endpoint, err := url.JoinPath(baseURL, path...)
	if err != nil {
		return nil, fmt.Errorf("cannot build crisp %s URL: %w", label, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create crisp %s request: %w", label, err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set(crispTierHeader, crispTierValue)

	httpResp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute crisp %s request: %w", label, err)
	}

	return httpResp, nil
}

func crispFullName(details crispOperatorDetails, fallback string) string {
	if name := strings.TrimSpace(details.FirstName + " " + details.LastName); name != "" {
		return name
	}

	return fallback
}

// crispNameResolver resolves the Crisp website name via GET /v1/website/{id},
// for the AccessReviewSource title. Like the driver it sends the X-Crisp-Tier
// header; the Basic credential is supplied by the connection transport.
type crispNameResolver struct {
	httpClient *http.Client
	websiteID  string
	baseURL    string
}

// NewCrispNameResolver resolves the website name against baseURL, the
// versioned Crisp REST API origin (e.g. https://api.crisp.chat/v1).
func NewCrispNameResolver(httpClient *http.Client, websiteID, baseURL string) NameResolver {
	return &crispNameResolver{httpClient: httpClient, websiteID: websiteID, baseURL: baseURL}
}

func (r *crispNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	if r.websiteID == "" {
		return "", nil
	}

	httpResp, err := crispGet(ctx, r.httpClient, r.baseURL, "website", "website", url.PathEscape(r.websiteID))
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
