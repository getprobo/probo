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

	"go.gearno.de/kit/log"
	scimclient "go.probo.inc/probo/pkg/iam/scim/bridge/client"
	"go.probo.inc/probo/pkg/iam/scim/bridge/provider"
)

type (
	Result struct {
		Created     int
		Updated     int
		Deleted     int
		Deactivated int
		Skipped     int
		PhotoETags  map[string]string
	}

	Bridge struct {
		provider          provider.Provider
		scimClient        *scimclient.Client
		excludedUserNames []string
		logger            *log.Logger
	}

	Option func(*Bridge)
)

func WithExcludedUserNames(excludedUserNames []string) Option {
	return func(s *Bridge) {
		s.excludedUserNames = excludedUserNames
	}
}

func WithLogger(logger *log.Logger) Option {
	return func(s *Bridge) {
		s.logger = logger
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

func (s *Bridge) Run(ctx context.Context, photoETags map[string]string) (Result, error) {
	providerUsers, err := s.provider.ListUsers(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("cannot list provider users: %w", err)
	}

	scimUsers, err := s.scimClient.ListUsers(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("cannot list scim users: %w", err)
	}

	result := Result{PhotoETags: make(map[string]string)}

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
			s.loadPhoto(ctx, &pu, "")

			profileID, err := s.scimClient.CreateUser(ctx, &pu)
			if err != nil {
				errs = append(errs, fmt.Errorf("cannot create user %q: %w", pu.ExternalID, err))
				continue
			}

			result.Created++
			result.PhotoETags[profileID] = pu.PhotoETag

			continue
		}

		s.loadPhoto(ctx, &pu, photoETags[existingSCIM.ID])

		needsUpdate := len(pu.Photo) > 0 ||
			existingSCIM.Active != pu.Active ||
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
			existingSCIM.PreferredLanguage != pu.PreferredLanguage ||
			existingSCIM.UserType != pu.UserType

		if !needsUpdate {
			result.Skipped++
			continue
		}

		if err := s.scimClient.UpdateUser(ctx, existingSCIM.ID, &pu); err != nil {
			errs = append(errs, fmt.Errorf("cannot update user %q: %w", pu.ExternalID, err))
			continue
		}

		result.Updated++
		result.PhotoETags[existingSCIM.ID] = pu.PhotoETag
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

			result.Deleted++

			continue
		}

		if !scimUser.Active {
			continue
		}

		if err := s.scimClient.DeactivateUser(ctx, scimUser.ID); err != nil {
			errs = append(errs, fmt.Errorf("cannot deactivate user %q: %w", scimUser.ExternalID, err))
			continue
		}

		result.Deactivated++
	}

	return result, errors.Join(errs...)
}

func (s *Bridge) loadPhoto(ctx context.Context, user *scimclient.User, previousETag string) {
	if !user.PhotoETagKnown {
		user.PhotoETag = previousETag

		return
	}

	if user.PhotoETag == previousETag {
		return
	}

	if user.PhotoETag == "" {
		user.Photo = nil

		return
	}

	photo, err := s.provider.UserPhoto(ctx, *user)
	if err != nil || (len(photo.Body) == 0 && !photo.Absent) {
		reason := "empty photo"
		if err != nil {
			reason = "photo fetch failed"
		}

		if s.logger != nil {
			s.logger.WarnCtx(
				ctx,
				"skipping directory photo",
				log.String("provider", s.provider.Name()),
				log.String("external_id", user.ExternalID),
				log.String("reason", reason),
			)
		}

		user.PhotoETag = previousETag
		user.Photo = nil

		return
	}

	if photo.Absent {
		user.PhotoETag = ""
		user.Photo = nil

		return
	}

	user.Photo = photo.Body
	user.PhotoMIME = photo.MIME
}

func (s *Bridge) isExcluded(email string) bool {
	for _, excluded := range s.excludedUserNames {
		if strings.EqualFold(excluded, email) {
			return true
		}
	}

	return false
}
