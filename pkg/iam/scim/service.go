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

package scim

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/elimity-com/scim"
	"github.com/elimity-com/scim/optional"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
)

type (
	Service struct {
		pg     *pg.Client
		logger *log.Logger
	}
)

func NewService(
	pg *pg.Client,
	logger *log.Logger,
) *Service {
	return &Service{
		pg:     pg,
		logger: logger,
	}
}

func HashToken(token string) []byte {
	hash := sha256.Sum256([]byte(token))
	return hash[:]
}

func GenerateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("cannot generate random token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// ValidateToken validates a bearer token and returns the SCIM configuration
func (s *Service) ValidateToken(ctx context.Context, token string) (*coredata.SCIMConfiguration, error) {
	hashedToken := HashToken(token)
	config := &coredata.SCIMConfiguration{}

	err := s.pg.WithConn(ctx, func(conn pg.Conn) error {
		err := config.LoadByHashedToken(ctx, conn, hashedToken)
		if err != nil {
			if err == coredata.ErrResourceNotFound {
				return NewSCIMInvalidTokenError()
			}
			return fmt.Errorf("cannot load SCIM configuration: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return config, nil
}

// CreateUser creates a new user via SCIM provisioning
func (s *Service) CreateUser(
	ctx context.Context,
	config *coredata.SCIMConfiguration,
	attributes scim.ResourceAttributes,
	ipAddress net.IP,
) (*coredata.Membership, error) {
	user := ParseUserFromAttributes(attributes)
	email := user.GetPrimaryEmail()
	if email == "" {
		return nil, NewSCIMInvalidRequestError("userName or email is required")
	}

	emailAddr, err := mail.ParseAddr(email)
	if err != nil {
		return nil, NewSCIMInvalidRequestError("invalid email format")
	}

	fullName := user.GetFullName()
	now := time.Now()

	var membership *coredata.Membership

	scope := coredata.NewScopeFromObjectID(config.OrganizationID)

	err = s.pg.WithTx(ctx, func(tx pg.Conn) error {
		// Check if identity exists
		identity := &coredata.Identity{}
		err := identity.LoadByEmail(ctx, tx, emailAddr)

		if err == coredata.ErrResourceNotFound {
			// Create new identity
			identity = &coredata.Identity{
				ID:                   gid.New(gid.NilTenant, coredata.IdentityEntityType),
				EmailAddress:         emailAddr,
				FullName:             fullName,
				HashedPassword:       nil,
				EmailAddressVerified: false,
				CreatedAt:            now,
				UpdatedAt:            now,
			}

			err = identity.Insert(ctx, tx)
			if err != nil {
				return fmt.Errorf("cannot insert identity: %w", err)
			}
		} else if err != nil {
			return fmt.Errorf("cannot load identity: %w", err)
		}

		// Check if membership exists
		membership = &coredata.Membership{}
		err = membership.LoadByIdentityAndOrg(ctx, tx, scope, identity.ID, config.OrganizationID)

		if err == coredata.ErrResourceNotFound {
			// Create new membership
			membership = &coredata.Membership{
				ID:             gid.New(config.OrganizationID.TenantID(), coredata.MembershipEntityType),
				IdentityID:     identity.ID,
				OrganizationID: config.OrganizationID,
				Role:           coredata.MembershipRoleViewer,
				Source:         coredata.MembershipSourceSCIM,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			err = membership.Insert(ctx, tx, scope)
			if err != nil {
				return fmt.Errorf("cannot insert membership: %w", err)
			}

			// Create membership profile
			membershipProfile := &coredata.MembershipProfile{
				ID:           gid.New(membership.ID.TenantID(), coredata.MembershipProfileEntityType),
				MembershipID: membership.ID,
				FullName:     fullName,
				CreatedAt:    now,
				UpdatedAt:    now,
			}

			err = membershipProfile.Insert(ctx, tx)
			if err != nil {
				return fmt.Errorf("cannot insert membership profile: %w", err)
			}
		} else if err != nil {
			return fmt.Errorf("cannot load membership: %w", err)
		} else {
			// Update existing membership source to SCIM
			membership.Source = coredata.MembershipSourceSCIM
			membership.UpdatedAt = now

			err = membership.Update(ctx, tx, scope)
			if err != nil {
				return fmt.Errorf("cannot update membership: %w", err)
			}
		}

		// Log SCIM event
		event := s.createEvent(config, "POST", "/Users", membership.ID, ipAddress, 201, nil)
		err = event.Insert(ctx, tx, scope)
		if err != nil {
			s.logger.ErrorCtx(ctx, "cannot log SCIM event", log.Error(err))
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return membership, nil
}

// GetUser gets a user by membership ID
func (s *Service) GetUser(
	ctx context.Context,
	config *coredata.SCIMConfiguration,
	membershipID gid.GID,
	ipAddress net.IP,
) (*coredata.Membership, *coredata.Identity, *coredata.MembershipProfile, error) {
	scope := coredata.NewScopeFromObjectID(config.OrganizationID)

	var membership *coredata.Membership
	var identity *coredata.Identity
	var profile *coredata.MembershipProfile

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			membership = &coredata.Membership{}
			err := membership.LoadByID(ctx, conn, scope, membershipID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewSCIMUserNotFoundError(membershipID)
				}
				return fmt.Errorf("cannot load membership: %w", err)
			}

			// Verify membership belongs to this organization
			if membership.OrganizationID != config.OrganizationID {
				return NewSCIMUserNotFoundError(membershipID)
			}

			identity = &coredata.Identity{}
			err = identity.LoadByID(ctx, conn, membership.IdentityID)
			if err != nil {
				return fmt.Errorf("cannot load identity: %w", err)
			}

			profile = &coredata.MembershipProfile{}
			err = profile.LoadByMembershipID(ctx, conn, scope, membershipID)
			if err != nil && err != coredata.ErrResourceNotFound {
				return fmt.Errorf("cannot load membership profile: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, nil, nil, err
	}

	return membership, identity, profile, nil
}

// ListUsers lists all users in an organization, with optional filter support
func (s *Service) ListUsers(
	ctx context.Context,
	config *coredata.SCIMConfiguration,
	filter *UserFilter,
	startIndex int,
	count int,
	ipAddress net.IP,
) ([]*coredata.Membership, int, error) {
	scope := coredata.NewScopeFromObjectID(config.OrganizationID)

	var memberships coredata.Memberships
	var totalCount int

	err := s.pg.WithConn(ctx, func(conn pg.Conn) error {
		// If we have a userName filter, query by email directly
		if filter != nil && filter.UserName != nil {
			emailAddr, err := mail.ParseAddr(*filter.UserName)
			if err != nil {
				// Invalid email format - return empty result
				totalCount = 0
				return nil
			}

			membership := &coredata.Membership{}
			err = membership.LoadByEmailAndOrganization(ctx, conn, scope, emailAddr, config.OrganizationID)
			if err == coredata.ErrResourceNotFound {
				totalCount = 0
				return nil
			}
			if err != nil {
				return fmt.Errorf("cannot load membership by email: %w", err)
			}

			memberships = append(memberships, membership)
			totalCount = 1
			return nil
		}

		// No filter - return all memberships with pagination
		var err error
		totalCount, err = memberships.CountByOrganizationID(ctx, conn, scope, config.OrganizationID)
		if err != nil {
			return fmt.Errorf("cannot count memberships: %w", err)
		}

		orderBy := page.OrderBy[coredata.MembershipOrderField]{
			Field:     coredata.MembershipOrderFieldCreatedAt,
			Direction: page.OrderDirectionDesc,
		}
		cursor := page.NewCursor(count, nil, page.Head, orderBy)

		err = memberships.LoadByOrganizationID(ctx, conn, scope, config.OrganizationID, cursor)
		if err != nil {
			return fmt.Errorf("cannot load memberships: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, 0, err
	}

	return memberships, totalCount, nil
}

// ReplaceUser replaces a user via SCIM PUT
// Returns the membership, a boolean indicating if user was deactivated, and an error
func (s *Service) ReplaceUser(
	ctx context.Context,
	config *coredata.SCIMConfiguration,
	membershipID gid.GID,
	attributes scim.ResourceAttributes,
	ipAddress net.IP,
) (*coredata.Membership, bool, error) {
	user := ParseUserFromReplaceAttributes(attributes)
	return s.updateUser(ctx, config, membershipID, user, "PUT", ipAddress)
}

// PatchUser patches a user via SCIM PATCH
// Returns the membership, a boolean indicating if user was deactivated, and an error
func (s *Service) PatchUser(
	ctx context.Context,
	config *coredata.SCIMConfiguration,
	membershipID gid.GID,
	operations []scim.PatchOperation,
	ipAddress net.IP,
) (*coredata.Membership, bool, error) {
	user := ParseUserFromPatchOperations(operations)
	return s.updateUser(ctx, config, membershipID, user, "PATCH", ipAddress)
}

func (s *Service) updateUser(
	ctx context.Context,
	config *coredata.SCIMConfiguration,
	membershipID gid.GID,
	user *User,
	method string,
	ipAddress net.IP,
) (*coredata.Membership, bool, error) {
	scope := coredata.NewScopeFromObjectID(config.OrganizationID)
	now := time.Now()

	var membership *coredata.Membership
	var deactivated bool

	err := s.pg.WithTx(ctx, func(tx pg.Conn) error {
		membership = &coredata.Membership{}
		err := membership.LoadByID(ctx, tx, scope, membershipID)
		if err != nil {
			if err == coredata.ErrResourceNotFound {
				return NewSCIMUserNotFoundError(membershipID)
			}
			return fmt.Errorf("cannot load membership: %w", err)
		}

		// Verify membership belongs to this organization
		if membership.OrganizationID != config.OrganizationID {
			return NewSCIMUserNotFoundError(membershipID)
		}

		// Handle deactivation - Okta sends PATCH with active=false to deprovision users
		if user.Active != nil && !*user.Active {
			err = membership.Delete(ctx, tx, scope, membershipID)
			if err != nil {
				return fmt.Errorf("cannot delete membership: %w", err)
			}

			deactivated = true

			// Log SCIM event for deactivation
			event := s.createEvent(config, method, fmt.Sprintf("/Users/%s", membershipID), membershipID, ipAddress, 200, nil)
			err = event.Insert(ctx, tx, scope)
			if err != nil {
				s.logger.ErrorCtx(ctx, "cannot log SCIM event", log.Error(err))
			}

			return nil
		}

		// Update membership source to SCIM if not already
		if membership.Source != coredata.MembershipSourceSCIM {
			membership.Source = coredata.MembershipSourceSCIM
			membership.UpdatedAt = now

			err = membership.Update(ctx, tx, scope)
			if err != nil {
				return fmt.Errorf("cannot update membership: %w", err)
			}
		}

		// Update membership profile
		profile := &coredata.MembershipProfile{}
		err = profile.LoadByMembershipID(ctx, tx, scope, membershipID)
		if err == nil {
			fullName := user.GetFullName()
			if fullName != "" {
				profile.FullName = fullName
				profile.UpdatedAt = now

				err = profile.Update(ctx, tx, scope)
				if err != nil {
					return fmt.Errorf("cannot update membership profile: %w", err)
				}
			}
		}

		// Log SCIM event
		event := s.createEvent(config, method, fmt.Sprintf("/Users/%s", membershipID), membership.ID, ipAddress, 200, nil)
		err = event.Insert(ctx, tx, scope)
		if err != nil {
			s.logger.ErrorCtx(ctx, "cannot log SCIM event", log.Error(err))
		}

		return nil
	})

	if err != nil {
		return nil, false, err
	}

	return membership, deactivated, nil
}

// DeleteUser removes a user's membership from the organization
func (s *Service) DeleteUser(
	ctx context.Context,
	config *coredata.SCIMConfiguration,
	membershipID gid.GID,
	ipAddress net.IP,
) error {
	scope := coredata.NewScopeFromObjectID(config.OrganizationID)

	return s.pg.WithTx(ctx, func(tx pg.Conn) error {
		membership := &coredata.Membership{}
		err := membership.LoadByID(ctx, tx, scope, membershipID)
		if err != nil {
			if err == coredata.ErrResourceNotFound {
				return NewSCIMUserNotFoundError(membershipID)
			}
			return fmt.Errorf("cannot load membership: %w", err)
		}

		// Verify membership belongs to this organization
		if membership.OrganizationID != config.OrganizationID {
			return NewSCIMUserNotFoundError(membershipID)
		}

		err = membership.Delete(ctx, tx, scope, membershipID)
		if err != nil {
			return fmt.Errorf("cannot delete membership: %w", err)
		}

		// Log SCIM event
		event := s.createEvent(config, "DELETE", fmt.Sprintf("/Users/%s", membershipID), membershipID, ipAddress, 204, nil)
		err = event.Insert(ctx, tx, scope)
		if err != nil {
			s.logger.ErrorCtx(ctx, "cannot log SCIM event", log.Error(err))
		}

		return nil
	})
}

// LogEvent logs a SCIM event
func (s *Service) LogEvent(
	ctx context.Context,
	config *coredata.SCIMConfiguration,
	method string,
	path string,
	membershipID *gid.GID,
	ipAddress net.IP,
	statusCode int,
	errorMessage *string,
) {
	scope := coredata.NewScopeFromObjectID(config.OrganizationID)

	var mID gid.GID
	if membershipID != nil {
		mID = *membershipID
	}

	event := s.createEvent(config, method, path, mID, ipAddress, statusCode, errorMessage)

	err := s.pg.WithConn(ctx, func(conn pg.Conn) error {
		return event.Insert(ctx, conn, scope)
	})

	if err != nil {
		s.logger.ErrorCtx(ctx, "cannot log SCIM event", log.Error(err))
	}
}

func (s *Service) createEvent(
	config *coredata.SCIMConfiguration,
	method string,
	path string,
	membershipID gid.GID,
	ipAddress net.IP,
	statusCode int,
	errorMessage *string,
) *coredata.SCIMEvent {
	event := &coredata.SCIMEvent{
		ID:                  gid.New(config.OrganizationID.TenantID(), coredata.SCIMEventEntityType),
		OrganizationID:      config.OrganizationID,
		SCIMConfigurationID: config.ID,
		Method:              method,
		Path:                path,
		StatusCode:          statusCode,
		ErrorMessage:        errorMessage,
		IPAddress:           ipAddress,
		CreatedAt:           time.Now(),
	}

	if membershipID != gid.Nil {
		event.MembershipID = &membershipID
	}

	return event
}

// ParseUserFromAttributes extracts a User from SCIM resource attributes
func ParseUserFromAttributes(attributes scim.ResourceAttributes) *User {
	userName, _ := attributes["userName"].(string)
	displayName, _ := attributes["displayName"].(string)

	var givenName, familyName string
	if name, ok := attributes["name"].(map[string]interface{}); ok {
		givenName, _ = name["givenName"].(string)
		familyName, _ = name["familyName"].(string)
	}

	// Get email from emails array or use userName
	email := userName
	if emails, ok := attributes["emails"].([]interface{}); ok && len(emails) > 0 {
		for _, e := range emails {
			if emailMap, ok := e.(map[string]interface{}); ok {
				if primary, _ := emailMap["primary"].(bool); primary {
					if value, ok := emailMap["value"].(string); ok {
						email = value
						break
					}
				}
			}
		}
		// If no primary email found, use the first one
		if email == userName {
			if emailMap, ok := emails[0].(map[string]interface{}); ok {
				if value, ok := emailMap["value"].(string); ok {
					email = value
				}
			}
		}
	}

	// Build full name
	fullName := displayName
	if fullName == "" {
		fullName = strings.TrimSpace(givenName + " " + familyName)
	}
	if fullName == "" {
		fullName = userName
	}

	user := &User{
		UserName:    userName,
		DisplayName: displayName,
		Name: &Name{
			GivenName:  givenName,
			FamilyName: familyName,
			Formatted:  fullName,
		},
		Emails: []Email{
			{
				Value:   email,
				Primary: true,
			},
		},
	}

	return user
}

// ParseUserFromReplaceAttributes extracts a User from SCIM replace attributes
func ParseUserFromReplaceAttributes(attributes scim.ResourceAttributes) *User {
	displayName, _ := attributes["displayName"].(string)

	var givenName, familyName string
	if name, ok := attributes["name"].(map[string]interface{}); ok {
		givenName, _ = name["givenName"].(string)
		familyName, _ = name["familyName"].(string)
	}

	fullName := displayName
	if fullName == "" {
		fullName = strings.TrimSpace(givenName + " " + familyName)
	}

	active := true
	if a, ok := attributes["active"].(bool); ok {
		active = a
	}

	return &User{
		DisplayName: fullName,
		Active:      &active,
		Name: &Name{
			GivenName:  givenName,
			FamilyName: familyName,
			Formatted:  fullName,
		},
	}
}

// ParseUserFromPatchOperations extracts a User from SCIM patch operations
func ParseUserFromPatchOperations(operations []scim.PatchOperation) *User {
	user := &User{}
	for _, op := range operations {
		if strings.EqualFold(op.Op, "replace") || strings.EqualFold(op.Op, "add") {
			path := ""
			if op.Path != nil {
				path = op.Path.String()
			}
			switch strings.ToLower(path) {
			case "active":
				if active, ok := op.Value.(bool); ok {
					user.Active = &active
				}
			case "displayname":
				if name, ok := op.Value.(string); ok {
					user.DisplayName = name
				}
			case "name.givenname":
				if user.Name == nil {
					user.Name = &Name{}
				}
				if name, ok := op.Value.(string); ok {
					user.Name.GivenName = name
				}
			case "name.familyname":
				if user.Name == nil {
					user.Name = &Name{}
				}
				if name, ok := op.Value.(string); ok {
					user.Name.FamilyName = name
				}
			}
		}
	}
	return user
}

// MembershipToResource converts a Membership to a SCIM resource
func MembershipToResource(m *coredata.Membership) scim.Resource {
	return MembershipToResourceWithActive(m, true)
}

// MembershipToResourceWithActive converts a Membership to a SCIM resource with a custom active state
func MembershipToResourceWithActive(m *coredata.Membership, active bool) scim.Resource {
	created := m.CreatedAt
	modified := m.UpdatedAt
	return scim.Resource{
		ID:         m.ID.String(),
		ExternalID: optional.NewString(m.ID.String()),
		Attributes: scim.ResourceAttributes{
			"userName":    m.EmailAddress.String(),
			"displayName": m.FullName,
			"active":      active,
			"name": map[string]interface{}{
				"formatted": m.FullName,
			},
			"emails": []map[string]interface{}{
				{
					"value":   m.EmailAddress.String(),
					"type":    "work",
					"primary": true,
				},
			},
		},
		Meta: scim.Meta{
			Created:      &created,
			LastModified: &modified,
		},
	}
}

// MembershipToResourceFull converts a Membership with full identity and profile to a SCIM resource
func MembershipToResourceFull(m *coredata.Membership, identity *coredata.Identity, profile *coredata.MembershipProfile) scim.Resource {
	fullName := identity.FullName
	if profile != nil && profile.FullName != "" {
		fullName = profile.FullName
	}

	created := m.CreatedAt
	modified := m.UpdatedAt
	return scim.Resource{
		ID:         m.ID.String(),
		ExternalID: optional.NewString(m.ID.String()),
		Attributes: scim.ResourceAttributes{
			"userName":    identity.EmailAddress.String(),
			"displayName": fullName,
			"active":      true,
			"name": map[string]interface{}{
				"formatted": fullName,
			},
			"emails": []map[string]interface{}{
				{
					"value":   identity.EmailAddress.String(),
					"type":    "work",
					"primary": true,
				},
			},
		},
		Meta: scim.Meta{
			Created:      &created,
			LastModified: &modified,
		},
	}
}
