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

// Package microsoft365 provides a Microsoft 365 (Microsoft Entra ID)
// identity provider for SCIM synchronization using OAuth2 against the
// Microsoft Graph API.
package microsoft365

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	scimclient "go.probo.inc/probo/pkg/iam/scim/bridge/client"
	"go.probo.inc/probo/pkg/iam/scim/bridge/provider"
)

// graphBaseURL is the Microsoft Graph v1.0 endpoint. The bridge uses raw
// HTTP rather than the official SDK because the SDK pulls in a large
// dependency tree for what is a small set of read-only calls.
const graphBaseURL = "https://graph.microsoft.com/v1.0"

// graphUserSelect is the projection used when listing users. Limiting the
// response to only the fields we map keeps payloads small and avoids the
// extra permission scopes some properties would otherwise require.
const graphUserSelect = "id,userPrincipalName,mail,displayName,givenName,surname,accountEnabled,jobTitle,department,companyName,employeeId,preferredLanguage,usageLocation"

// graphPageSize is the maximum page size for /users. Microsoft Graph
// caps /users at 999.
const graphPageSize = 999

// graphMaxPages bounds pagination to prevent unbounded loops if the
// API misbehaves.
const graphMaxPages = 1000

// graphMemberFilter restricts /users to home-tenant members and
// excludes B2B guest users invited from external tenants — those guests
// should not be provisioned into the connected organization.
const graphMemberFilter = "userType eq 'Member'"

const (
	graphPhotoConcurrency      = 8
	graphPhotoAuthFailureLimit = 3
	maxPhotoBytes              = 5 << 20
)

var (
	_ provider.Provider    = (*Provider)(nil)
	_ provider.PhotoSource = (*Provider)(nil)
)

type Provider struct {
	httpClient        *http.Client
	excludedUserNames []string
	graphBaseURL      string

	// photoLookupConcurrency and photoAuthFailureLimit override the defaults
	// in tests. Zero uses the production limits.
	photoLookupConcurrency int
	photoAuthFailureLimit  int
}

func New(httpClient *http.Client, excludedUserNames []string) *Provider {
	return &Provider{
		httpClient:        httpClient,
		excludedUserNames: excludedUserNames,
	}
}

func (p *Provider) Name() string {
	return "microsoft-365"
}

func (p *Provider) isExcluded(email string) bool {
	emailLower := strings.ToLower(email)
	for _, excluded := range p.excludedUserNames {
		if strings.ToLower(excluded) == emailLower {
			return true
		}
	}

	return false
}

// graphUser is the subset of the Microsoft Graph user resource the
// bridge consumes. Managers are intentionally not fetched here: the
// /users endpoint does not return them and per-user lookups would
// multiply the call volume by the directory size. Manager data can be
// added later via /users/{id}/manager when needed.
type graphUser struct {
	ID                string `json:"id"`
	UserPrincipalName string `json:"userPrincipalName"`
	Mail              string `json:"mail"`
	DisplayName       string `json:"displayName"`
	GivenName         string `json:"givenName"`
	Surname           string `json:"surname"`
	AccountEnabled    bool   `json:"accountEnabled"`
	JobTitle          string `json:"jobTitle"`
	Department        string `json:"department"`
	CompanyName       string `json:"companyName"`
	EmployeeID        string `json:"employeeId"`
	PreferredLanguage string `json:"preferredLanguage"`
	UsageLocation     string `json:"usageLocation"`
}

type graphUsersResponse struct {
	Value    []graphUser `json:"value"`
	NextLink string      `json:"@odata.nextLink"`
}

func (p *Provider) baseURL() string {
	if p.graphBaseURL != "" {
		return p.graphBaseURL
	}

	return graphBaseURL
}

func (p *Provider) ListUsers(ctx context.Context) (scimclient.Users, error) {
	endpoint, err := p.listUsersURL()
	if err != nil {
		return nil, fmt.Errorf("cannot build list users URL: %w", err)
	}

	var allUsers scimclient.Users

	for range graphMaxPages {
		users, next, err := p.fetchPage(ctx, endpoint)
		if err != nil {
			return nil, err
		}

		for _, u := range users {
			email := u.Mail
			if email == "" {
				email = u.UserPrincipalName
			}

			if email == "" {
				continue
			}

			if p.isExcluded(email) {
				continue
			}

			allUsers = append(allUsers, scimclient.User{
				UserName:               email,
				DisplayName:            u.DisplayName,
				GivenName:              u.GivenName,
				FamilyName:             u.Surname,
				Active:                 u.AccountEnabled,
				ExternalID:             u.ID,
				Title:                  u.JobTitle,
				Department:             u.Department,
				EnterpriseOrganization: u.CompanyName,
				EmployeeNumber:         u.EmployeeID,
				PreferredLanguage:      u.PreferredLanguage,
			})
		}

		if next == "" {
			p.attachPhotoETags(ctx, allUsers)

			return allUsers, nil
		}

		endpoint = next
	}

	return nil, fmt.Errorf("microsoft 365: pagination limit of %d pages reached", graphMaxPages)
}

func (p *Provider) listUsersURL() (string, error) {
	joined, err := url.JoinPath(p.baseURL(), "users")
	if err != nil {
		return "", fmt.Errorf("cannot join graph users URL: %w", err)
	}

	u, err := url.Parse(joined)
	if err != nil {
		return "", fmt.Errorf("cannot parse graph users URL: %w", err)
	}

	q := u.Query()
	q.Set("$select", graphUserSelect)
	q.Set("$top", strconv.Itoa(graphPageSize))
	q.Set("$filter", graphMemberFilter)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func (p *Provider) fetchPage(ctx context.Context, endpoint string) ([]graphUser, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, "", fmt.Errorf("cannot create graph users request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("cannot list graph users: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("microsoft graph error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var page graphUsersResponse
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, "", fmt.Errorf("cannot decode graph users response: %w", err)
	}

	return page.Value, page.NextLink, nil
}

func (p *Provider) attachPhotoETags(ctx context.Context, users scimclient.Users) {
	var (
		wg       sync.WaitGroup
		sem      = make(chan struct{}, p.photoWorkers())
		disabled atomic.Bool
		authFail atomic.Int32
	)

	for i := range users {
		if users[i].ExternalID == "" || disabled.Load() || ctx.Err() != nil {
			continue
		}

		select {
		case <-ctx.Done():
			wg.Wait()

			return
		case sem <- struct{}{}:
		}

		wg.Add(1)

		go func(i int) {
			defer wg.Done()
			defer func() { <-sem }()

			if disabled.Load() || ctx.Err() != nil {
				return
			}

			etag, status, err := p.photoETag(ctx, users[i].ExternalID)
			if err != nil {
				return
			}

			// Only a photo response or a confirmed 404 records a version.
			// A transport error, an empty etag, or any other status stays
			// unknown so the bridge keeps the stored picture.
			switch status {
			case http.StatusOK:
				if etag == "" {
					return
				}

				users[i].PhotoETag = etag
				users[i].PhotoETagKnown = true
			case http.StatusNotFound:
				users[i].PhotoETag = ""
				users[i].PhotoETagKnown = true
			case http.StatusUnauthorized, http.StatusForbidden:
				// One forbidden user can be a per-user restriction. Repeated
				// auth failures mean the token cannot read photos, so the
				// remaining lookups stop.
				if authFail.Add(1) >= int32(p.authFailureLimit()) {
					disabled.Store(true)
				}
			}
		}(i)
	}

	wg.Wait()
}

func (p *Provider) photoWorkers() int {
	if p.photoLookupConcurrency > 0 {
		return p.photoLookupConcurrency
	}

	return graphPhotoConcurrency
}

func (p *Provider) authFailureLimit() int {
	if p.photoAuthFailureLimit > 0 {
		return p.photoAuthFailureLimit
	}

	return graphPhotoAuthFailureLimit
}

func (p *Provider) photoETag(ctx context.Context, userID string) (string, int, error) {
	endpoint, err := p.userPhotoURL(userID, false)
	if err != nil {
		return "", 0, fmt.Errorf("cannot build graph photo URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", 0, fmt.Errorf("cannot create graph photo request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("cannot fetch graph photo: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return "", resp.StatusCode, nil
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", resp.StatusCode, nil
	}

	var meta struct {
		MediaETag string `json:"@odata.mediaEtag"`
	}

	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&meta); err != nil {
		return "", resp.StatusCode, fmt.Errorf("cannot decode graph photo: %w", err)
	}

	return meta.MediaETag, resp.StatusCode, nil
}

func (p *Provider) UserPhoto(ctx context.Context, user scimclient.User) (provider.Photo, error) {
	if user.ExternalID == "" {
		return provider.Photo{}, nil
	}

	endpoint, err := p.userPhotoURL(user.ExternalID, true)
	if err != nil {
		return provider.Photo{}, fmt.Errorf("cannot build graph photo URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return provider.Photo{}, fmt.Errorf("cannot create graph photo request: %w", err)
	}

	req.Header.Set("Accept", "image/jpeg, image/png, */*")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return provider.Photo{}, fmt.Errorf("cannot fetch graph photo: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return provider.Photo{Absent: true}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return provider.Photo{}, fmt.Errorf("cannot fetch graph photo: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxPhotoBytes+1))
	if err != nil {
		return provider.Photo{}, fmt.Errorf("cannot read graph photo: %w", err)
	}

	if len(body) == 0 || len(body) > maxPhotoBytes {
		return provider.Photo{}, fmt.Errorf("cannot read graph photo: unexpected size")
	}

	return provider.Photo{
		ETag: user.PhotoETag,
		Body: body,
		MIME: photoMIME(resp.Header.Get("Content-Type")),
	}, nil
}

func (p *Provider) userPhotoURL(userID string, content bool) (string, error) {
	parts := []string{"users", url.PathEscape(userID), "photo"}
	if content {
		parts = append(parts, "$value")
	}

	joined, err := url.JoinPath(p.baseURL(), parts...)
	if err != nil {
		return "", fmt.Errorf("cannot join graph photo URL: %w", err)
	}

	return joined, nil
}

func photoMIME(header string) string {
	mime := header
	if i := strings.Index(mime, ";"); i >= 0 {
		mime = mime[:i]
	}

	mime = strings.ToLower(strings.TrimSpace(mime))
	if mime == "image/png" {
		return mime
	}

	return "image/jpeg"
}
