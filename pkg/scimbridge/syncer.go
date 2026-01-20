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
package scimbridge

import (
	"context"
	"fmt"
	"strings"

	"go.gearno.de/kit/log"

	"go.probo.inc/probo/pkg/scimbridge/provider"
	"go.probo.inc/probo/pkg/scimbridge/scim"
)

type (
	Syncer struct {
		provider    provider.Provider
		scimClient  *scim.Client
		forceUpdate bool
		dryRun      bool
		logger      *log.Logger
	}

	Option func(*Syncer)
)

func WithDryRun(dryRun bool) Option {
	return func(s *Syncer) {
		s.dryRun = dryRun
	}
}

func WithForceUpdate(forceUpdate bool) Option {
	return func(s *Syncer) {
		s.forceUpdate = forceUpdate
	}
}

func NewSyncer(logger *log.Logger, provider provider.Provider, scimClient *scim.Client, opts ...Option) *Syncer {
	s := &Syncer{
		logger:     logger,
		provider:   provider,
		scimClient: scimClient,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Syncer) Run(ctx context.Context) (created, updated, deactivated, skipped, errors int, err error) {
	s.logger.InfoCtx(ctx, "starting SCIM bridge sync",
		log.String("provider", s.provider.Name()),
		log.Bool("dry_run", s.dryRun),
		log.Bool("force_update", s.forceUpdate),
	)

	providerUsers, err := s.provider.ListUsers(ctx)
	if err != nil {
		return 0, 0, 0, 0, 0, fmt.Errorf("cannot list provider users: %w", err)
	}
	s.logger.InfoCtx(ctx, "fetched users from provider",
		log.String("provider", s.provider.Name()),
		log.Int("count", len(providerUsers)),
	)

	scimUsers, err := s.scimClient.ListUsers(ctx)
	if err != nil {
		return 0, 0, 0, 0, 0, fmt.Errorf("cannot list scim users: %w", err)
	}
	s.logger.InfoCtx(ctx, "fetched existing SCIM users",
		log.Int("count", len(scimUsers)),
	)

	scimUsersByEmail := make(map[string]*scim.User)
	for i := range scimUsers {
		email := strings.ToLower(scimUsers[i].UserName)
		scimUsersByEmail[email] = &scimUsers[i]
	}

	providerEmails := make(map[string]bool)

	for _, pu := range providerUsers {
		email := strings.ToLower(pu.UserName)
		providerEmails[email] = true

		existingSCIM, exists := scimUsersByEmail[email]
		if !exists {
			if !s.dryRun {
				if err := s.scimClient.CreateUser(ctx, &pu); err != nil {
					s.logger.ErrorCtx(ctx, "cannot create user",
						log.String("email", pu.UserName),
						log.Error(err),
					)
					errors++
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
						s.logger.ErrorCtx(ctx, "cannot update user",
							log.String("email", pu.UserName),
							log.Error(err),
						)
						errors++
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

		if !scimUser.Active {
			continue
		}

		if !s.dryRun {
			if err := s.scimClient.DeactivateUser(ctx, scimUser.ID); err != nil {
				s.logger.ErrorCtx(ctx, "cannot deactivate user",
					log.String("email", email),
					log.Error(err),
				)
				errors++
				continue
			}
		}
		deactivated++
	}

	s.logger.InfoCtx(ctx, "sync completed",
		log.Int("created", created),
		log.Int("updated", updated),
		log.Int("deactivated", deactivated),
		log.Int("skipped", skipped),
		log.Int("errors", errors),
	)

	return created, updated, deactivated, skipped, errors, nil
}
