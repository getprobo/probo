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

package iam

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/hash"
	cryptorand "go.probo.inc/probo/pkg/crypto/rand"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	ServiceAccountService struct {
		*Service
	}

	CreateServiceAccountRequest struct {
		OrganizationID gid.GID
		Name           string
		Description    *string
		Scopes         coredata.OAuth2Scopes
	}

	UpdateServiceAccountRequest struct {
		Name        *string
		Description **string
		Scopes      *coredata.OAuth2Scopes
	}

	CreateServiceAccountCredentialRequest struct {
		ServiceAccountID gid.GID
		Name             string
		Scopes           coredata.OAuth2Scopes
		ExpiresAt        time.Time
	}
)

const (
	serviceAccountNameMaxLength        = 100
	serviceAccountDescriptionMaxLength = 5000
	serviceAccountTokenByteLength      = 32
)

var (
	ErrInvalidServiceAccountInput = errors.New("invalid service account input")
	ErrServiceAccountDisabled     = errors.New("service account is disabled")
)

func NewServiceAccountService(svc *Service) *ServiceAccountService {
	return &ServiceAccountService{Service: svc}
}

func (s *ServiceAccountService) validateScopes(scopes coredata.OAuth2Scopes) error {
	if err := s.OAuth2ScopeRegistry.ValidateScopes(scopes); err != nil {
		return fmt.Errorf("invalid scopes: %w", err)
	}

	return nil
}

func validateServiceAccountName(name string) error {
	if name == "" {
		return fmt.Errorf("name is required")
	}
	if len(name) > serviceAccountNameMaxLength {
		return fmt.Errorf("name must be at most %d bytes", serviceAccountNameMaxLength)
	}

	return nil
}

func validateServiceAccountDescription(description *string) error {
	if description != nil && len(*description) > serviceAccountDescriptionMaxLength {
		return fmt.Errorf("description must be at most %d bytes", serviceAccountDescriptionMaxLength)
	}

	return nil
}

func validateServiceAccountCredentialScopes(
	accountScopes coredata.OAuth2Scopes,
	credentialScopes coredata.OAuth2Scopes,
) error {
	if !accountScopes.ContainsAll(credentialScopes.Values()) {
		return fmt.Errorf("credential scopes exceed service account scopes")
	}

	return nil
}

func (s *ServiceAccountService) Create(
	ctx context.Context,
	scope coredata.Scoper,
	req CreateServiceAccountRequest,
) (*coredata.ServiceAccount, error) {
	req.Name = strings.TrimSpace(req.Name)
	if err := validateServiceAccountName(req.Name); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidServiceAccountInput, err)
	}
	if err := validateServiceAccountDescription(req.Description); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidServiceAccountInput, err)
	}
	if err := s.validateScopes(req.Scopes); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidServiceAccountInput, err)
	}

	now := time.Now()
	account := &coredata.ServiceAccount{
		ID:             gid.New(scope.GetTenantID(), coredata.ServiceAccountEntityType),
		OrganizationID: req.OrganizationID,
		Name:           req.Name,
		Description:    req.Description,
		Scopes:         req.Scopes,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, tx, scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			if err := account.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert service account: %w", err)
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	return account, nil
}

func (s *ServiceAccountService) Get(
	ctx context.Context,
	scope coredata.Scoper,
	id gid.GID,
) (*coredata.ServiceAccount, error) {
	account := &coredata.ServiceAccount{}
	if err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := account.LoadByID(ctx, conn, scope, id); err != nil {
				return fmt.Errorf("cannot load service account: %w", err)
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	return account, nil
}

func (s *ServiceAccountService) List(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.ServiceAccountOrderField],
) (*page.Page[*coredata.ServiceAccount, coredata.ServiceAccountOrderField], error) {
	accounts := coredata.ServiceAccounts{}
	if err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := accounts.LoadByOrganizationID(ctx, conn, scope, organizationID, cursor); err != nil {
				return fmt.Errorf("cannot load service accounts: %w", err)
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	return page.NewPage(accounts, cursor), nil
}

func (s *ServiceAccountService) Count(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) (int, error) {
	var count int
	if err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			accounts := coredata.ServiceAccounts{}
			var err error
			count, err = accounts.CountByOrganizationID(ctx, conn, scope, organizationID)
			if err != nil {
				return fmt.Errorf("cannot count service accounts: %w", err)
			}

			return nil
		},
	); err != nil {
		return 0, err
	}

	return count, nil
}

func (s *ServiceAccountService) Update(
	ctx context.Context,
	scope coredata.Scoper,
	id gid.GID,
	req UpdateServiceAccountRequest,
) (*coredata.ServiceAccount, error) {
	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		req.Name = &trimmed
		if err := validateServiceAccountName(trimmed); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidServiceAccountInput, err)
		}
	}
	if req.Description != nil {
		if err := validateServiceAccountDescription(*req.Description); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidServiceAccountInput, err)
		}
	}
	if req.Scopes != nil {
		if err := s.validateScopes(*req.Scopes); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidServiceAccountInput, err)
		}
	}

	account := &coredata.ServiceAccount{}
	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := account.LoadByIDForUpdate(ctx, tx, scope, id); err != nil {
				return fmt.Errorf("cannot load service account: %w", err)
			}

			if req.Name != nil {
				account.Name = *req.Name
			}
			if req.Description != nil {
				account.Description = *req.Description
			}
			if req.Scopes != nil {
				account.Scopes = *req.Scopes
			}
			account.UpdatedAt = time.Now()

			if err := account.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update service account: %w", err)
			}

			if req.Scopes != nil {
				credentials := &coredata.ServiceAccountCredentials{}
				if err := credentials.RevokeOutsideScopesByServiceAccountID(
					ctx,
					tx,
					scope,
					account.ID,
					account.Scopes,
					account.UpdatedAt,
				); err != nil {
					return fmt.Errorf("cannot constrain service account credentials: %w", err)
				}
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (s *ServiceAccountService) Disable(
	ctx context.Context,
	scope coredata.Scoper,
	id gid.GID,
) (*coredata.ServiceAccount, error) {
	account := &coredata.ServiceAccount{}
	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := account.LoadByIDForUpdate(ctx, tx, scope, id); err != nil {
				return fmt.Errorf("cannot load service account: %w", err)
			}
			if account.DisabledAt != nil {
				return nil
			}

			now := time.Now()
			account.DisabledAt = &now
			account.UpdatedAt = now
			if err := account.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot disable service account: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (s *ServiceAccountService) Delete(ctx context.Context, scope coredata.Scoper, id gid.GID) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			account := &coredata.ServiceAccount{}
			if err := account.LoadByIDForUpdate(ctx, tx, scope, id); err != nil {
				return fmt.Errorf("cannot load service account: %w", err)
			}

			now := time.Now()
			account.DisabledAt = &now
			account.DeletedAt = &now
			account.UpdatedAt = now

			credentials := &coredata.ServiceAccountCredentials{}
			if err := credentials.RevokeByServiceAccountID(ctx, tx, scope, id, now); err != nil {
				return fmt.Errorf("cannot revoke service account credentials: %w", err)
			}
			if err := account.SoftDelete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete service account: %w", err)
			}

			return nil
		},
	)
}

func (s *ServiceAccountService) CreateCredential(
	ctx context.Context,
	scope coredata.Scoper,
	req CreateServiceAccountCredentialRequest,
) (*coredata.ServiceAccountCredential, string, error) {
	req.Name = strings.TrimSpace(req.Name)
	if err := validateServiceAccountName(req.Name); err != nil {
		return nil, "", fmt.Errorf("%w: %w", ErrInvalidServiceAccountInput, err)
	}
	if !req.ExpiresAt.After(time.Now()) {
		return nil, "", fmt.Errorf("%w: expires_at must be in the future", ErrInvalidServiceAccountInput)
	}
	if err := s.validateScopes(req.Scopes); err != nil {
		return nil, "", fmt.Errorf("%w: %w", ErrInvalidServiceAccountInput, err)
	}

	rawToken, err := cryptorand.HexString(serviceAccountTokenByteLength)
	if err != nil {
		return nil, "", fmt.Errorf("cannot generate service account credential: %w", err)
	}

	credential := &coredata.ServiceAccountCredential{}
	err = s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			account := &coredata.ServiceAccount{}
			if err := account.LoadByIDForUpdate(ctx, tx, scope, req.ServiceAccountID); err != nil {
				return fmt.Errorf("cannot load service account: %w", err)
			}
			if account.DisabledAt != nil {
				return ErrServiceAccountDisabled
			}
			if err := validateServiceAccountCredentialScopes(account.Scopes, req.Scopes); err != nil {
				return fmt.Errorf("%w: %w", ErrInvalidServiceAccountInput, err)
			}

			now := time.Now()
			credential = &coredata.ServiceAccountCredential{
				ID:               gid.New(scope.GetTenantID(), coredata.ServiceAccountCredentialEntityType),
				OrganizationID:   account.OrganizationID,
				ServiceAccountID: account.ID,
				Name:             req.Name,
				HashedToken:      hash.SHA256String(rawToken),
				Scopes:           req.Scopes,
				ExpiresAt:        req.ExpiresAt,
				CreatedAt:        now,
				UpdatedAt:        now,
			}

			if err := credential.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert service account credential: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, "", err
	}

	return credential, rawToken, nil
}

func (s *ServiceAccountService) ListCredentials(
	ctx context.Context,
	scope coredata.Scoper,
	serviceAccountID gid.GID,
	cursor *page.Cursor[coredata.ServiceAccountCredentialOrderField],
) (*page.Page[*coredata.ServiceAccountCredential, coredata.ServiceAccountCredentialOrderField], error) {
	credentials := coredata.ServiceAccountCredentials{}
	if err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			account := &coredata.ServiceAccount{}
			if err := account.LoadByID(ctx, conn, scope, serviceAccountID); err != nil {
				return fmt.Errorf("cannot load service account: %w", err)
			}

			if err := credentials.LoadByServiceAccountID(ctx, conn, scope, serviceAccountID, cursor); err != nil {
				return fmt.Errorf("cannot load service account credentials: %w", err)
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	return page.NewPage(credentials, cursor), nil
}

func (s *ServiceAccountService) GetCredential(
	ctx context.Context,
	scope coredata.Scoper,
	id gid.GID,
) (*coredata.ServiceAccountCredential, error) {
	credential := &coredata.ServiceAccountCredential{}
	if err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := credential.LoadByID(ctx, conn, scope, id); err != nil {
				return fmt.Errorf("cannot load service account credential: %w", err)
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	return credential, nil
}

func (s *ServiceAccountService) CountCredentials(
	ctx context.Context,
	scope coredata.Scoper,
	serviceAccountID gid.GID,
) (int, error) {
	var count int
	if err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			credentials := coredata.ServiceAccountCredentials{}
			var err error
			count, err = credentials.CountByServiceAccountID(ctx, conn, scope, serviceAccountID)
			if err != nil {
				return fmt.Errorf("cannot count service account credentials: %w", err)
			}

			return nil
		},
	); err != nil {
		return 0, err
	}

	return count, nil
}

func (s *ServiceAccountService) RevokeCredential(
	ctx context.Context,
	scope coredata.Scoper,
	serviceAccountID gid.GID,
	credentialID gid.GID,
) (*coredata.ServiceAccountCredential, error) {
	credential := &coredata.ServiceAccountCredential{}
	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			account := &coredata.ServiceAccount{}
			if err := account.LoadByIDForUpdate(ctx, tx, scope, serviceAccountID); err != nil {
				return fmt.Errorf("cannot load service account: %w", err)
			}
			if err := credential.LoadByID(ctx, tx, scope, credentialID); err != nil {
				return fmt.Errorf("cannot load service account credential: %w", err)
			}
			if credential.ServiceAccountID != account.ID {
				return coredata.ErrResourceNotFound
			}
			if credential.RevokedAt != nil {
				return nil
			}

			now := time.Now()
			credential.RevokedAt = &now
			credential.UpdatedAt = now
			if err := credential.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot revoke service account credential: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return credential, nil
}

func (s *ServiceAccountService) Authenticate(
	ctx context.Context,
	rawToken string,
) (*coredata.ServiceAccount, *coredata.ServiceAccountCredential, error) {
	invalidCredentials := NewInvalidCredentialsError("invalid service account credential")
	if rawToken == "" {
		return nil, nil, invalidCredentials
	}

	hashedToken := hash.SHA256String(rawToken)
	account := &coredata.ServiceAccount{}
	credential := &coredata.ServiceAccountCredential{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := credential.LoadByHashedToken(ctx, tx, hashedToken); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return invalidCredentials
				}

				return fmt.Errorf("cannot load service account credential: %w", err)
			}

			scope := coredata.NewScopeFromObjectID(credential.OrganizationID)
			if err := account.LoadByIDForUpdate(ctx, tx, scope, credential.ServiceAccountID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return invalidCredentials
				}

				return fmt.Errorf("cannot load service account: %w", err)
			}
			if err := credential.LoadByHashedTokenForUpdate(ctx, tx, hashedToken); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return invalidCredentials
				}

				return fmt.Errorf("cannot lock service account credential: %w", err)
			}

			now := time.Now()
			if credential.OrganizationID != account.OrganizationID ||
				credential.RevokedAt != nil ||
				!credential.ExpiresAt.After(now) ||
				account.DisabledAt != nil ||
				account.DeletedAt != nil ||
				!account.Scopes.ContainsAll(credential.Scopes.Values()) {
				return invalidCredentials
			}

			credential.LastUsedAt = &now
			credential.UpdatedAt = now
			if err := credential.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update service account credential use: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return account, credential, nil
}
