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

const deepgramAPIBaseURL = "https://api.deepgram.com"

// DeepgramDriver lists the members of every Deepgram project the API key
// can access. The key (presented in the `Authorization: Token <key>`
// scheme by the connection transport) is scoped to a single account, whose
// members may span several projects; the driver aggregates them and dedupes
// by member_id, unioning each member's per-project scopes.
type DeepgramDriver struct {
	httpClient *http.Client
}

var _ Driver = (*DeepgramDriver)(nil)

type deepgramProject struct {
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
}

type deepgramProjectsResponse struct {
	Projects []deepgramProject `json:"projects"`
}

type deepgramMember struct {
	MemberID  string   `json:"member_id"`
	Email     string   `json:"email"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Scopes    []string `json:"scopes"`
}

type deepgramMembersResponse struct {
	Members []deepgramMember `json:"members"`
}

func NewDeepgramDriver(httpClient *http.Client) *DeepgramDriver {
	return &DeepgramDriver{httpClient: httpClient}
}

func (d *DeepgramDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	projects, err := d.fetchProjects(ctx)
	if err != nil {
		return nil, err
	}

	// Aggregate members across projects, preserving first-seen order and
	// unioning the scopes a member holds in each project.
	order := make([]string, 0)
	merged := make(map[string]*deepgramMember)

	for _, project := range projects {
		members, err := d.fetchProjectMembers(ctx, project.ProjectID)
		if err != nil {
			return nil, err
		}

		for _, m := range members {
			existing, ok := merged[m.MemberID]
			if !ok {
				copied := m
				merged[m.MemberID] = &copied
				order = append(order, m.MemberID)

				continue
			}

			existing.Scopes = deepgramUnionScopes(existing.Scopes, m.Scopes)
		}
	}

	records := make([]AccountRecord, 0, len(order))

	for _, id := range order {
		m := merged[id]

		email := strings.TrimSpace(m.Email)
		if email == "" {
			continue
		}

		records = append(records, AccountRecord{
			Email:       email,
			FullName:    deepgramFullName(*m, email),
			Roles:       deepgramRoles(m.Scopes),
			IsAdmin:     deepgramIsAdmin(m.Scopes),
			MFAStatus:   coredata.MFAStatusUnknown,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
			ExternalID:  strings.TrimSpace(m.MemberID),
		})
	}

	return records, nil
}

func (d *DeepgramDriver) fetchProjects(ctx context.Context) ([]deepgramProject, error) {
	endpoint, err := url.JoinPath(deepgramAPIBaseURL, "v1", "projects")
	if err != nil {
		return nil, fmt.Errorf("cannot build deepgram projects URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create deepgram projects request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute deepgram projects request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch deepgram projects: unexpected status %d", httpResp.StatusCode)
	}

	var resp deepgramProjectsResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode deepgram projects response: %w", err)
	}

	return resp.Projects, nil
}

func (d *DeepgramDriver) fetchProjectMembers(ctx context.Context, projectID string) ([]deepgramMember, error) {
	endpoint, err := url.JoinPath(deepgramAPIBaseURL, "v1", "projects", url.PathEscape(projectID), "members")
	if err != nil {
		return nil, fmt.Errorf("cannot build deepgram members URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create deepgram members request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	httpResp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute deepgram members request: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch deepgram members: unexpected status %d", httpResp.StatusCode)
	}

	var resp deepgramMembersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode deepgram members response: %w", err)
	}

	return resp.Members, nil
}

func deepgramFullName(m deepgramMember, fallback string) string {
	fullName := strings.TrimSpace(strings.TrimSpace(m.FirstName) + " " + strings.TrimSpace(m.LastName))
	if fullName != "" {
		return fullName
	}

	return fallback
}

func deepgramUnionScopes(a, b []string) []string {
	seen := make(map[string]bool, len(a))
	out := make([]string, 0, len(a)+len(b))

	for _, scopes := range [][]string{a, b} {
		for _, s := range scopes {
			if seen[s] {
				continue
			}

			seen[s] = true

			out = append(out, s)
		}
	}

	return out
}

// deepgramRoles derives a role label from a member's scopes. Deepgram has no
// dedicated role field; ownership/administration is expressed through the
// scope list. Unknown scope sets fall back to "Member".
func deepgramRoles(scopes []string) []string {
	switch {
	case deepgramHasScope(scopes, "owner"):
		return []string{"Owner"}
	case deepgramHasScope(scopes, "admin"):
		return []string{"Admin"}
	default:
		return []string{"Member"}
	}
}

func deepgramIsAdmin(scopes []string) bool {
	return deepgramHasScope(scopes, "owner") || deepgramHasScope(scopes, "admin")
}

func deepgramHasScope(scopes []string, want string) bool {
	for _, s := range scopes {
		if strings.EqualFold(strings.TrimSpace(s), want) {
			return true
		}
	}

	return false
}
