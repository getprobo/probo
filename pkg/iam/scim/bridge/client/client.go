// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

package scimclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type (
	Client struct {
		endpoint   string
		token      string
		httpClient *http.Client
	}

	User struct {
		ID                     string `json:"id,omitempty"`
		UserName               string `json:"userName"`
		DisplayName            string `json:"displayName"`
		GivenName              string `json:"-"`
		FamilyName             string `json:"-"`
		Active                 bool   `json:"active"`
		Title                  string `json:"title"`
		ExternalID             string `json:"-"`
		Department             string `json:"-"`
		CostCenter             string `json:"-"`
		EnterpriseOrganization string `json:"-"`
		Division               string `json:"-"`
		EmployeeNumber         string `json:"-"`
		UserType               string `json:"-"`
		ManagerValue           string `json:"-"`
		PreferredLanguage      string `json:"-"`
	}

	Users []User

	ListResponse struct {
		Schemas      []string `json:"schemas"`
		TotalResults int      `json:"totalResults"`
		StartIndex   int      `json:"startIndex"`
		ItemsPerPage int      `json:"itemsPerPage"`
		Resources    Users    `json:"Resources"`
	}
)

func NewClient(httpClient *http.Client, endpoint, token string) *Client {
	return &Client{
		endpoint:   strings.TrimSuffix(endpoint, "/"),
		token:      token,
		httpClient: httpClient,
	}
}

func (c *Client) ListUsers(ctx context.Context) (Users, error) {
	var allUsers Users
	startIndex := 1
	count := 100

	for {
		users, total, err := c.listUsersPage(ctx, startIndex, count)
		if err != nil {
			return nil, err
		}

		allUsers = append(allUsers, users...)

		if len(allUsers) >= total {
			break
		}

		startIndex += count
	}

	return allUsers, nil
}

func (c *Client) listUsersPage(ctx context.Context, startIndex, count int) (Users, int, error) {
	reqURL := fmt.Sprintf("%s/Users?startIndex=%d&count=%d", c.endpoint, startIndex, count)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("cannot create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("cannot fetch users: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, 0, fmt.Errorf("SCIM API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var listResp ListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, 0, fmt.Errorf("cannot decode response: %w", err)
	}

	return listResp.Resources, listResp.TotalResults, nil
}

func (c *Client) CreateUser(ctx context.Context, user *User) error {
	payload := buildUserPayload(user)

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("cannot marshal user: %w", err)
	}

	reqURL := fmt.Sprintf("%s/Users", c.endpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("cannot create request: %w", err)
	}

	c.setHeaders(req)
	req.Header.Set("Content-Type", "application/scim+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot create user: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("SCIM API error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func (c *Client) UpdateUser(ctx context.Context, userID string, user *User) error {
	payload := buildUserPayload(user)

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("cannot marshal user: %w", err)
	}

	reqURL := fmt.Sprintf("%s/Users/%s", c.endpoint, url.PathEscape(userID))
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, reqURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("cannot create request: %w", err)
	}

	c.setHeaders(req)
	req.Header.Set("Content-Type", "application/scim+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot update user: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("SCIM API error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func buildUserPayload(user *User) map[string]any {
	schemas := []string{"urn:ietf:params:scim:schemas:core:2.0:User"}

	enterprise := map[string]any{}
	if user.EmployeeNumber != "" {
		enterprise["employeeNumber"] = user.EmployeeNumber
	}
	if user.Department != "" {
		enterprise["department"] = user.Department
	}
	if user.CostCenter != "" {
		enterprise["costCenter"] = user.CostCenter
	}
	if user.EnterpriseOrganization != "" {
		enterprise["organization"] = user.EnterpriseOrganization
	}
	if user.Division != "" {
		enterprise["division"] = user.Division
	}
	if user.ManagerValue != "" {
		enterprise["manager"] = map[string]string{"value": user.ManagerValue}
	}

	payload := map[string]any{
		"schemas":  schemas,
		"userName": user.UserName,
		"name": map[string]string{
			"givenName":  user.GivenName,
			"familyName": user.FamilyName,
			"formatted":  user.DisplayName,
		},
		"displayName": user.DisplayName,
		"active":      user.Active,
		"emails": []map[string]any{
			{
				"value":   user.UserName,
				"type":    "work",
				"primary": true,
			},
		},
		"title": user.Title,
	}

	if user.ExternalID != "" {
		payload["externalId"] = user.ExternalID
	}
	if user.UserType != "" {
		payload["userType"] = user.UserType
	}
	if user.PreferredLanguage != "" {
		payload["preferredLanguage"] = user.PreferredLanguage
	}

	if len(enterprise) > 0 {
		schemas = append(schemas, "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User")
		payload["schemas"] = schemas
		payload["urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"] = enterprise
	}

	return payload
}

func (c *Client) DeactivateUser(ctx context.Context, userID string) error {
	payload := map[string]any{
		"schemas": []string{"urn:ietf:params:scim:api:messages:2.0:PatchOp"},
		"Operations": []map[string]any{
			{
				"op":    "replace",
				"path":  "active",
				"value": false,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("cannot marshal patch: %w", err)
	}

	reqURL := fmt.Sprintf("%s/Users/%s", c.endpoint, url.PathEscape(userID))
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, reqURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("cannot create request: %w", err)
	}

	c.setHeaders(req)
	req.Header.Set("Content-Type", "application/scim+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot deactivate user: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("SCIM API error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func (c *Client) DeleteUser(ctx context.Context, userID string) error {
	reqURL := fmt.Sprintf("%s/Users/%s", c.endpoint, url.PathEscape(userID))
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqURL, nil)
	if err != nil {
		return fmt.Errorf("cannot create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot delete user: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("SCIM API error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func (u *User) UnmarshalJSON(data []byte) error {
	var raw struct {
		ID                string `json:"id"`
		UserName          string `json:"userName"`
		DisplayName       string `json:"displayName"`
		Active            bool   `json:"active"`
		Title             string `json:"title"`
		ExternalID        string `json:"externalId"`
		UserType          string `json:"userType"`
		PreferredLanguage string `json:"preferredLanguage"`
		Name              struct {
			GivenName  string `json:"givenName"`
			FamilyName string `json:"familyName"`
		} `json:"name"`
		Enterprise struct {
			EmployeeNumber string `json:"employeeNumber"`
			Department     string `json:"department"`
			CostCenter     string `json:"costCenter"`
			Organization   string `json:"organization"`
			Division       string `json:"division"`
			Manager        struct {
				Value string `json:"value"`
			} `json:"manager"`
		} `json:"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	u.ID = raw.ID
	u.UserName = raw.UserName
	u.DisplayName = raw.DisplayName
	u.Active = raw.Active
	u.Title = raw.Title
	u.ExternalID = raw.ExternalID
	u.UserType = raw.UserType
	u.PreferredLanguage = raw.PreferredLanguage
	u.GivenName = raw.Name.GivenName
	u.FamilyName = raw.Name.FamilyName
	u.EmployeeNumber = raw.Enterprise.EmployeeNumber
	u.Department = raw.Enterprise.Department
	u.CostCenter = raw.Enterprise.CostCenter
	u.EnterpriseOrganization = raw.Enterprise.Organization
	u.Division = raw.Enterprise.Division
	u.ManagerValue = raw.Enterprise.Manager.Value

	return nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/scim+json")
}
