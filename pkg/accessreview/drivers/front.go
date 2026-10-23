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

const (
	frontTeammatesPath = "/teammates"
	frontMePath        = "/me"
)

// FrontDriver lists a Front company's teammates. GET /teammates omits blocked
// teammates and lists pending invitations as ordinary teammates; Front exposes
// no MFA, last-login or creation date.
type FrontDriver struct {
	httpClient *http.Client
	baseURL    string
}

var _ Driver = (*FrontDriver)(nil)

type (
	frontTeammatesResponse struct {
		Pagination *struct {
			Next *string `json:"next"`
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

func NewFrontDriver(httpClient *http.Client, baseURL string) *FrontDriver {
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

// ListAccounts fetches the roster in a single request: GET /teammates is
// unpaginated. A next-page cursor is refused rather than ignored, so a roster
// Front starts paginating is never reported as complete.
func (d *FrontDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	endpoint, err := url.JoinPath(d.baseURL, frontTeammatesPath)
	if err != nil {
		return nil, fmt.Errorf("cannot build front teammates URL: %w", err)
	}

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

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cannot fetch front teammates: unexpected status %d", httpResp.StatusCode)
	}

	var resp frontTeammatesResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode front teammates response: %w", err)
	}

	if resp.Results == nil {
		return nil, errors.New("cannot list front teammates: response has no teammates array")
	}

	if resp.Pagination != nil && resp.Pagination.Next != nil && *resp.Pagination.Next != "" {
		return nil, errors.New("cannot list front teammates: response is paginated")
	}

	records := make([]AccountRecord, 0, len(resp.Results))
	for _, teammate := range resp.Results {
		if teammate.ID == "" {
			return nil, errors.New("cannot list front teammates: teammate has no id")
		}

		records = append(records, frontAccountRecord(teammate))
	}

	return records, nil
}

func frontAccountRecord(teammate frontTeammate) AccountRecord {
	accountType := frontAccountType(teammate.Type)

	authMethod := coredata.AccessReviewEntryAuthMethodUnknown
	if accountType == coredata.AccessReviewEntryAccountTypeServiceAccount {
		authMethod = coredata.AccessReviewEntryAuthMethodServiceAccount
	}

	return AccountRecord{
		Email:       teammate.Email,
		FullName:    frontFullName(teammate),
		Roles:       frontRoles(teammate),
		Active:      new(!teammate.IsBlocked),
		IsAdmin:     new(teammate.IsAdmin),
		MFAStatus:   coredata.MFAStatusUnknown,
		AuthMethod:  authMethod,
		AccountType: accountType,
		ExternalID:  teammate.ID,
	}
}

// frontFullName falls back to the username: invited teammates and bots carry
// no first or last name.
func frontFullName(teammate frontTeammate) string {
	name := strings.TrimSpace(teammate.FirstName + " " + teammate.LastName)
	if name != "" {
		return name
	}

	return teammate.Username
}

func frontRoles(teammate frontTeammate) []string {
	if teammate.IsAdmin {
		return []string{"Admin"}
	}

	return []string{"Teammate"}
}

// frontAccountType maps the documented bot types to service accounts; any other
// value, including one Front adds later, stays a user.
func frontAccountType(teammateType string) coredata.AccessReviewEntryAccountType {
	switch teammateType {
	case "ai", "api", "application", "bulk_reply", "csat", "integration", "macro", "rule", "smart_csat":
		return coredata.AccessReviewEntryAccountTypeServiceAccount
	default:
		return coredata.AccessReviewEntryAccountTypeUser
	}
}

type frontNameResolver struct {
	httpClient *http.Client
	baseURL    string
}

var _ NameResolver = (*frontNameResolver)(nil)

func NewFrontNameResolver(httpClient *http.Client, baseURL string) NameResolver {
	return &frontNameResolver{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

// ResolveInstanceName returns the company name from GET /me.
func (r *frontNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	endpoint, err := url.JoinPath(r.baseURL, frontMePath)
	if err != nil {
		return "", fmt.Errorf("cannot build front company URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create front company request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute front company request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode != http.StatusOK {
		return "", nameStatusError("front company", httpResp.StatusCode)
	}

	var resp struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("cannot decode front company response: %w", err)
	}

	return resp.Name, nil
}
