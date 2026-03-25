// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

// Package googleworkspace provides a Google Workspace identity provider
// for SCIM synchronization using OAuth2.
package googleworkspace

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"

	scimclient "go.probo.inc/probo/pkg/iam/scim/bridge/client"
	"go.probo.inc/probo/pkg/iam/scim/bridge/provider"
)

var _ provider.Provider = (*Provider)(nil)

type Provider struct {
	httpClient        *http.Client
	excludedUserNames []string
}

func New(httpClient *http.Client, excludedUserNames []string) *Provider {
	return &Provider{
		httpClient:        httpClient,
		excludedUserNames: excludedUserNames,
	}
}

func (p *Provider) Name() string {
	return "google-workspace"
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

func (p *Provider) ListUsers(ctx context.Context) (scimclient.Users, error) {
	adminService, err := admin.NewService(ctx, option.WithHTTPClient(p.httpClient))
	if err != nil {
		return nil, fmt.Errorf("cannot create admin service: %w", err)
	}

	var allUsers scimclient.Users
	pageToken := ""

	for {
		call := adminService.Users.List().Customer("my_customer").MaxResults(500).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("cannot list users: %w", err)
		}

		for _, u := range resp.Users {
			if p.isExcluded(u.PrimaryEmail) {
				continue
			}

			user := scimclient.User{
				UserName:    u.PrimaryEmail,
				DisplayName: u.Name.FullName,
				GivenName:   u.Name.GivenName,
				FamilyName:  u.Name.FamilyName,
				Active:      !u.Suspended && !u.Archived,
				ExternalID:  u.Id,
			}

			p.extractOrganizationFields(u.Organizations, &user)
			p.extractEmployeeNumber(u.ExternalIds, &user)
			p.extractRelations(u.Relations, &user)
			p.extractPreferredLanguage(u.Languages, &user)

			allUsers = append(allUsers, user)
		}

		pageToken = resp.NextPageToken
		if pageToken == "" {
			break
		}
	}

	return allUsers, nil
}

func (p *Provider) extractOrganizationFields(raw any, user *scimclient.User) {
	if raw == nil {
		return
	}

	data, err := json.Marshal(raw)
	if err != nil {
		return
	}

	var orgs []admin.UserOrganization
	if err := json.Unmarshal(data, &orgs); err != nil {
		return
	}

	var primary *admin.UserOrganization
	var first *admin.UserOrganization
	for i := range orgs {
		if first == nil {
			first = &orgs[i]
		}
		if orgs[i].Primary {
			primary = &orgs[i]
			break
		}
	}

	org := primary
	if org == nil {
		org = first
	}
	if org == nil {
		return
	}

	user.Title = org.Title
	user.Department = org.Department
	user.CostCenter = org.CostCenter
	user.EnterpriseOrganization = org.Name
	user.UserType = org.Description
}

func (p *Provider) extractEmployeeNumber(raw any, user *scimclient.User) {
	if raw == nil {
		return
	}

	data, err := json.Marshal(raw)
	if err != nil {
		return
	}

	var ids []admin.UserExternalId
	if err := json.Unmarshal(data, &ids); err != nil {
		return
	}

	for _, id := range ids {
		if id.Type == "organization" && id.Value != "" {
			user.EmployeeNumber = id.Value
			return
		}
	}

	if len(ids) > 0 && ids[0].Value != "" {
		user.EmployeeNumber = ids[0].Value
	}
}

func (p *Provider) extractRelations(raw any, user *scimclient.User) {
	if raw == nil {
		return
	}

	data, err := json.Marshal(raw)
	if err != nil {
		return
	}

	var relations []admin.UserRelation
	if err := json.Unmarshal(data, &relations); err != nil {
		return
	}

	for _, rel := range relations {
		if rel.Type == "manager" && rel.Value != "" {
			user.ManagerValue = rel.Value
			return
		}
	}
}

func (p *Provider) extractPreferredLanguage(raw any, user *scimclient.User) {
	if raw == nil {
		return
	}

	data, err := json.Marshal(raw)
	if err != nil {
		return
	}

	var languages []admin.UserLanguage
	if err := json.Unmarshal(data, &languages); err != nil {
		return
	}

	for _, lang := range languages {
		if lang.Preference == "preferred" && lang.LanguageCode != "" {
			user.PreferredLanguage = lang.LanguageCode
			return
		}
	}

	if len(languages) > 0 && languages[0].LanguageCode != "" {
		user.PreferredLanguage = languages[0].LanguageCode
	}
}
