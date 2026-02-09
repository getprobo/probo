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
		forceUpdate       bool
		dryRun            bool
	}

	Option func(*Bridge)
)

func WithDryRun(dryRun bool) Option {
	return func(s *Bridge) {
		s.dryRun = dryRun
	}
}

func WithForceUpdate(forceUpdate bool) Option {
	return func(s *Bridge) {
		s.forceUpdate = forceUpdate
	}
}

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
			if !s.dryRun {
				if err := s.scimClient.CreateUser(ctx, &pu); err != nil {
					errs = append(errs, fmt.Errorf("cannot create user %q: %w", pu.UserName, err))
					continue
				}
			}
			created++
		} else {
			needsUpdate := s.forceUpdate

			if existingSCIM.Active != pu.Active {
				needsUpdate = true
			}
			if existingSCIM.DisplayName != pu.DisplayName {
				needsUpdate = true
			}

			if needsUpdate {
				if !s.dryRun {
					if err := s.scimClient.UpdateUser(ctx, existingSCIM.ID, &pu); err != nil {
						errs = append(errs, fmt.Errorf("cannot update user %q: %w", pu.UserName, err))
						continue
					}
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
			if !s.dryRun {
				if err := s.scimClient.DeleteUser(ctx, scimUser.ID); err != nil {
					errs = append(errs, fmt.Errorf("cannot delete user %q: %w", email, err))
					continue
				}
			}
			deleted++
			continue
		}

		if !scimUser.Active {
			continue
		}

		if !s.dryRun {
			if err := s.scimClient.DeactivateUser(ctx, scimUser.ID); err != nil {
				errs = append(errs, fmt.Errorf("cannot deactivate user %q: %w", email, err))
				continue
			}
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
