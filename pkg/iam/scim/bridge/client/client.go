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

package scimclient

import (
	"bytes"
	"context"
	"encoding/base64"
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
		PhotoETag              string `json:"-"`
		// PhotoETagKnown reports that PhotoETag was read from the directory.
		// An empty PhotoETag is a removed picture only when this is true.
		// A false value means the version was not read, and the bridge keeps
		// the stored picture.
		PhotoETagKnown bool   `json:"-"`
		Photo          []byte `json:"-"`
		PhotoMIME      string `json:"-"`
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

// PhotoExtensionURN carries photoETag. RFC 7643 does not define that
// attribute on the core User, so it travels in the Probo extension.
const PhotoExtensionURN = "urn:ietf:params:scim:schemas:extension:probo:2.0:User"

func buildUserPayload(user *User) map[string]any {
	schemas := []string{
		"urn:ietf:params:scim:schemas:core:2.0:User",
		"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
		PhotoExtensionURN,
	}

	enterprise := map[string]any{
		"employeeNumber": user.EmployeeNumber,
		"department":     user.Department,
		"costCenter":     user.CostCenter,
		"organization":   user.EnterpriseOrganization,
		"division":       user.Division,
		"manager":        map[string]string{"value": user.ManagerValue},
	}

	payload := map[string]any{
		"schemas":  schemas,
		"userName": user.UserName,
		"name": map[string]string{
			"givenName":  user.GivenName,
			"familyName": user.FamilyName,
			"formatted":  user.DisplayName,
		},
		"displayName":       user.DisplayName,
		"active":            user.Active,
		"title":             user.Title,
		"externalId":        user.ExternalID,
		"userType":          user.UserType,
		"preferredLanguage": user.PreferredLanguage,
		"emails": []map[string]any{
			{
				"value":   user.UserName,
				"type":    "work",
				"primary": true,
			},
		},
		"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": enterprise,
		PhotoExtensionURN: map[string]any{
			"photoETag": user.PhotoETag,
		},
	}

	if len(user.Photo) > 0 {
		payload["photos"] = []map[string]any{
			{
				"value":   photoDataURI(user.PhotoMIME, user.Photo),
				"type":    "photo",
				"primary": true,
			},
		}
	}

	return payload
}

func photoDataURI(mime string, body []byte) string {
	if mime != "image/png" {
		mime = "image/jpeg"
	}

	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(body)
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

	var photo struct {
		PhotoETag string `json:"photoETag"`
	}

	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(data, &envelope); err != nil {
		return err
	}

	if extension, ok := envelope[PhotoExtensionURN]; ok {
		if err := json.Unmarshal(extension, &photo); err != nil {
			return err
		}
	}

	u.ID = raw.ID
	u.UserName = raw.UserName
	u.DisplayName = raw.DisplayName
	u.Active = raw.Active
	u.Title = raw.Title
	u.ExternalID = raw.ExternalID
	u.UserType = raw.UserType
	u.PreferredLanguage = raw.PreferredLanguage
	u.PhotoETag = photo.PhotoETag
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
