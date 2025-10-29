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

package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"go.gearno.de/kit/pg"
)

type (
	CreateSAMLConfigurationRequest struct {
		OrganizationID     gid.GID
		EmailDomain        string
		EnforcementPolicy  coredata.SAMLEnforcementPolicy
		IdPEntityID        string
		IdPSsoURL          string
		IdPCertificate     string
		IdPMetadataURL     *string
		AttributeEmail     string
		AttributeFirstname string
		AttributeLastname  string
		AttributeRole      string
		DefaultRole        string
		AutoSignupEnabled  bool
	}

	UpdateSAMLConfigurationRequest struct {
		ID                 gid.GID
		Enabled            *bool
		EnforcementPolicy  *coredata.SAMLEnforcementPolicy
		IdPEntityID        *string
		IdPSsoURL          *string
		IdPCertificate     *string
		IdPMetadataURL     *string
		AttributeEmail     *string
		AttributeFirstname *string
		AttributeLastname  *string
		AttributeRole      *string
		DefaultRole        *string
		AutoSignupEnabled  *bool
	}
)

func (s TenantAuthService) CreateSAMLConfiguration(
	ctx context.Context,
	req CreateSAMLConfigurationRequest,
) (*coredata.SAMLConfiguration, error) {
	// Validate only the IdP configuration (user-provided data)
	validationErrors := ValidateIdPConfiguration(
		req.IdPEntityID,
		req.IdPSsoURL,
		req.IdPCertificate,
	)

	if len(validationErrors) > 0 {
		var errMsgs []string
		for _, err := range validationErrors {
			errMsgs = append(errMsgs, err.Error())
		}
		return nil, fmt.Errorf("SAML configuration validation failed: %s", strings.Join(errMsgs, "; "))
	}

	var config *coredata.SAMLConfiguration

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			now := time.Now()
			tenantID := s.scope.GetTenantID()

			var org coredata.Organization
			if err := org.LoadByID(ctx, tx, s.scope, req.OrganizationID); err != nil {
				return fmt.Errorf("organization not found: %w", err)
			}

			config = &coredata.SAMLConfiguration{
				ID:                 gid.New(tenantID, coredata.SAMLConfigurationEntityType),
				OrganizationID:     org.ID,
				EmailDomain:        req.EmailDomain,
				EnforcementPolicy:  req.EnforcementPolicy,
				Enabled:            false,
				IdPEntityID:        req.IdPEntityID,
				IdPSsoURL:          req.IdPSsoURL,
				IdPCertificate:     req.IdPCertificate,
				IdPMetadataURL:     req.IdPMetadataURL,
				AttributeEmail:     req.AttributeEmail,
				AttributeFirstname: req.AttributeFirstname,
				AttributeLastname:  req.AttributeLastname,
				AttributeRole:      req.AttributeRole,
				DefaultRole:        req.DefaultRole,
				AutoSignupEnabled:  req.AutoSignupEnabled,
				CreatedAt:          now,
				UpdatedAt:          now,
			}

			if err := config.Insert(ctx, tx, s.scope); err != nil {
				return fmt.Errorf("cannot insert saml configuration: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (s TenantAuthService) UpdateSAMLConfiguration(
	ctx context.Context,
	req UpdateSAMLConfigurationRequest,
) (*coredata.SAMLConfiguration, error) {
	var config *coredata.SAMLConfiguration

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			var cfg coredata.SAMLConfiguration
			if err := cfg.LoadByID(ctx, tx, s.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load saml configuration: %w", err)
			}

			if req.Enabled != nil {
				cfg.Enabled = *req.Enabled
			}
			if req.EnforcementPolicy != nil {
				cfg.EnforcementPolicy = *req.EnforcementPolicy
			}
			if req.IdPEntityID != nil {
				cfg.IdPEntityID = *req.IdPEntityID
			}
			if req.IdPSsoURL != nil {
				cfg.IdPSsoURL = *req.IdPSsoURL
			}
			if req.IdPCertificate != nil {
				cfg.IdPCertificate = *req.IdPCertificate
			}
			if req.IdPMetadataURL != nil {
				cfg.IdPMetadataURL = req.IdPMetadataURL
			}
			if req.AttributeEmail != nil {
				cfg.AttributeEmail = *req.AttributeEmail
			}
			if req.AttributeFirstname != nil {
				cfg.AttributeFirstname = *req.AttributeFirstname
			}
			if req.AttributeLastname != nil {
				cfg.AttributeLastname = *req.AttributeLastname
			}
			if req.AttributeRole != nil {
				cfg.AttributeRole = *req.AttributeRole
			}
			if req.DefaultRole != nil {
				cfg.DefaultRole = *req.DefaultRole
			}
			if req.AutoSignupEnabled != nil {
				cfg.AutoSignupEnabled = *req.AutoSignupEnabled
			}

			cfg.UpdatedAt = time.Now()

			if err := cfg.Update(ctx, tx, s.scope); err != nil {
				return fmt.Errorf("cannot update saml configuration: %w", err)
			}

			config = &cfg
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (s TenantAuthService) DeleteSAMLConfiguration(
	ctx context.Context,
	configID gid.GID,
) error {
	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			var config coredata.SAMLConfiguration
			if err := config.LoadByID(ctx, tx, s.scope, configID); err != nil {
				return fmt.Errorf("cannot load saml configuration: %w", err)
			}

			if err := config.Delete(ctx, tx, s.scope); err != nil {
				return fmt.Errorf("cannot delete saml configuration: %w", err)
			}

			return nil
		},
	)
}

func (s TenantAuthService) EnableSAMLConfiguration(
	ctx context.Context,
	configID gid.GID,
) (*coredata.SAMLConfiguration, error) {
	enabled := true
	return s.UpdateSAMLConfiguration(ctx, UpdateSAMLConfigurationRequest{
		ID:      configID,
		Enabled: &enabled,
	})
}

func (s TenantAuthService) DisableSAMLConfiguration(
	ctx context.Context,
	configID gid.GID,
) (*coredata.SAMLConfiguration, error) {
	disabled := false
	return s.UpdateSAMLConfiguration(ctx, UpdateSAMLConfigurationRequest{
		ID:      configID,
		Enabled: &disabled,
	})
}

func (s TenantAuthService) GetSAMLConfigurationByID(
	ctx context.Context,
	configID gid.GID,
) (*coredata.SAMLConfiguration, error) {
	var config coredata.SAMLConfiguration

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return config.LoadByID(ctx, conn, s.scope, configID)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load saml configuration: %w", err)
	}

	return &config, nil
}

func (s TenantAuthService) GetSAMLConfigurationsByOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
) ([]*coredata.SAMLConfiguration, error) {
	var configs []*coredata.SAMLConfiguration

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			var err error
			configs, err = coredata.LoadSAMLConfigurationsByOrganizationID(ctx, conn, s.scope, organizationID)
			return err
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load saml configurations: %w", err)
	}

	return configs, nil
}

func (s Service) CheckSSOAvailabilityByEmail(
	ctx context.Context,
	email string,
) ([]*coredata.SAMLConfiguration, error) {
	// Extract domain from email
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid email format")
	}
	domain := parts[1]

	var configs []*coredata.SAMLConfiguration
	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			var err error
			configs, err = coredata.LoadAllEnabledSAMLConfigurationsByEmailDomain(ctx, conn, domain)
			return err
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load saml configurations: %w", err)
	}

	return configs, nil
}
