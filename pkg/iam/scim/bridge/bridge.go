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

// Package scimbridge provides a bridge for synchronizing users from identity
// providers to SCIM-compliant systems.
package bridge

import (
	"context"
	"errors"
	"fmt"
	"strings"

	scimclient "go.probo.inc/probo/pkg/iam/scim/bridge/client"
	"go.probo.inc/probo/pkg/iam/scim/bridge/provider"
)

type (
	Bridge struct {
		provider          provider.Provider
		scimClient        *scimclient.Client
		excludedUserNames []string
	}

	Option func(*Bridge)
)

func WithExcludedUserNames(excludedUserNames []string) Option {
	return func(s *Bridge) {
		s.excludedUserNames = excludedUserNames
	}
}

func NewBridge(provider provider.Provider, scimClient *scimclient.Client, opts ...Option) *Bridge {
	s := &Bridge{
		provider:   provider,
		scimClient: scimClient,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Bridge) Run(ctx context.Context) (created, updated, deleted, deactivated, skipped int, err error) {
	providerUsers, err := s.provider.ListUsers(ctx)
	if err != nil {
		return 0, 0, 0, 0, 0, fmt.Errorf("cannot list provider users: %w", err)
	}

	scimUsers, err := s.scimClient.ListUsers(ctx)
	if err != nil {
		return 0, 0, 0, 0, 0, fmt.Errorf("cannot list scim users: %w", err)
	}

	scimUsersByEmail := make(map[string]*scimclient.User)

	for i := range scimUsers {
		email := strings.ToLower(scimUsers[i].UserName)
		scimUsersByEmail[email] = &scimUsers[i]
	}

	providerEmails := make(map[string]bool)

	var errs []error

	for _, pu := range providerUsers {
		email := strings.ToLower(pu.UserName)
		providerEmails[email] = true

		existingSCIM, exists := scimUsersByEmail[email]
		if !exists {
			if err := s.scimClient.CreateUser(ctx, &pu); err != nil {
				errs = append(errs, fmt.Errorf("cannot create user %q: %w", pu.ExternalID, err))
				continue
			}

			created++
		} else {
			needsUpdate := existingSCIM.Active != pu.Active ||
				existingSCIM.DisplayName != pu.DisplayName ||
				existingSCIM.Title != pu.Title ||
				existingSCIM.GivenName != pu.GivenName ||
				existingSCIM.FamilyName != pu.FamilyName ||
				existingSCIM.ExternalID != pu.ExternalID ||
				existingSCIM.Department != pu.Department ||
				existingSCIM.CostCenter != pu.CostCenter ||
				existingSCIM.EnterpriseOrganization != pu.EnterpriseOrganization ||
				existingSCIM.Division != pu.Division ||
				existingSCIM.EmployeeNumber != pu.EmployeeNumber ||
				existingSCIM.ManagerValue != pu.ManagerValue ||
				existingSCIM.PreferredLanguage != pu.PreferredLanguage

			if needsUpdate {
				if err := s.scimClient.UpdateUser(ctx, existingSCIM.ID, &pu); err != nil {
					errs = append(errs, fmt.Errorf("cannot update user %q: %w", pu.ExternalID, err))
					continue
				}

				updated++
			} else {
				skipped++
			}
		}
	}

	for email, scimUser := range scimUsersByEmail {
		if providerEmails[email] {
			continue
		}

		if s.isExcluded(email) {
			if err := s.scimClient.DeleteUser(ctx, scimUser.ID); err != nil {
				errs = append(errs, fmt.Errorf("cannot delete user %q: %w", scimUser.ExternalID, err))
				continue
			}

			deleted++

			continue
		}

		if !scimUser.Active {
			continue
		}

		if err := s.scimClient.DeactivateUser(ctx, scimUser.ID); err != nil {
			errs = append(errs, fmt.Errorf("cannot deactivate user %q: %w", scimUser.ExternalID, err))
			continue
		}

		deactivated++
	}

	return created, updated, deleted, deactivated, skipped, errors.Join(errs...)
}

func (s *Bridge) isExcluded(email string) bool {
	for _, excluded := range s.excludedUserNames {
		if strings.EqualFold(excluded, email) {
			return true
		}
	}

	return false
}
