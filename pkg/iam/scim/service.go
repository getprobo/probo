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
	scimerrors "github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/optional"
	"github.com/prometheus/client_golang/prometheus"
	scimfilter "github.com/scim2/filter-parser/v2"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.opentelemetry.io/otel/trace"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
)

type (
	Service struct {
		pg           *pg.Client
		logger       *log.Logger
		bridgeRunner *BridgeRunner
	}

	ServiceConfig struct {
		TracerProvider    trace.TracerProvider
		Registerer        prometheus.Registerer
		EncryptionKey     cipher.EncryptionKey
		ConnectorRegistry *connector.ConnectorRegistry
		BridgeRunner      BridgeRunnerConfig
	}
)

func NewService(
	pg *pg.Client,
	logger *log.Logger,
	cfg ServiceConfig,
) *Service {
	bridgeRunner := NewBridgeRunner(
		pg,
		logger.Named("bridge-runner"),
		cfg.TracerProvider,
		cfg.Registerer,
		cfg.EncryptionKey,
		cfg.ConnectorRegistry,
		cfg.BridgeRunner,
	)

	return &Service{
		pg:           pg,
		logger:       logger,
		bridgeRunner: bridgeRunner,
	}
}

// Run starts the SCIM service background processes.
func (s *Service) Run(ctx context.Context) error {
	return s.bridgeRunner.Run(ctx)
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

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := config.LoadByHashedToken(ctx, conn, hashedToken)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewSCIMInvalidTokenError()
				}
				return fmt.Errorf("cannot load SCIM configuration: %w", err)
			}
			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return config, nil
}

func (s *Service) CreateUser(
	ctx context.Context,
	config *coredata.SCIMConfiguration,
	attributes scim.ResourceAttributes,
) (scim.Resource, error) {
	email, fullName, active := ParseUserFromAttributes(attributes)
	if email == "" {
		return scim.Resource{}, scimerrors.ScimErrorBadRequest("userName or email is required")
	}

	emailAddr, err := mail.ParseAddr(email)
	if err != nil {
		return scim.Resource{}, scimerrors.ScimErrorBadRequest("invalid email format")
	}
	now := time.Now()

	membershipState := coredata.MembershipStateActive
	if !active {
		membershipState = coredata.MembershipStateInactive
	}

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
				Role:           coredata.MembershipRoleEmployee,
				Source:         coredata.MembershipSourceSCIM,
				State:          membershipState,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			err = membership.Insert(ctx, tx, scope)
			if err != nil {
				return fmt.Errorf("cannot insert membership: %w", err)
			}

			// Create membership profile
			membershipProfile := &coredata.MembershipProfile{
				ID:             gid.New(membership.ID.TenantID(), coredata.MembershipProfileEntityType),
				IdentityID:     identity.ID,
				OrganizationID: config.OrganizationID,
				MembershipID:   membership.ID,
				FullName:       fullName,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			err = membershipProfile.Insert(ctx, tx)
			if err != nil {
				return fmt.Errorf("cannot insert membership profile: %w", err)
			}

			// Expire all pending invitations for email in organization
			invitations := &coredata.Invitations{}
			onlyPending := coredata.NewInvitationFilter([]coredata.InvitationStatus{coredata.InvitationStatusPending})
			err := invitations.ExpireByEmailAndOrganization(
				ctx,
				tx,
				coredata.NewScopeFromObjectID(config.OrganizationID),
				emailAddr,
				config.OrganizationID,
				onlyPending,
			)

			if err != nil {
				return fmt.Errorf("cannot expire pending invitations by email")
			}
		} else if err != nil {
			return fmt.Errorf("cannot load membership: %w", err)
		} else {
			// Update existing membership - follow what SCIM tells us
			membership.Source = coredata.MembershipSourceSCIM
			membership.State = membershipState
			membership.UpdatedAt = now

			err = membership.Update(ctx, tx, scope)
			if err != nil {
				return fmt.Errorf("cannot update membership: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return scim.Resource{}, err
	}

	return membershipToResource(membership), nil
}

func (s *Service) GetUser(
	ctx context.Context,
	config *coredata.SCIMConfiguration,
	membershipID gid.GID,
) (scim.Resource, error) {
	scope := coredata.NewScopeFromObjectID(config.OrganizationID)

	var membership *coredata.Membership

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			membership = &coredata.Membership{}
			err := membership.LoadByID(ctx, conn, scope, membershipID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return scimerrors.ScimErrorResourceNotFound(membershipID.String())
				}
				return fmt.Errorf("cannot load membership: %w", err)
			}

			if membership.OrganizationID != config.OrganizationID {
				return scimerrors.ScimErrorResourceNotFound(membershipID.String())
			}

			return nil
		},
	)

	if err != nil {
		return scim.Resource{}, err
	}

	return membershipToResource(membership), nil
}

func (s *Service) ListUsers(
	ctx context.Context,
	config *coredata.SCIMConfiguration,
	filterExpr scimfilter.Expression,
	startIndex int,
	count int,
) ([]scim.Resource, int, error) {
	filter, err := ParseUserFilter(filterExpr)
	if err != nil {
		return nil, 0, err
	}

	// Only return SCIM-managed users. This ensures that:
	// 1. Users created through other means (manual, SAML) are not deactivated
	//    when they don't exist in the identity provider.
	// 2. When a manual user exists in the identity provider but not in the
	//    SCIM list, CreateUser is called which enrolls them into SCIM management.
	filter.WithSource(coredata.MembershipSourceSCIM)

	scope := coredata.NewScopeFromObjectID(config.OrganizationID)

	var memberships coredata.Memberships
	var totalCount int

	err = s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			var err error
			totalCount, err = memberships.CountByOrganizationID(ctx, conn, scope, config.OrganizationID, filter)
			if err != nil {
				return fmt.Errorf("cannot count memberships: %w", err)
			}

			orderBy := page.OrderBy[coredata.MembershipOrderField]{
				Field:     coredata.MembershipOrderFieldCreatedAt,
				Direction: page.OrderDirectionDesc,
			}
			cursor := page.NewCursor(count, nil, page.Head, orderBy)

			err = memberships.LoadByOrganizationID(ctx, conn, scope, config.OrganizationID, cursor, filter)
			if err != nil {
				return fmt.Errorf("cannot load memberships: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, 0, err
	}

	resources := make([]scim.Resource, 0, len(memberships))
	for _, m := range memberships {
		resources = append(resources, membershipToResource(m))
	}

	return resources, totalCount, nil
}

func (s *Service) ReplaceUser(
	ctx context.Context,
	config *coredata.SCIMConfiguration,
	membershipID gid.GID,
	attributes scim.ResourceAttributes,
) (scim.Resource, error) {
	fullName, active := ParseUserFromReplaceAttributes(attributes)
	membership, err := s.updateUser(ctx, config, membershipID, fullName, active)
	if err != nil {
		return scim.Resource{}, err
	}

	return membershipToResource(membership), nil
}

func (s *Service) PatchUser(
	ctx context.Context,
	config *coredata.SCIMConfiguration,
	membershipID gid.GID,
	operations []scim.PatchOperation,
) (scim.Resource, error) {
	fullName, active := ParseUserFromPatchOperations(operations)
	membership, err := s.updateUser(ctx, config, membershipID, fullName, active)
	if err != nil {
		return scim.Resource{}, err
	}

	return membershipToResource(membership), nil
}

func (s *Service) updateUser(
	ctx context.Context,
	config *coredata.SCIMConfiguration,
	membershipID gid.GID,
	fullName string,
	active *bool,
) (*coredata.Membership, error) {
	scope := coredata.NewScopeFromObjectID(config.OrganizationID)
	now := time.Now()

	var membership *coredata.Membership

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			membership = &coredata.Membership{}
			if err := membership.LoadByID(ctx, tx, scope, membershipID); err != nil {
				if err == coredata.ErrResourceNotFound {
					return scimerrors.ScimErrorResourceNotFound(membershipID.String())
				}
				return fmt.Errorf("cannot load membership: %w", err)
			}

			if membership.OrganizationID != config.OrganizationID {
				return scimerrors.ScimErrorResourceNotFound(membershipID.String())
			}

			needsUpdate := false

			if active != nil {
				identity := &coredata.Identity{}
				if err := identity.LoadByID(ctx, tx, membership.IdentityID); err != nil {
					return fmt.Errorf("cannot load identity: %w", err)
				}

				if *active && membership.State == coredata.MembershipStateInactive {
					membership.State = coredata.MembershipStateActive
					membership.Role = coredata.MembershipRoleEmployee
					needsUpdate = true

					// Expire all pending invitations for email in organization
					invitations := &coredata.Invitations{}
					onlyPending := coredata.NewInvitationFilter([]coredata.InvitationStatus{coredata.InvitationStatusPending})
					if err := invitations.ExpireByEmailAndOrganization(
						ctx,
						tx,
						coredata.NewScopeFromObjectID(config.OrganizationID),
						identity.EmailAddress,
						config.OrganizationID,
						onlyPending,
					); err != nil {
						return fmt.Errorf("cannot expire pending invitations by email: %w", err)
					}
				} else if !*active && membership.State == coredata.MembershipStateActive {
					membership.State = coredata.MembershipStateInactive
					needsUpdate = true

					// Expire all pending invitations for email in organization
					invitations := &coredata.Invitations{}
					onlyPending := coredata.NewInvitationFilter([]coredata.InvitationStatus{coredata.InvitationStatusPending})
					if err := invitations.ExpireByEmailAndOrganization(
						ctx,
						tx,
						coredata.NewScopeFromObjectID(config.OrganizationID),
						identity.EmailAddress,
						config.OrganizationID,
						onlyPending,
					); err != nil {
						return fmt.Errorf("cannot expire pending invitations: %w", err)
					}
				}
			}

			if membership.Source != coredata.MembershipSourceSCIM {
				membership.Source = coredata.MembershipSourceSCIM
				needsUpdate = true
			}

			if needsUpdate {
				membership.UpdatedAt = now
				if err := membership.Update(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot update membership: %w", err)
				}
			}

			profile := &coredata.MembershipProfile{}
			if err := profile.LoadByMembershipID(ctx, tx, scope, membershipID); err == nil {
				if fullName != "" {
					profile.FullName = fullName
					profile.UpdatedAt = now

					if err := profile.Update(ctx, tx, scope); err != nil {
						return fmt.Errorf("cannot update membership profile: %w", err)
					}
				}
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return membership, nil
}

func (s *Service) DeleteUser(
	ctx context.Context,
	config *coredata.SCIMConfiguration,
	membershipID gid.GID,
) error {
	scope := coredata.NewScopeFromObjectID(config.OrganizationID)

	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			membership := &coredata.Membership{}
			err := membership.LoadByID(ctx, tx, scope, membershipID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return scimerrors.ScimErrorResourceNotFound(membershipID.String())
				}
				return fmt.Errorf("cannot load membership: %w", err)
			}

			if membership.OrganizationID != config.OrganizationID {
				return scimerrors.ScimErrorResourceNotFound(membershipID.String())
			}

			err = membership.Delete(ctx, tx, scope, membershipID)
			if err != nil {
				return fmt.Errorf("cannot delete membership: %w", err)
			}

			// Expire all pending invitations for email in organization
			identity := &coredata.Identity{}
			if err := identity.LoadByID(ctx, tx, membership.IdentityID); err != nil {
				return fmt.Errorf("cannot load identity: %w", err)
			}
			invitations := &coredata.Invitations{}
			onlyPending := coredata.NewInvitationFilter([]coredata.InvitationStatus{coredata.InvitationStatusPending})
			if err := invitations.ExpireByEmailAndOrganization(
				ctx,
				tx,
				coredata.NewScopeFromObjectID(config.OrganizationID),
				identity.EmailAddress,
				config.OrganizationID,
				onlyPending,
			); err != nil {
				return fmt.Errorf("cannot expire pending invitations: %w", err)
			}

			return nil
		},
	)
}

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

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := event.Insert(ctx, conn, scope)
			if err != nil {
				return fmt.Errorf("cannot insert SCIM event: %w", err)
			}
			return nil
		},
	)

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

func ParseUserFromAttributes(attributes scim.ResourceAttributes) (email string, fullName string, active bool) {
	userName, _ := attributes["userName"].(string)
	displayName, _ := attributes["displayName"].(string)

	// Default to active if the attribute is not present.
	active = true
	if a, ok := attributes["active"].(bool); ok {
		active = a
	}

	var givenName, familyName string
	if name, ok := attributes["name"].(map[string]any); ok {
		givenName, _ = name["givenName"].(string)
		familyName, _ = name["familyName"].(string)
	}

	// Get email from emails array or use userName
	email = userName
	if emails, ok := attributes["emails"].([]any); ok && len(emails) > 0 {
		for _, e := range emails {
			if emailMap, ok := e.(map[string]any); ok {
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
			if emailMap, ok := emails[0].(map[string]any); ok {
				if value, ok := emailMap["value"].(string); ok {
					email = value
				}
			}
		}
	}

	// Build full name: prefer displayName, then given+family, then userName
	fullName = displayName
	if fullName == "" {
		fullName = strings.TrimSpace(givenName + " " + familyName)
	}
	if fullName == "" {
		fullName = userName
	}

	return email, fullName, active
}

func ParseUserFromReplaceAttributes(attributes scim.ResourceAttributes) (fullName string, active *bool) {
	displayName, _ := attributes["displayName"].(string)

	var givenName, familyName string
	if name, ok := attributes["name"].(map[string]any); ok {
		givenName, _ = name["givenName"].(string)
		familyName, _ = name["familyName"].(string)
	}

	fullName = displayName
	if fullName == "" {
		fullName = strings.TrimSpace(givenName + " " + familyName)
	}

	activeVal := true
	if a, ok := attributes["active"].(bool); ok {
		activeVal = a
	}

	return fullName, &activeVal
}

func ParseUserFromPatchOperations(operations []scim.PatchOperation) (fullName string, active *bool) {
	var givenName, familyName string

	for _, op := range operations {
		if strings.EqualFold(op.Op, "replace") || strings.EqualFold(op.Op, "add") {
			path := ""
			if op.Path != nil {
				path = op.Path.String()
			}

			// Handle empty path with value map (Okta style)
			// e.g., { "op": "Replace", "value": { "active": false } }
			if path == "" {
				if valueMap, ok := op.Value.(map[string]any); ok {
					if a, ok := valueMap["active"].(bool); ok {
						active = &a
					}
					if name, ok := valueMap["displayName"].(string); ok {
						fullName = name
					}
					if nameMap, ok := valueMap["name"].(map[string]any); ok {
						if gn, ok := nameMap["givenName"].(string); ok {
							givenName = gn
						}
						if fn, ok := nameMap["familyName"].(string); ok {
							familyName = fn
						}
					}
				}
				continue
			}

			switch strings.ToLower(path) {
			case "active":
				if a, ok := op.Value.(bool); ok {
					active = &a
				}
			case "displayname":
				if name, ok := op.Value.(string); ok {
					fullName = name
				}
			case "name.givenname":
				if name, ok := op.Value.(string); ok {
					givenName = name
				}
			case "name.familyname":
				if name, ok := op.Value.(string); ok {
					familyName = name
				}
			}
		}
	}

	// If no displayName was set but we have name parts, build full name
	if fullName == "" && (givenName != "" || familyName != "") {
		fullName = strings.TrimSpace(givenName + " " + familyName)
	}

	return fullName, active
}

func membershipToResource(m *coredata.Membership) scim.Resource {
	return scim.Resource{
		ID:         m.ID.String(),
		ExternalID: optional.NewString(m.ID.String()),
		Attributes: scim.ResourceAttributes{
			"userName":    m.EmailAddress.String(),
			"displayName": m.FullName,
			"active":      m.State == coredata.MembershipStateActive,
			"name": map[string]any{
				"formatted": m.FullName,
			},
			"emails": []map[string]any{
				{
					"value":   m.EmailAddress.String(),
					"type":    "work",
					"primary": true,
				},
			},
		},
		Meta: scim.Meta{
			Created:      &m.CreatedAt,
			LastModified: &m.UpdatedAt,
		},
	}
}
