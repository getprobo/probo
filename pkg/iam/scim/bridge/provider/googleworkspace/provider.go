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

// Package googleworkspace provides a Google Workspace identity provider
// for SCIM synchronization using OAuth2.
package googleworkspace

import (
	"context"
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

			allUsers = append(
				allUsers,
				scimclient.User{
					UserName:    u.PrimaryEmail,
					DisplayName: u.Name.FullName,
					GivenName:   u.Name.GivenName,
					FamilyName:  u.Name.FamilyName,
					Active:      !u.Suspended && !u.Archived,
				},
			)
		}

		pageToken = resp.NextPageToken
		if pageToken == "" {
			break
		}
	}

	return allUsers, nil
}
