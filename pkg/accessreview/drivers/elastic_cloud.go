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
)

type (
	ElasticCloudDriver struct {
		httpClient     *http.Client
		organizationID string
		baseURL        string
	}

	elasticCloudMembersResponse struct {
		Members []elasticCloudMember `json:"members"`
	}

	elasticCloudMember struct {
		UserID          string                      `json:"user_id"`
		Name            string                      `json:"name"`
		Email           string                      `json:"email"`
		RoleAssignments elasticCloudRoleAssignments `json:"role_assignments"`
	}

	elasticCloudRoleAssignment struct {
		RoleID string `json:"role_id"`
	}

	elasticCloudRoleAssignments struct {
		Platform     []elasticCloudRoleAssignment `json:"platform"`
		Organization []elasticCloudRoleAssignment `json:"organization"`
		Deployment   []elasticCloudRoleAssignment `json:"deployment"`
		Project      struct {
			Elasticsearch []elasticCloudRoleAssignment `json:"elasticsearch"`
			Observability []elasticCloudRoleAssignment `json:"observability"`
			Security      []elasticCloudRoleAssignment `json:"security"`
			WorkplaceAI   []elasticCloudRoleAssignment `json:"workplaceai"`
		} `json:"project"`
	}
)

var _ Driver = (*ElasticCloudDriver)(nil)

func NewElasticCloudDriver(httpClient *http.Client, organizationID, baseURL string) *ElasticCloudDriver {
	return &ElasticCloudDriver{
		httpClient:     httpClient,
		organizationID: organizationID,
		baseURL:        baseURL,
	}
}

func (d *ElasticCloudDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	endpoint, err := url.JoinPath(
		d.baseURL,
		"organizations",
		url.PathEscape(d.organizationID),
		"members",
	)
	if err != nil {
		return nil, fmt.Errorf("cannot build elastic cloud members URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create elastic cloud members request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute elastic cloud members request: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch elastic cloud members: unexpected status %d", resp.StatusCode)
	}

	var payload elasticCloudMembersResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("cannot decode elastic cloud members response: %w", err)
	}

	records := make([]AccountRecord, 0, len(payload.Members))
	for _, member := range payload.Members {
		if member.Email == "" {
			continue
		}

		roles := elasticCloudRoles(member.RoleAssignments)
		records = append(records, AccountRecord{
			Email:      member.Email,
			FullName:   member.Name,
			Roles:      roles,
			IsAdmin:    new(elasticCloudRolesIncludeAdmin(roles)),
			ExternalID: member.UserID,
		})
	}

	return records, nil
}

// Role labels retain both scope and role ID; project labels also retain type.
func elasticCloudRoles(assignments elasticCloudRoleAssignments) []string {
	roles := make([]string, 0)
	roles = appendElasticCloudRoles(roles, "platform", assignments.Platform)
	roles = appendElasticCloudRoles(roles, "organization", assignments.Organization)
	roles = appendElasticCloudRoles(roles, "deployment", assignments.Deployment)
	roles = appendElasticCloudRoles(roles, "project:elasticsearch", assignments.Project.Elasticsearch)
	roles = appendElasticCloudRoles(roles, "project:observability", assignments.Project.Observability)
	roles = appendElasticCloudRoles(roles, "project:security", assignments.Project.Security)
	roles = appendElasticCloudRoles(roles, "project:workplaceai", assignments.Project.WorkplaceAI)

	return roles
}

func appendElasticCloudRoles(
	roles []string,
	scope string,
	assignments []elasticCloudRoleAssignment,
) []string {
	for _, assignment := range assignments {
		if assignment.RoleID != "" {
			roles = append(roles, scope+":"+assignment.RoleID)
		}
	}

	return roles
}

func elasticCloudRolesIncludeAdmin(roles []string) bool {
	for _, role := range roles {
		_, roleID, ok := strings.Cut(role, ":")
		if !ok {
			continue
		}

		roleID = strings.ToLower(roleID)
		if roleID == "admin" || strings.HasSuffix(roleID, "-admin") {
			return true
		}
	}

	return false
}

type elasticCloudNameResolver struct {
	httpClient     *http.Client
	organizationID string
	baseURL        string
}

func NewElasticCloudNameResolver(httpClient *http.Client, organizationID, baseURL string) NameResolver {
	return &elasticCloudNameResolver{
		httpClient:     httpClient,
		organizationID: organizationID,
		baseURL:        baseURL,
	}
}

func (r *elasticCloudNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	if r.organizationID == "" {
		return "", nil
	}

	endpoint, err := url.JoinPath(
		r.baseURL,
		"organizations",
		url.PathEscape(r.organizationID),
	)
	if err != nil {
		return "", fmt.Errorf("cannot build elastic cloud organization URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create elastic cloud organization request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute elastic cloud organization request: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", nil
	}

	var organization struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&organization); err != nil {
		return "", fmt.Errorf("cannot decode elastic cloud organization response: %w", err)
	}

	return organization.Name, nil
}
