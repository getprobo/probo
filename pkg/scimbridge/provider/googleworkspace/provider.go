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
// for SCIM synchronization.
package googleworkspace

import (
	"context"
	"fmt"

	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"

	"go.probo.inc/probo/pkg/scimbridge/scim"
)

type (
	Provider struct {
		serviceAccountKey []byte
		adminEmail        string
	}
)

func New(serviceAccountKey []byte, adminEmail string) *Provider {
	return &Provider{
		serviceAccountKey: serviceAccountKey,
		adminEmail:        adminEmail,
	}
}

func (p *Provider) Name() string {
	return "google-workspace"
}

func (p *Provider) ListUsers(ctx context.Context) ([]scim.User, error) {
	config, err := google.JWTConfigFromJSON(p.serviceAccountKey, admin.AdminDirectoryUserReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("cannot create JWT config: %w", err)
	}

	config.Subject = p.adminEmail

	adminService, err := admin.NewService(ctx, option.WithHTTPClient(config.Client(ctx)))
	if err != nil {
		return nil, fmt.Errorf("cannot create admin service: %w", err)
	}

	var allUsers []scim.User
	pageToken := ""

	for {
		call := adminService.Users.List().Customer("my_customer").MaxResults(500)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("cannot list users: %w", err)
		}

		for _, u := range resp.Users {
			allUsers = append(
				allUsers,
				scim.User{
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
