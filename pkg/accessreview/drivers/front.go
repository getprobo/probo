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
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"go.probo.inc/probo/pkg/coredata"
)

// FrontDriver reports who holds access to a Front company, from
// GET /teammates — the Core API's company-wide teammate list, which the
// connection's credential (OAuth token or API token) is already scoped to.
//
// What is reported: every teammate the company exposes, blocked ones included,
// since classification belongs to the reviewer. `is_blocked` is the API's only
// account-status signal and maps to Active; `is_admin` maps to IsAdmin and to
// an "Admin" role label. Front's bot teammates (rules, macros, integrations,
// OAuth clients — the non-"user"/"visitor" values of `type`) are access holders
// too and are reported as service accounts, with their type carried as a role
// qualifier so a reviewer can see which automation the grant belongs to.
//
// What is deliberately NOT reported:
//
//   - Availability. `is_available` is a presence toggle ("away"), not an
//     account state; reading it as Active would mark everyone who logged off
//     as deactivated.
//   - Inbox and team membership. GET /teammates/{id}/inboxes and the team
//     endpoints would describe WHICH resources a teammate reaches, one request
//     per teammate; the campaign question is who has access to Front, and the
//     resources are not access holders.
//   - Custom fields. They are customer-defined and may carry personal data, so
//     they are left undecoded rather than folded into a record.
//
// What the API cannot answer: Front exposes no MFA state, no last-login, and
// no account-creation timestamp on a teammate, so MFAStatus stays UNKNOWN and
// LastLogin/CreatedAt stay nil rather than being invented.
type FrontDriver struct {
	httpClient *http.Client
	baseURL    string
}

var _ Driver = (*FrontDriver)(nil)

const (
	frontTeammatesPath = "/teammates"
	frontMePath        = "/me"
)

// NewFrontDriver builds a driver against baseURL, the Core API origin (e.g.
// https://api2.frontapp.com).
func NewFrontDriver(httpClient *http.Client, baseURL string) *FrontDriver {
	// Copy the caller's client and swap only its transport: the connection's
	// client carries SSRF protection in both its transport dial check and its
	// CheckRedirect, and a fresh &http.Client{} would silently drop the second.
	retryClient := *httpClient
	retryClient.Transport = &retryRoundTripper{
		next:       httpClient.Transport,
		maxRetries: 3,
	}

	return &FrontDriver{
		httpClient: &retryClient,
		baseURL:    baseURL,
	}
}

type (
	frontTeammatesPage struct {
		Pagination *struct {
			Next string `json:"next"`
		} `json:"_pagination"`
		Results []frontTeammate `json:"_results"`
	}

	frontTeammate struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		Username  string `json:"username"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		IsAdmin   bool   `json:"is_admin"`
		IsBlocked bool   `json:"is_blocked"`
		Type      string `json:"type"`
	}
)

func (d *FrontDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	next, err := url.JoinPath(d.baseURL, frontTeammatesPath)
	if err != nil {
		return nil, fmt.Errorf("cannot build front teammates URL: %w", err)
	}

	var records []AccountRecord

	for range maxPaginationPages {
		page, err := d.fetchTeammates(ctx, next)
		if err != nil {
			return nil, err
		}

		for _, teammate := range page.Results {
			if teammate.ID == "" {
				continue
			}

			records = append(records, frontAccountRecord(teammate))
		}

		// An empty page ends the walk regardless of the cursor: the Core API
		// keeps `_pagination` on the response even where the collection is
		// unpaginated, and trusting a stale cursor over an empty page would
		// spin until the guard trips.
		if page.Pagination == nil || page.Pagination.Next == "" || len(page.Results) == 0 {
			return records, nil
		}

		next, err = sameHostNextPageURL("front", d.baseURL, page.Pagination.Next)
		if err != nil {
			return nil, err
		}
	}

	return nil, fmt.Errorf("cannot list all front teammates: %w", ErrPaginationLimitReached)
}

func (d *FrontDriver) fetchTeammates(ctx context.Context, endpoint string) (*frontTeammatesPage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create front teammates request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute front teammates request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < http.StatusOK || httpResp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("cannot fetch front teammates: unexpected status %d", httpResp.StatusCode)
	}

	var page frontTeammatesPage
	if err := json.NewDecoder(httpResp.Body).Decode(&page); err != nil {
		return nil, fmt.Errorf("cannot decode front teammates response: %w", err)
	}

	return &page, nil
}

func frontAccountRecord(teammate frontTeammate) AccountRecord {
	active := !teammate.IsBlocked

	return AccountRecord{
		Email:       teammate.Email,
		FullName:    frontFullName(teammate),
		Roles:       frontRoles(teammate),
		Active:      &active,
		IsAdmin:     teammate.IsAdmin,
		MFAStatus:   coredata.MFAStatusUnknown,
		AuthMethod:  frontAuthMethod(teammate.Type),
		AccountType: frontAccountType(teammate.Type),
		ExternalID:  teammate.ID,
	}
}

// frontFullName joins the name fields, falling back to the "@" mention
// username: Front's bot teammates carry no first/last name, and an entry with
// no display name at all is harder for a reviewer to place than one named after
// the automation.
func frontFullName(teammate frontTeammate) string {
	name := strings.TrimSpace(strings.Join([]string{teammate.FirstName, teammate.LastName}, " "))
	if name != "" {
		return name
	}

	return teammate.Username
}

// frontRoles labels the grant with the only permission Front exposes on a
// teammate (admin or not), plus the account's type for the non-human ones so a
// reviewer can tell an integration's access from a rule's.
func frontRoles(teammate frontTeammate) []string {
	roles := []string{"Teammate"}
	if teammate.IsAdmin {
		roles = []string{"Admin"}
	}

	if label := frontTypeLabel(teammate.Type); label != "" {
		roles = append(roles, label)
	}

	return roles
}

// frontTypeLabel renders a bot teammate's type as a human-readable qualifier.
// A human teammate ("user") and a chat visitor ("visitor") need none: the base
// role already describes them.
func frontTypeLabel(accountType string) string {
	switch strings.ToLower(strings.TrimSpace(accountType)) {
	case "", "user", "visitor":
		return ""
	case "ai":
		return "Type: AI"
	case "api":
		return "Type: API"
	case "bulk_reply":
		return "Type: Bulk reply"
	case "csat":
		return "Type: CSAT"
	case "smart_csat":
		return "Type: Smart CSAT"
	default:
		// The remaining documented values (application, integration, macro,
		// rule) and any type Front adds later read fine capitalised.
		return "Type: " + strings.ToUpper(accountType[:1]) + accountType[1:]
	}
}

// frontAccountType maps Front's teammate type to the account taxonomy. The
// documented enum is fully enumerated here so a value Front adds later lands on
// USER — the conservative side, since a wrongly-labelled service account drops
// a real person out of the human-review path.
func frontAccountType(accountType string) coredata.AccessReviewEntryAccountType {
	switch strings.ToLower(strings.TrimSpace(accountType)) {
	case "ai", "api", "application", "bulk_reply", "csat", "integration", "macro", "rule", "smart_csat":
		return coredata.AccessReviewEntryAccountTypeServiceAccount
	default:
		return coredata.AccessReviewEntryAccountTypeUser
	}
}

// frontAuthMethod reports how the account authenticates. Front states nothing
// about how a human teammate signs in (SSO vs password), so that stays UNKNOWN;
// a bot teammate acts on behalf of a rule, macro or OAuth client rather than a
// person, which is exactly SERVICE_ACCOUNT.
func frontAuthMethod(accountType string) coredata.AccessReviewEntryAuthMethod {
	if frontAccountType(accountType) == coredata.AccessReviewEntryAccountTypeServiceAccount {
		return coredata.AccessReviewEntryAuthMethodServiceAccount
	}

	return coredata.AccessReviewEntryAuthMethodUnknown
}

// frontNameResolver names the source after the Front company the credential
// belongs to, from GET /me.
type frontNameResolver struct {
	httpClient *http.Client
	baseURL    string
}

var _ NameResolver = (*frontNameResolver)(nil)

func NewFrontNameResolver(httpClient *http.Client, baseURL string) NameResolver {
	return &frontNameResolver{httpClient: httpClient, baseURL: baseURL}
}

func (r *frontNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	endpoint, err := url.JoinPath(r.baseURL, frontMePath)
	if err != nil {
		return "", fmt.Errorf("cannot build front identity URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create front identity request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute front identity request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode < http.StatusOK || httpResp.StatusCode >= http.StatusMultipleChoices {
		return "", nameStatusError("front identity", httpResp.StatusCode)
	}

	var resp struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode front identity response: %w", err)
	}

	return resp.Name, nil
}
