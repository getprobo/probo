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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"go.probo.inc/probo/pkg/coredata"
)

// ErrCrispPluginNotSubscribed is returned by GetCrispSubscriptionSettings when
// Crisp answers 404: the plugin is not subscribed to the given website, so no
// per-website settings exist yet. It is an expected verification state (the
// customer has not installed/configured the plugin on that website) rather than
// a failure, and callers distinguish it with errors.Is.
var ErrCrispPluginNotSubscribed = errors.New("crisp plugin not subscribed to website")

// CrispSubscriptionSettings is the schema-defined, per-website configuration of
// the Probo Crisp plugin. Only the field Probo relies on for ownership
// verification is modeled; unknown schema properties are ignored on decode.
type CrispSubscriptionSettings struct {
	ProboVerificationCode string `json:"probo_verification_code"`
}

// crispSubscriptionSettingsResponse is the envelope of
// GET /v1/plugins/subscription/{website_id}/{plugin_id}/settings. The active
// per-website configuration lives at data.settings; data itself also carries
// subscription metadata (ids, secret token, JSONSchema, form/callback URLs)
// that verification does not need.
type crispSubscriptionSettingsResponse struct {
	Error bool `json:"error"`
	Data  struct {
		Settings CrispSubscriptionSettings `json:"settings"`
	} `json:"data"`
}

const (
	crispAPIBaseURL = "https://api.crisp.chat/v1"
	// crispTierHeader selects the token tier on every Crisp request. A Probo
	// connection uses a plugin token, so the value is always "plugin". This is
	// not authentication (the Basic credential is attached by the transport),
	// so the driver, probe and name resolver each set it explicitly.
	crispTierHeader = "X-Crisp-Tier"
	crispTierValue  = "plugin"
)

// CrispDriver lists the operators (dashboard agents) of a single Crisp website.
// A plugin token can be connected to several websites, so the website is
// captured up front as a connector setting; the Basic credential
// (identifier:key) is applied by the connection transport.
type CrispDriver struct {
	httpClient *http.Client
	websiteID  string
}

var _ Driver = (*CrispDriver)(nil)

type crispOperatorsResponse struct {
	Data []struct {
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
}

func NewCrispDriver(httpClient *http.Client, websiteID string) *CrispDriver {
	return &CrispDriver{
		httpClient: httpClient,
		websiteID:  websiteID,
	}
}

func (d *CrispDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	endpoint, err := url.JoinPath(crispAPIBaseURL, "website", url.PathEscape(d.websiteID), "operators", "list")
	if err != nil {
		return nil, fmt.Errorf("cannot build crisp operators URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create crisp operators request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set(crispTierHeader, crispTierValue)

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute crisp operators request: %w", err)
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
		details := op.Details

		email := strings.TrimSpace(details.Email)
		if email == "" {
			continue
		}

		records = append(records, AccountRecord{
			Email:       email,
			FullName:    crispFullName(details, email),
			Roles:       crispRoles(details.Role),
			JobTitle:    strings.TrimSpace(details.Title),
			IsAdmin:     strings.EqualFold(strings.TrimSpace(details.Role), "owner"),
			MFAStatus:   coredata.MFAStatusUnknown,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
			ExternalID:  strings.TrimSpace(details.UserID),
		})
	}

	return records, nil
}

// GetCrispSubscriptionSettings reads the Probo plugin's per-website
// subscription settings so the create-connector resolver can verify website
// ownership (matching probo_verification_code). The httpClient must already
// attach the plugin Basic credential (identifier:key); this helper only sets
// the Accept and X-Crisp-Tier headers, mirroring ListAccounts. A 404 (plugin
// not subscribed to the website) is reported as ErrCrispPluginNotSubscribed so
// callers can message it distinctly from a hard failure.
func GetCrispSubscriptionSettings(
	ctx context.Context,
	httpClient *http.Client,
	websiteID string,
	pluginID string,
) (*CrispSubscriptionSettings, error) {
	endpoint, err := url.JoinPath(
		crispAPIBaseURL,
		"plugins", "subscription",
		url.PathEscape(websiteID),
		url.PathEscape(pluginID),
		"settings",
	)
	if err != nil {
		return nil, fmt.Errorf("cannot build crisp subscription settings URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create crisp subscription settings request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set(crispTierHeader, crispTierValue)

	httpResp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute crisp subscription settings request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode == http.StatusNotFound {
		return nil, ErrCrispPluginNotSubscribed
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch crisp subscription settings: unexpected status %d", httpResp.StatusCode)
	}

	var resp crispSubscriptionSettingsResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode crisp subscription settings response: %w", err)
	}

	if resp.Error {
		return nil, fmt.Errorf("cannot fetch crisp subscription settings: crisp reported an error")
	}

	return &resp.Data.Settings, nil
}

func crispFullName(details crispOperatorDetails, fallback string) string {
	if name := strings.TrimSpace(details.FirstName + " " + details.LastName); name != "" {
		return name
	}

	return fallback
}

// crispRoles maps a Crisp operator role to a display label. Documented roles
// are owner/member; an unknown future value is passed through verbatim and no
// role yields an empty slice.
func crispRoles(role string) []string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "owner":
		return []string{"Owner"}
	case "member":
		return []string{"Member"}
	default:
		if r := strings.TrimSpace(role); r != "" {
			return []string{r}
		}

		return []string{}
	}
}
