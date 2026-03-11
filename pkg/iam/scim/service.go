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
	"errors"
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
	"go.gearno.de/x/ref"
	"go.opentelemetry.io/otel/trace"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/webhook"
	webhooktypes "go.probo.inc/probo/pkg/webhook/types"
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
	attrs := ParseUserFromAttributes(attributes)
	if attrs.UserName == "" {
		return scim.Resource{}, scimerrors.ScimErrorBadRequest("userName is required")
	}
	if attrs.Email == "" {
		return scim.Resource{}, scimerrors.ScimErrorBadRequest("a valid email is required (via emails array or userName)")
	}

	emailAddr, err := mail.ParseAddr(attrs.Email)
	if err != nil {
		return scim.Resource{}, scimerrors.ScimErrorBadRequest("invalid email format")
	}
	now := time.Now()

	profileState := coredata.ProfileStateActive
	if !attrs.Active {
		profileState = coredata.ProfileStateInactive
	}

	var externalIdPtr *string
	if attrs.ExternalID != "" {
		externalIdPtr = &attrs.ExternalID
	}

	var membership *coredata.Membership
	var profile *coredata.MembershipProfile

	scope := coredata.NewScopeFromObjectID(config.OrganizationID)

	err = s.pg.WithTx(ctx, func(tx pg.Conn) error {
		identity := &coredata.Identity{}
		if err := identity.LoadByEmail(ctx, tx, emailAddr); err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				identity = &coredata.Identity{
					ID:                   gid.New(gid.NilTenant, coredata.IdentityEntityType),
					EmailAddress:         emailAddr,
					FullName:             attrs.FullName,
					HashedPassword:       nil,
					EmailAddressVerified: false,
					CreatedAt:            now,
					UpdatedAt:            now,
				}

				if err := identity.Insert(ctx, tx); err != nil {
					return fmt.Errorf("cannot insert identity: %w", err)
				}
			} else {
				return fmt.Errorf("cannot load identity: %w", err)
			}
		}

		eventType := coredata.WebhookEventTypeUserUpdated
		profile = &coredata.MembershipProfile{}
		if err := profile.LoadByIdentityIDAndOrganizationID(
			ctx,
			tx,
			coredata.NewScopeFromObjectID(config.OrganizationID),
			identity.ID,
			config.OrganizationID,
		); err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				profile = &coredata.MembershipProfile{
					ID:                     gid.New(config.OrganizationID.TenantID(), coredata.MembershipProfileEntityType),
					IdentityID:             identity.ID,
					OrganizationID:         config.OrganizationID,
					EmailAddress:           emailAddr,
					Source:                 coredata.ProfileSourceSCIM,
					State:                  profileState,
					FullName:               attrs.FullName,
					Position:               &attrs.Title,
					UserName:               &attrs.UserName,
					ExternalID:             externalIdPtr,
					Nickname:               ref.RefOrNil(attrs.Nickname),
					Locale:                 ref.RefOrNil(attrs.Locale),
					Timezone:               ref.RefOrNil(attrs.Timezone),
					ProfileUrl:             ref.RefOrNil(attrs.ProfileUrl),
					PreferredLanguage:      ref.RefOrNil(attrs.PreferredLanguage),
					GivenName:              ref.RefOrNil(attrs.GivenName),
					FamilyName:             ref.RefOrNil(attrs.FamilyName),
					FormattedName:          ref.RefOrNil(attrs.FormattedName),
					MiddleName:             ref.RefOrNil(attrs.MiddleName),
					HonorificPrefix:        ref.RefOrNil(attrs.HonorificPrefix),
					HonorificSuffix:        ref.RefOrNil(attrs.HonorificSuffix),
					EmployeeNumber:         ref.RefOrNil(attrs.EmployeeNumber),
					Department:             ref.RefOrNil(attrs.Department),
					CostCenter:             ref.RefOrNil(attrs.CostCenter),
					EnterpriseOrganization: ref.RefOrNil(attrs.EnterpriseOrganization),
					Division:               ref.RefOrNil(attrs.Division),
					ManagerValue:           ref.RefOrNil(attrs.ManagerValue),
					CreatedAt:              now,
					UpdatedAt:              now,
				}
			if attrs.UserType != "" {
				kind := attrs.UserType
				profile.Kind = &kind
			}

			err = profile.Insert(ctx, tx)
				if err != nil {
					if errors.Is(err, coredata.ErrResourceAlreadyExists) {
						return scimerrors.ScimErrorUniqueness
					}
					return fmt.Errorf("cannot insert profile: %w", err)
				}
				eventType = coredata.WebhookEventTypeUserCreated
			} else {
				return fmt.Errorf("cannot load profile: %w", err)
			}
		} else {
			if profile.Source == coredata.ProfileSourceSCIM {
				return scimerrors.ScimErrorUniqueness
			}

			profile.Source = coredata.ProfileSourceSCIM
			profile.State = profileState
			profile.FullName = attrs.FullName
			profile.Position = &attrs.Title
			profile.UserName = &attrs.UserName
			profile.ExternalID = externalIdPtr
			profile.Nickname = ref.RefOrNil(attrs.Nickname)
			profile.Locale = ref.RefOrNil(attrs.Locale)
			profile.Timezone = ref.RefOrNil(attrs.Timezone)
			profile.ProfileUrl = ref.RefOrNil(attrs.ProfileUrl)
			profile.PreferredLanguage = ref.RefOrNil(attrs.PreferredLanguage)
			profile.GivenName = ref.RefOrNil(attrs.GivenName)
			profile.FamilyName = ref.RefOrNil(attrs.FamilyName)
			profile.FormattedName = ref.RefOrNil(attrs.FormattedName)
			profile.MiddleName = ref.RefOrNil(attrs.MiddleName)
			profile.HonorificPrefix = ref.RefOrNil(attrs.HonorificPrefix)
			profile.HonorificSuffix = ref.RefOrNil(attrs.HonorificSuffix)
			profile.EmployeeNumber = ref.RefOrNil(attrs.EmployeeNumber)
			profile.Department = ref.RefOrNil(attrs.Department)
			profile.CostCenter = ref.RefOrNil(attrs.CostCenter)
			profile.EnterpriseOrganization = ref.RefOrNil(attrs.EnterpriseOrganization)
			profile.Division = ref.RefOrNil(attrs.Division)
			profile.ManagerValue = ref.RefOrNil(attrs.ManagerValue)
			profile.UpdatedAt = now
		if attrs.UserType != "" {
			kind := attrs.UserType
			profile.Kind = &kind
		}
		if err := profile.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update profile: %w", err)
			}
		}

		if !attrs.Active {
			invitations := &coredata.Invitations{}
			onlyPending := coredata.NewInvitationFilter([]coredata.InvitationStatus{coredata.InvitationStatusPending})
			if err := invitations.ExpireByUserID(
				ctx,
				tx,
				coredata.NewScopeFromObjectID(profile.OrganizationID),
				profile.ID,
				onlyPending,
			); err != nil {
				return fmt.Errorf("cannot expire pending invitations: %w", err)
			}
		}

		membership = &coredata.Membership{}
		if err := membership.LoadByIdentityIDAndOrganizationID(
			ctx,
			tx,
			scope,
			identity.ID,
			config.OrganizationID,
		); err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				membership = &coredata.Membership{
					ID:             gid.New(config.OrganizationID.TenantID(), coredata.MembershipEntityType),
					IdentityID:     identity.ID,
					OrganizationID: config.OrganizationID,
					Role:           coredata.MembershipRoleEmployee,
					CreatedAt:      now,
					UpdatedAt:      now,
				}

				err = membership.Insert(ctx, tx, scope)
				if err != nil {
					return fmt.Errorf("cannot insert membership: %w", err)
				}
			} else {
				return fmt.Errorf("cannot load membership: %w", err)
			}
		}

		if err := webhook.InsertData(ctx, tx, scope, config.OrganizationID, eventType, webhooktypes.NewUser(profile)); err != nil {
			return fmt.Errorf("cannot insert webhook event: %w", err)
		}

		return nil
	})

	if err != nil {
		return scim.Resource{}, err
	}

	return userToResource(profile), nil
}

func (s *Service) GetUser(
	ctx context.Context,
	config *coredata.SCIMConfiguration,
	profileID gid.GID,
) (scim.Resource, error) {
	scope := coredata.NewScopeFromObjectID(config.OrganizationID)

	var profile *coredata.MembershipProfile

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			profile = &coredata.MembershipProfile{}
			if err := profile.LoadByID(ctx, conn, scope, profileID); err != nil {
				if err == coredata.ErrResourceNotFound {
					return scimerrors.ScimErrorResourceNotFound(profileID.String())
				}
				return fmt.Errorf("cannot load profile: %w", err)
			}

			if profile.OrganizationID != config.OrganizationID {
				return scimerrors.ScimErrorResourceNotFound(profileID.String())
			}

			return nil
		},
	)

	if err != nil {
		return scim.Resource{}, err
	}

	return userToResource(profile), nil
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
	filter.WithSource(coredata.ProfileSourceSCIM)

	scope := coredata.NewScopeFromObjectID(config.OrganizationID)

	var profiles coredata.MembershipProfiles
	var totalCount int

	err = s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			var err error
			totalCount, err = profiles.CountByOrganizationID(ctx, conn, scope, config.OrganizationID, filter)
			if err != nil {
				return fmt.Errorf("cannot count profiles: %w", err)
			}

			orderBy := page.OrderBy[coredata.MembershipProfileOrderField]{
				Field:     coredata.MembershipProfileOrderFieldCreatedAt,
				Direction: page.OrderDirectionDesc,
			}
			cursor := page.NewCursor(count, nil, page.Head, orderBy)

			err = profiles.LoadByOrganizationID(ctx, conn, scope, config.OrganizationID, cursor, filter)
			if err != nil {
				return fmt.Errorf("cannot load profiles: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, 0, err
	}

	resources := make([]scim.Resource, 0, len(profiles))
	for _, p := range profiles {
		resources = append(resources, userToResource(p))
	}

	return resources, totalCount, nil
}

func (s *Service) ReplaceUser(
	ctx context.Context,
	config *coredata.SCIMConfiguration,
	profileID gid.GID,
	attributes scim.ResourceAttributes,
) (scim.Resource, error) {
	attrs := ParseUserFromReplaceAttributes(attributes)
	profile, err := s.updateUser(ctx, config, profileID, attrs)
	if err != nil {
		return scim.Resource{}, err
	}

	return userToResource(profile), nil
}

func (s *Service) PatchUser(
	ctx context.Context,
	config *coredata.SCIMConfiguration,
	profileID gid.GID,
	operations []scim.PatchOperation,
) (scim.Resource, error) {
	attrs := ParseUserFromPatchOperations(operations)

	profile, err := s.updateUser(ctx, config, profileID, attrs)
	if err != nil {
		return scim.Resource{}, err
	}

	return userToResource(profile), nil
}

func (s *Service) updateUser(
	ctx context.Context,
	config *coredata.SCIMConfiguration,
	profileID gid.GID,
	attrs scimReplaceAttributes,
) (*coredata.MembershipProfile, error) {
	scope := coredata.NewScopeFromObjectID(config.OrganizationID)
	now := time.Now()

	var (
		membership *coredata.Membership
		profile    *coredata.MembershipProfile
	)

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			profile = &coredata.MembershipProfile{}
			if err := profile.LoadByID(ctx, tx, scope, profileID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return scimerrors.ScimErrorResourceNotFound(profileID.String())
				}

				return fmt.Errorf("cannot load profile: %w", err)
			}

			if profile.OrganizationID != config.OrganizationID {
				return scimerrors.ScimErrorResourceNotFound(profileID.String())
			}

			membership = &coredata.Membership{}
			if err := membership.LoadByIdentityIDAndOrganizationID(ctx, tx, scope, profile.IdentityID, profile.OrganizationID); err != nil {
				return fmt.Errorf("cannot load membership: %w", err)
			}

			shouldReactivate := attrs.Active != nil && *attrs.Active && profile.State == coredata.ProfileStateInactive
			shouldDeactivate := attrs.Active != nil && !*attrs.Active && profile.State == coredata.ProfileStateActive

			if attrs.FullName != "" {
				profile.FullName = attrs.FullName
				profile.UpdatedAt = now
			}

		if attrs.Title != nil {
			if *attrs.Title == "" {
				profile.Position = nil
			} else {
				profile.Position = attrs.Title
			}
		}

			if attrs.UserName != nil {
				profile.UserName = attrs.UserName
				profile.UpdatedAt = now
			}

		if attrs.ExternalID != nil {
			if *attrs.ExternalID == "" {
				profile.ExternalID = nil
			} else {
				profile.ExternalID = attrs.ExternalID
			}
			profile.UpdatedAt = now
		}

		if attrs.UserType != nil {
		if *attrs.UserType == "" {
			profile.Kind = nil
		} else {
			profile.Kind = attrs.UserType
		}
		profile.UpdatedAt = now
	}

		if attrs.Nickname != nil {
			if *attrs.Nickname == "" {
				profile.Nickname = nil
			} else {
				profile.Nickname = attrs.Nickname
			}
			profile.UpdatedAt = now
		}

		if attrs.Locale != nil {
			if *attrs.Locale == "" {
				profile.Locale = nil
			} else {
				profile.Locale = attrs.Locale
			}
			profile.UpdatedAt = now
		}

		if attrs.Timezone != nil {
			if *attrs.Timezone == "" {
				profile.Timezone = nil
			} else {
				profile.Timezone = attrs.Timezone
			}
			profile.UpdatedAt = now
		}

	if attrs.ProfileUrl != nil {
		if *attrs.ProfileUrl == "" {
			profile.ProfileUrl = nil
		} else {
			profile.ProfileUrl = attrs.ProfileUrl
		}
		profile.UpdatedAt = now
	}

	if attrs.PreferredLanguage != nil {
		if *attrs.PreferredLanguage == "" {
			profile.PreferredLanguage = nil
		} else {
			profile.PreferredLanguage = attrs.PreferredLanguage
		}
		profile.UpdatedAt = now
	}

	if attrs.GivenName != nil {
		if *attrs.GivenName == "" {
			profile.GivenName = nil
		} else {
			profile.GivenName = attrs.GivenName
		}
		profile.UpdatedAt = now
	}

	if attrs.FamilyName != nil {
		if *attrs.FamilyName == "" {
			profile.FamilyName = nil
		} else {
			profile.FamilyName = attrs.FamilyName
		}
		profile.UpdatedAt = now
	}

	if attrs.FormattedName != nil {
		if *attrs.FormattedName == "" {
			profile.FormattedName = nil
		} else {
			profile.FormattedName = attrs.FormattedName
		}
		profile.UpdatedAt = now
	}

	if attrs.MiddleName != nil {
		if *attrs.MiddleName == "" {
			profile.MiddleName = nil
		} else {
			profile.MiddleName = attrs.MiddleName
		}
		profile.UpdatedAt = now
	}

	if attrs.HonorificPrefix != nil {
		if *attrs.HonorificPrefix == "" {
			profile.HonorificPrefix = nil
		} else {
			profile.HonorificPrefix = attrs.HonorificPrefix
		}
		profile.UpdatedAt = now
	}

	if attrs.HonorificSuffix != nil {
		if *attrs.HonorificSuffix == "" {
			profile.HonorificSuffix = nil
		} else {
			profile.HonorificSuffix = attrs.HonorificSuffix
		}
		profile.UpdatedAt = now
	}

	if attrs.EmployeeNumber != nil {
		if *attrs.EmployeeNumber == "" {
			profile.EmployeeNumber = nil
		} else {
			profile.EmployeeNumber = attrs.EmployeeNumber
		}
		profile.UpdatedAt = now
	}

	if attrs.Department != nil {
		if *attrs.Department == "" {
			profile.Department = nil
		} else {
			profile.Department = attrs.Department
		}
		profile.UpdatedAt = now
	}

	if attrs.CostCenter != nil {
		if *attrs.CostCenter == "" {
			profile.CostCenter = nil
		} else {
			profile.CostCenter = attrs.CostCenter
		}
		profile.UpdatedAt = now
	}

	if attrs.EnterpriseOrganization != nil {
		if *attrs.EnterpriseOrganization == "" {
			profile.EnterpriseOrganization = nil
		} else {
			profile.EnterpriseOrganization = attrs.EnterpriseOrganization
		}
		profile.UpdatedAt = now
	}

	if attrs.Division != nil {
		if *attrs.Division == "" {
			profile.Division = nil
		} else {
			profile.Division = attrs.Division
		}
		profile.UpdatedAt = now
	}

	if attrs.ManagerValue != nil {
		if *attrs.ManagerValue == "" {
			profile.ManagerValue = nil
		} else {
			profile.ManagerValue = attrs.ManagerValue
		}
		profile.UpdatedAt = now
	}

			if shouldReactivate {
				profile.State = coredata.ProfileStateActive
				profile.UpdatedAt = now
			} else if shouldDeactivate {
				profile.State = coredata.ProfileStateInactive
				profile.UpdatedAt = now
			}

			if profile.Source != coredata.ProfileSourceSCIM {
				profile.Source = coredata.ProfileSourceSCIM
				profile.UpdatedAt = now
			}

			if err := profile.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update membership profile: %w", err)
			}

			if shouldDeactivate {
				invitations := &coredata.Invitations{}
				onlyPending := coredata.NewInvitationFilter([]coredata.InvitationStatus{coredata.InvitationStatusPending})
				if err := invitations.ExpireByUserID(
					ctx,
					tx,
					coredata.NewScopeFromObjectID(profile.OrganizationID),
					profile.ID,
					onlyPending,
				); err != nil {
					return fmt.Errorf("cannot expire pending invitations: %w", err)
				}
			}

			needsUpdate := shouldReactivate || shouldDeactivate

			if attrs.Active != nil {
				identity := &coredata.Identity{}
				if err := identity.LoadByID(ctx, tx, membership.IdentityID); err != nil {
					return fmt.Errorf("cannot load identity: %w", err)
				}

				if shouldReactivate {
					membership.Role = coredata.MembershipRoleEmployee
				}
			}

			if needsUpdate {
				membership.UpdatedAt = now
				if err := membership.Update(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot update membership: %w", err)
				}
			}

			if err := webhook.InsertData(ctx, tx, scope, config.OrganizationID, coredata.WebhookEventTypeUserUpdated, webhooktypes.NewUser(profile)); err != nil {
				return fmt.Errorf("cannot insert webhook event: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return profile, nil
}

func (s *Service) DeleteUser(
	ctx context.Context,
	config *coredata.SCIMConfiguration,
	profileID gid.GID,
) error {
	scope := coredata.NewScopeFromObjectID(config.OrganizationID)

	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			profile := &coredata.MembershipProfile{}
			if err := profile.LoadByID(ctx, tx, scope, profileID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return scimerrors.ScimErrorResourceNotFound(profileID.String())
				}
				return fmt.Errorf("cannot load profile: %w", err)
			}

			if profile.OrganizationID != config.OrganizationID {
				return scimerrors.ScimErrorResourceNotFound(profileID.String())
			}

			invitations := &coredata.Invitations{}
			onlyPending := coredata.NewInvitationFilter([]coredata.InvitationStatus{coredata.InvitationStatusPending})
			if err := invitations.ExpireByUserID(
				ctx,
				tx,
				scope,
				profile.ID,
				onlyPending,
			); err != nil {
				return fmt.Errorf("cannot expire pending invitations: %w", err)
			}

			membership := &coredata.Membership{}
			if err := membership.LoadByIdentityIDAndOrganizationID(
				ctx, tx, scope, profile.IdentityID, config.OrganizationID,
			); err != nil {
				if !errors.Is(err, coredata.ErrResourceNotFound) {
					return fmt.Errorf("cannot load membership: %w", err)
				}
			} else {
				if err := membership.Delete(ctx, tx, scope, membership.ID); err != nil {
					return fmt.Errorf("cannot delete membership: %w", err)
				}
			}

			if err := profile.Delete(ctx, tx, scope, profile.ID); err != nil {
				return fmt.Errorf("cannot delete profile: %w", err)
			}

			if err := webhook.InsertData(ctx, tx, scope, config.OrganizationID, coredata.WebhookEventTypeUserDeleted, webhooktypes.NewUser(profile)); err != nil {
				return fmt.Errorf("cannot insert webhook event: %w", err)
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
	userName string,
	ipAddress net.IP,
	statusCode int,
	errorMessage *string,
) {
	scope := coredata.NewScopeFromObjectID(config.OrganizationID)

	event := s.createEvent(config, method, path, userName, ipAddress, statusCode, errorMessage)

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
	userName string,
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
		UserName:            userName,
		CreatedAt:           time.Now(),
	}

	return event
}

type scimUserAttributes struct {
	UserName               string
	Email                  string
	FullName               string
	Active                 bool
	Title                  string
	ExternalID             string
	UserType               string
	Nickname               string
	Locale                 string
	Timezone               string
	ProfileUrl             string
	PreferredLanguage      string
	GivenName              string
	FamilyName             string
	FormattedName          string
	MiddleName             string
	HonorificPrefix        string
	HonorificSuffix        string
	EmployeeNumber         string
	Department             string
	CostCenter             string
	EnterpriseOrganization string
	Division               string
	ManagerValue           string
}

func ParseUserFromAttributes(attributes scim.ResourceAttributes) scimUserAttributes {
	var attrs scimUserAttributes

	attrs.UserName, _ = attributes["userName"].(string)
	displayName, _ := attributes["displayName"].(string)
	attrs.ExternalID, _ = attributes["externalId"].(string)

	attrs.Active = true
	if a, ok := attributes["active"].(bool); ok {
		attrs.Active = a
	}

	var givenName, familyName string
	if name, ok := attributes["name"].(map[string]any); ok {
		givenName, _ = name["givenName"].(string)
		familyName, _ = name["familyName"].(string)
		attrs.FormattedName, _ = name["formatted"].(string)
		attrs.MiddleName, _ = name["middleName"].(string)
		attrs.HonorificPrefix, _ = name["honorificPrefix"].(string)
		attrs.HonorificSuffix, _ = name["honorificSuffix"].(string)
	}
	attrs.GivenName = givenName
	attrs.FamilyName = familyName

	if emails, ok := attributes["emails"].([]any); ok && len(emails) > 0 {
		for _, e := range emails {
			if emailMap, ok := e.(map[string]any); ok {
				if primary, _ := emailMap["primary"].(bool); primary {
					if value, ok := emailMap["value"].(string); ok {
						attrs.Email = value
						break
					}
				}
			}
		}
		if attrs.Email == "" {
			if emailMap, ok := emails[0].(map[string]any); ok {
				if value, ok := emailMap["value"].(string); ok {
					attrs.Email = value
				}
			}
		}
	}

	if attrs.Email == "" {
		if _, err := mail.ParseAddr(attrs.UserName); err == nil {
			attrs.Email = attrs.UserName
		}
	}

	attrs.FullName = displayName
	if attrs.FullName == "" {
		attrs.FullName = strings.TrimSpace(givenName + " " + familyName)
	}
	if attrs.FullName == "" {
		attrs.FullName = attrs.UserName
	}

	attrs.Title, _ = attributes["title"].(string)
	attrs.UserType, _ = attributes["userType"].(string)
	attrs.Nickname, _ = attributes["nickName"].(string)
	attrs.Locale, _ = attributes["locale"].(string)
	attrs.Timezone, _ = attributes["timezone"].(string)
	attrs.ProfileUrl, _ = attributes["profileUrl"].(string)
	attrs.PreferredLanguage, _ = attributes["preferredLanguage"].(string)

	if enterprise, ok := attributes["urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"].(map[string]any); ok {
		attrs.EmployeeNumber, _ = enterprise["employeeNumber"].(string)
		attrs.Department, _ = enterprise["department"].(string)
		attrs.CostCenter, _ = enterprise["costCenter"].(string)
		attrs.EnterpriseOrganization, _ = enterprise["organization"].(string)
		attrs.Division, _ = enterprise["division"].(string)
		if manager, ok := enterprise["manager"].(map[string]any); ok {
			attrs.ManagerValue, _ = manager["value"].(string)
		}
	}

	return attrs
}

type scimReplaceAttributes struct {
	FullName               string
	Active                 *bool
	Title                  *string
	UserName               *string
	ExternalID             *string
	UserType               *string
	Nickname               *string
	Locale                 *string
	Timezone               *string
	ProfileUrl             *string
	PreferredLanguage      *string
	GivenName              *string
	FamilyName             *string
	FormattedName          *string
	MiddleName             *string
	HonorificPrefix        *string
	HonorificSuffix        *string
	EmployeeNumber         *string
	Department             *string
	CostCenter             *string
	EnterpriseOrganization *string
	Division               *string
	ManagerValue           *string
}

func ParseUserFromReplaceAttributes(attributes scim.ResourceAttributes) scimReplaceAttributes {
	var attrs scimReplaceAttributes
	displayName, _ := attributes["displayName"].(string)

	var givenName, familyName string
	if name, ok := attributes["name"].(map[string]any); ok {
		givenName, _ = name["givenName"].(string)
		familyName, _ = name["familyName"].(string)
		if fn, ok := name["formatted"].(string); ok {
			attrs.FormattedName = &fn
		}
		if mn, ok := name["middleName"].(string); ok {
			attrs.MiddleName = &mn
		}
		if hp, ok := name["honorificPrefix"].(string); ok {
			attrs.HonorificPrefix = &hp
		}
		if hs, ok := name["honorificSuffix"].(string); ok {
			attrs.HonorificSuffix = &hs
		}
	}
	attrs.GivenName = &givenName
	attrs.FamilyName = &familyName

	attrs.FullName = displayName
	if attrs.FullName == "" {
		attrs.FullName = strings.TrimSpace(givenName + " " + familyName)
	}

	activeVal := true
	if a, ok := attributes["active"].(bool); ok {
		activeVal = a
	}
	attrs.Active = &activeVal

	t, _ := attributes["title"].(string)
	attrs.Title = &t

	if un, ok := attributes["userName"].(string); ok && un != "" {
		attrs.UserName = &un
	}

	if eid, ok := attributes["externalId"].(string); ok && eid != "" {
		attrs.ExternalID = &eid
	}

	if ut, ok := attributes["userType"].(string); ok {
		attrs.UserType = &ut
	}
	if nn, ok := attributes["nickName"].(string); ok {
		attrs.Nickname = &nn
	}
	if l, ok := attributes["locale"].(string); ok {
		attrs.Locale = &l
	}
	if tz, ok := attributes["timezone"].(string); ok {
		attrs.Timezone = &tz
	}
	if pu, ok := attributes["profileUrl"].(string); ok {
		attrs.ProfileUrl = &pu
	}
	if pl, ok := attributes["preferredLanguage"].(string); ok {
		attrs.PreferredLanguage = &pl
	}

	if enterprise, ok := attributes["urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"].(map[string]any); ok {
		if en, ok := enterprise["employeeNumber"].(string); ok {
			attrs.EmployeeNumber = &en
		}
		if dept, ok := enterprise["department"].(string); ok {
			attrs.Department = &dept
		}
		if cc, ok := enterprise["costCenter"].(string); ok {
			attrs.CostCenter = &cc
		}
		if org, ok := enterprise["organization"].(string); ok {
			attrs.EnterpriseOrganization = &org
		}
		if div, ok := enterprise["division"].(string); ok {
			attrs.Division = &div
		}
		if manager, ok := enterprise["manager"].(map[string]any); ok {
			if mv, ok := manager["value"].(string); ok {
				attrs.ManagerValue = &mv
			}
		}
	}

	return attrs
}

func ParseUserFromPatchOperations(operations []scim.PatchOperation) scimReplaceAttributes {
	var attrs scimReplaceAttributes
	var givenName, familyName string
	empty := ""

	for _, op := range operations {
		if strings.EqualFold(op.Op, "remove") {
			path := ""
			if op.Path != nil {
				path = op.Path.String()
			}

			switch strings.ToLower(path) {
			case "title":
				attrs.Title = &empty
			case "usertype":
				attrs.UserType = &empty
			case "nickname":
				attrs.Nickname = &empty
			case "locale":
				attrs.Locale = &empty
			case "timezone":
				attrs.Timezone = &empty
			case "profileurl":
				attrs.ProfileUrl = &empty
			case "externalid":
				attrs.ExternalID = &empty
			case "preferredlanguage":
				attrs.PreferredLanguage = &empty
			case "name":
				attrs.GivenName = &empty
				attrs.FamilyName = &empty
				attrs.FormattedName = &empty
				attrs.MiddleName = &empty
				attrs.HonorificPrefix = &empty
				attrs.HonorificSuffix = &empty
			case "name.givenname":
				attrs.GivenName = &empty
			case "name.familyname":
				attrs.FamilyName = &empty
			case "name.formatted":
				attrs.FormattedName = &empty
			case "name.middlename":
				attrs.MiddleName = &empty
			case "name.honorificprefix":
				attrs.HonorificPrefix = &empty
			case "name.honorificsuffix":
				attrs.HonorificSuffix = &empty
			case "urn:ietf:params:scim:schemas:extension:enterprise:2.0:user":
				attrs.EmployeeNumber = &empty
				attrs.Department = &empty
				attrs.CostCenter = &empty
				attrs.EnterpriseOrganization = &empty
				attrs.Division = &empty
				attrs.ManagerValue = &empty
			case "urn:ietf:params:scim:schemas:extension:enterprise:2.0:user:employeenumber":
				attrs.EmployeeNumber = &empty
			case "urn:ietf:params:scim:schemas:extension:enterprise:2.0:user:department":
				attrs.Department = &empty
			case "urn:ietf:params:scim:schemas:extension:enterprise:2.0:user:costcenter":
				attrs.CostCenter = &empty
			case "urn:ietf:params:scim:schemas:extension:enterprise:2.0:user:organization":
				attrs.EnterpriseOrganization = &empty
			case "urn:ietf:params:scim:schemas:extension:enterprise:2.0:user:division":
				attrs.Division = &empty
			case "urn:ietf:params:scim:schemas:extension:enterprise:2.0:user:manager",
				"urn:ietf:params:scim:schemas:extension:enterprise:2.0:user:manager.value":
				attrs.ManagerValue = &empty
			}
			continue
		}

		if strings.EqualFold(op.Op, "replace") || strings.EqualFold(op.Op, "add") {
			path := ""
			if op.Path != nil {
				path = op.Path.String()
			}

			if path == "" {
				if valueMap, ok := op.Value.(map[string]any); ok {
					if a, ok := valueMap["active"].(bool); ok {
						attrs.Active = &a
					}
					if name, ok := valueMap["displayName"].(string); ok {
						attrs.FullName = name
					}
					if nameMap, ok := valueMap["name"].(map[string]any); ok {
						if gn, ok := nameMap["givenName"].(string); ok {
							givenName = gn
						}
						if fn, ok := nameMap["familyName"].(string); ok {
							familyName = fn
						}
						if fm, ok := nameMap["formatted"].(string); ok {
							attrs.FormattedName = &fm
						}
						if mn, ok := nameMap["middleName"].(string); ok {
							attrs.MiddleName = &mn
						}
						if hp, ok := nameMap["honorificPrefix"].(string); ok {
							attrs.HonorificPrefix = &hp
						}
						if hs, ok := nameMap["honorificSuffix"].(string); ok {
							attrs.HonorificSuffix = &hs
						}
					}
					if un, ok := valueMap["userName"].(string); ok && un != "" {
						attrs.UserName = &un
					}
					if eid, ok := valueMap["externalId"].(string); ok && eid != "" {
						attrs.ExternalID = &eid
					}
					if t, ok := valueMap["title"].(string); ok {
						attrs.Title = &t
					}
					if ut, ok := valueMap["userType"].(string); ok {
						attrs.UserType = &ut
					}
					if nn, ok := valueMap["nickName"].(string); ok {
						attrs.Nickname = &nn
					}
					if l, ok := valueMap["locale"].(string); ok {
						attrs.Locale = &l
					}
					if tz, ok := valueMap["timezone"].(string); ok {
						attrs.Timezone = &tz
					}
					if pu, ok := valueMap["profileUrl"].(string); ok {
						attrs.ProfileUrl = &pu
					}
					if pl, ok := valueMap["preferredLanguage"].(string); ok {
						attrs.PreferredLanguage = &pl
					}
					if enterprise, ok := valueMap["urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"].(map[string]any); ok {
						if en, ok := enterprise["employeeNumber"].(string); ok {
							attrs.EmployeeNumber = &en
						}
						if dept, ok := enterprise["department"].(string); ok {
							attrs.Department = &dept
						}
						if cc, ok := enterprise["costCenter"].(string); ok {
							attrs.CostCenter = &cc
						}
						if org, ok := enterprise["organization"].(string); ok {
							attrs.EnterpriseOrganization = &org
						}
						if div, ok := enterprise["division"].(string); ok {
							attrs.Division = &div
						}
						if manager, ok := enterprise["manager"].(map[string]any); ok {
							if mv, ok := manager["value"].(string); ok {
								attrs.ManagerValue = &mv
							}
						}
					}

					for key, val := range valueMap {
						switch strings.ToLower(key) {
						case "name.givenname":
							if s, ok := val.(string); ok {
								givenName = s
								attrs.GivenName = &givenName
							}
						case "name.familyname":
							if s, ok := val.(string); ok {
								familyName = s
								attrs.FamilyName = &familyName
							}
						case "name.formatted":
							if s, ok := val.(string); ok {
								attrs.FormattedName = &s
							}
						case "name.middlename":
							if s, ok := val.(string); ok {
								attrs.MiddleName = &s
							}
						case "name.honorificprefix":
							if s, ok := val.(string); ok {
								attrs.HonorificPrefix = &s
							}
						case "name.honorificsuffix":
							if s, ok := val.(string); ok {
								attrs.HonorificSuffix = &s
							}
						case "urn:ietf:params:scim:schemas:extension:enterprise:2.0:user:employeenumber":
							if s, ok := val.(string); ok {
								attrs.EmployeeNumber = &s
							}
						case "urn:ietf:params:scim:schemas:extension:enterprise:2.0:user:department":
							if s, ok := val.(string); ok {
								attrs.Department = &s
							}
						case "urn:ietf:params:scim:schemas:extension:enterprise:2.0:user:costcenter":
							if s, ok := val.(string); ok {
								attrs.CostCenter = &s
							}
						case "urn:ietf:params:scim:schemas:extension:enterprise:2.0:user:organization":
							if s, ok := val.(string); ok {
								attrs.EnterpriseOrganization = &s
							}
						case "urn:ietf:params:scim:schemas:extension:enterprise:2.0:user:division":
							if s, ok := val.(string); ok {
								attrs.Division = &s
							}
						case "urn:ietf:params:scim:schemas:extension:enterprise:2.0:user:manager.value":
							if s, ok := val.(string); ok {
								attrs.ManagerValue = &s
							}
						}
					}
				}
				continue
			}

			switch strings.ToLower(path) {
			case "active":
				if a, ok := op.Value.(bool); ok {
					attrs.Active = &a
				}
			case "displayname":
				if name, ok := op.Value.(string); ok {
					attrs.FullName = name
				}
			case "name":
				if nameMap, ok := op.Value.(map[string]any); ok {
					if gn, ok := nameMap["givenName"].(string); ok {
						givenName = gn
						attrs.GivenName = &givenName
					}
					if fn, ok := nameMap["familyName"].(string); ok {
						familyName = fn
						attrs.FamilyName = &familyName
					}
					if fm, ok := nameMap["formatted"].(string); ok {
						attrs.FormattedName = &fm
					}
					if mn, ok := nameMap["middleName"].(string); ok {
						attrs.MiddleName = &mn
					}
					if hp, ok := nameMap["honorificPrefix"].(string); ok {
						attrs.HonorificPrefix = &hp
					}
					if hs, ok := nameMap["honorificSuffix"].(string); ok {
						attrs.HonorificSuffix = &hs
					}
				}
			case "name.givenname":
				if name, ok := op.Value.(string); ok {
					givenName = name
					attrs.GivenName = &givenName
				}
			case "name.familyname":
				if name, ok := op.Value.(string); ok {
					familyName = name
					attrs.FamilyName = &familyName
				}
			case "name.formatted":
				if fm, ok := op.Value.(string); ok {
					attrs.FormattedName = &fm
				}
			case "name.middlename":
				if mn, ok := op.Value.(string); ok {
					attrs.MiddleName = &mn
				}
			case "name.honorificprefix":
				if hp, ok := op.Value.(string); ok {
					attrs.HonorificPrefix = &hp
				}
			case "name.honorificsuffix":
				if hs, ok := op.Value.(string); ok {
					attrs.HonorificSuffix = &hs
				}
			case "title":
				if t, ok := op.Value.(string); ok {
					attrs.Title = &t
				}
			case "username":
				if un, ok := op.Value.(string); ok && un != "" {
					attrs.UserName = &un
				}
			case "externalid":
				if eid, ok := op.Value.(string); ok && eid != "" {
					attrs.ExternalID = &eid
				}
			case "usertype":
				if ut, ok := op.Value.(string); ok {
					attrs.UserType = &ut
				}
			case "nickname":
				if nn, ok := op.Value.(string); ok {
					attrs.Nickname = &nn
				}
			case "locale":
				if l, ok := op.Value.(string); ok {
					attrs.Locale = &l
				}
			case "timezone":
				if tz, ok := op.Value.(string); ok {
					attrs.Timezone = &tz
				}
			case "profileurl":
				if pu, ok := op.Value.(string); ok {
					attrs.ProfileUrl = &pu
				}
			case "preferredlanguage":
				if pl, ok := op.Value.(string); ok {
					attrs.PreferredLanguage = &pl
				}
			case "urn:ietf:params:scim:schemas:extension:enterprise:2.0:user":
				if enterprise, ok := op.Value.(map[string]any); ok {
					if en, ok := enterprise["employeeNumber"].(string); ok {
						attrs.EmployeeNumber = &en
					}
					if dept, ok := enterprise["department"].(string); ok {
						attrs.Department = &dept
					}
					if cc, ok := enterprise["costCenter"].(string); ok {
						attrs.CostCenter = &cc
					}
					if org, ok := enterprise["organization"].(string); ok {
						attrs.EnterpriseOrganization = &org
					}
					if div, ok := enterprise["division"].(string); ok {
						attrs.Division = &div
					}
					if manager, ok := enterprise["manager"].(map[string]any); ok {
						if mv, ok := manager["value"].(string); ok {
							attrs.ManagerValue = &mv
						}
					}
				}
			case "urn:ietf:params:scim:schemas:extension:enterprise:2.0:user:employeenumber":
				if en, ok := op.Value.(string); ok {
					attrs.EmployeeNumber = &en
				}
			case "urn:ietf:params:scim:schemas:extension:enterprise:2.0:user:department":
				if dept, ok := op.Value.(string); ok {
					attrs.Department = &dept
				}
			case "urn:ietf:params:scim:schemas:extension:enterprise:2.0:user:costcenter":
				if cc, ok := op.Value.(string); ok {
					attrs.CostCenter = &cc
				}
			case "urn:ietf:params:scim:schemas:extension:enterprise:2.0:user:organization":
				if org, ok := op.Value.(string); ok {
					attrs.EnterpriseOrganization = &org
				}
			case "urn:ietf:params:scim:schemas:extension:enterprise:2.0:user:division":
				if div, ok := op.Value.(string); ok {
					attrs.Division = &div
				}
			case "urn:ietf:params:scim:schemas:extension:enterprise:2.0:user:manager":
				if manager, ok := op.Value.(map[string]any); ok {
					if mv, ok := manager["value"].(string); ok {
						attrs.ManagerValue = &mv
					}
				} else if mv, ok := op.Value.(string); ok {
					attrs.ManagerValue = &mv
				}
			case "urn:ietf:params:scim:schemas:extension:enterprise:2.0:user:manager.value":
				if mv, ok := op.Value.(string); ok {
					attrs.ManagerValue = &mv
				}
			}
		}
	}

	if attrs.FullName == "" && (givenName != "" || familyName != "") {
		attrs.FullName = strings.TrimSpace(givenName + " " + familyName)
	}

	if givenName != "" && attrs.GivenName == nil {
		attrs.GivenName = &givenName
	}
	if familyName != "" && attrs.FamilyName == nil {
		attrs.FamilyName = &familyName
	}

	return attrs
}

func userToResource(p *coredata.MembershipProfile) scim.Resource {
	externalID := optional.NewString(p.ID.String())
	if p.ExternalID != nil {
		externalID = optional.NewString(*p.ExternalID)
	}

	formattedName := p.FullName
	if p.FormattedName != nil {
		formattedName = *p.FormattedName
	}

	nameMap := map[string]any{
		"formatted":       formattedName,
		"givenName":       ref.UnrefOrZero(p.GivenName),
		"familyName":      ref.UnrefOrZero(p.FamilyName),
		"middleName":      ref.UnrefOrZero(p.MiddleName),
		"honorificPrefix": ref.UnrefOrZero(p.HonorificPrefix),
		"honorificSuffix": ref.UnrefOrZero(p.HonorificSuffix),
	}

	enterpriseAttrs := map[string]any{
		"employeeNumber": ref.UnrefOrZero(p.EmployeeNumber),
		"department":     ref.UnrefOrZero(p.Department),
		"costCenter":     ref.UnrefOrZero(p.CostCenter),
		"organization":   ref.UnrefOrZero(p.EnterpriseOrganization),
		"division":       ref.UnrefOrZero(p.Division),
		"manager": map[string]any{
			"value": ref.UnrefOrZero(p.ManagerValue),
		},
	}

	return scim.Resource{
		ID:         p.ID.String(),
		ExternalID: externalID,
		Attributes: scim.ResourceAttributes{
			"userName":    *p.UserName,
			"displayName": p.FullName,
			"active":      p.State == coredata.ProfileStateActive,
			"name":        nameMap,
			"emails": []map[string]any{
				{
					"value":   p.EmailAddress.String(),
					"type":    "work",
					"primary": true,
				},
			},
			"title":             p.Position,
			"userType":          string(ref.UnrefOrZero(p.Kind)),
			"nickName":          ref.UnrefOrZero(p.Nickname),
			"locale":            ref.UnrefOrZero(p.Locale),
			"timezone":          ref.UnrefOrZero(p.Timezone),
			"profileUrl":        ref.UnrefOrZero(p.ProfileUrl),
			"preferredLanguage": ref.UnrefOrZero(p.PreferredLanguage),
			"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": enterpriseAttrs,
		},
		Meta: scim.Meta{
			Created:      &p.CreatedAt,
			LastModified: &p.UpdatedAt,
		},
	}
}
