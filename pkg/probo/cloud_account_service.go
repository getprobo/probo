// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package probo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/cloudaccount"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
	"go.probo.inc/probo/pkg/webhook"
	webhooktypes "go.probo.inc/probo/pkg/webhook/types"
)

const (
	// CloudAccountLabelMaxLength is the inclusive upper bound on
	// the customer-supplied label.
	CloudAccountLabelMaxLength = 100

	// CloudAccountScopeIdentifierMaxLength bounds AWS account ids,
	// GCP project / org ids, and Azure subscription / MG ids.
	CloudAccountScopeIdentifierMaxLength = 256

	// cloudAccountAccessSourceFKConstraint is the foreign-key
	// constraint name created by the access-source linkage migration.
	// CloudAccountService.Delete maps PG 23503 with this constraint
	// to coredata.ErrResourceInUse so resolvers can render the
	// "this cloud account is in use" conflict response.
	cloudAccountAccessSourceFKConstraint = "access_sources_cloud_account_id_fkey"
)

type (
	// CloudAccountService is a sub-service of TenantService that
	// owns the cloud-account lifecycle (Create, RotateCredentials,
	// Delete, List, Get, GenerateInstallAssets, Verify).
	CloudAccountService struct {
		svc      *TenantService
		registry *cloudaccount.Registry
		awsCfg   cloudaccount.AWSInstallTemplateConfig
	}

	// CreateCloudAccountRequest carries the non-secret metadata
	// for a new cloud account. The decrypted credentials envelope
	// is built from the same request: AWS callers pass RoleARN +
	// ExternalID, GCP callers pass ServiceAccountJSON, Azure
	// callers pass TenantID/ClientID/ClientSecret.
	CreateCloudAccountRequest struct {
		OrganizationID      gid.GID
		Label               string
		Provider            coredata.CloudAccountProvider
		CredentialKind      coredata.CloudAccountCredentialKind
		ScopeKind           coredata.CloudAccountScopeKind
		ScopeIdentifier     string
		EnabledAuditModules []coredata.CloudAccountAuditModule

		// AWS inputs (required when Provider = AWS).
		AWSRoleARN    string
		AWSExternalID string

		// GCP inputs (required when Provider = GCP).
		GCPServiceAccountJSON []byte
		GCPProjectID          string
		GCPOrganizationID     string

		// Azure inputs (required when Provider = AZURE).
		AzureTenantID        string
		AzureClientID        string
		AzureClientSecret    string
		AzureSubscriptionID  string
		AzureManagementGroup string
	}

	// RotateCloudAccountCredentialsRequest re-uses the
	// (provider, kind) discriminator as Create. Provider/kind must
	// match the loaded row -- mismatches surface as
	// cloudaccount.ErrCredentialsInvalid.
	RotateCloudAccountCredentialsRequest struct {
		CloudAccountID gid.GID
		Provider       coredata.CloudAccountProvider
		CredentialKind coredata.CloudAccountCredentialKind

		AWSRoleARN    string
		AWSExternalID string

		GCPServiceAccountJSON []byte

		AzureTenantID     string
		AzureClientID     string
		AzureClientSecret string
	}

	// VerifyCloudAccountRequest just names the row to probe.
	VerifyCloudAccountRequest struct {
		CloudAccountID gid.GID
	}

	// VerifyCloudAccountResult carries the fresh status and the
	// last_probe_error a "Verify now" UI control needs to render
	// the outcome inline.
	VerifyCloudAccountResult struct {
		Status         coredata.CloudAccountStatus
		LastProbeError *string
	}

	// GenerateInstallAssetsRequest specifies the (provider, scope,
	// modules) tuple the install-assets endpoint builds for. AWS
	// also accepts a region for the Quick-Create URL.
	GenerateInstallAssetsRequest struct {
		OrganizationID  gid.GID
		Provider        coredata.CloudAccountProvider
		ScopeKind       coredata.CloudAccountScopeKind
		ScopeIdentifier string
		Modules         []coredata.CloudAccountAuditModule
		AWSRegion       string
	}

	// GeneratedInstallAssets is a discriminated union of the
	// per-provider install payloads. Exactly one of the three
	// pointer fields is non-nil.
	GeneratedInstallAssets struct {
		AWS   *cloudaccount.AWSInstallAssets
		GCP   *cloudaccount.GCPInstallAssets
		Azure *cloudaccount.AzureInstallGuide
	}
)

// Validate performs field-level shape checks. It does not open any
// DB connection or perform external I/O -- cross-entity existence
// and tenant-match checks happen in the service method body.
func (r *CreateCloudAccountRequest) Validate() error {
	v := validator.New()
	v.Check(r.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(r.Label, "label", validator.Required(), validator.SafeTextNoNewLine(CloudAccountLabelMaxLength))
	v.Check(r.Provider, "provider", validator.Required(), validator.OneOfSlice(coredata.CloudAccountProviders()))
	v.Check(r.CredentialKind, "credential_kind", validator.Required(), validator.OneOfSlice(coredata.CloudAccountCredentialKinds()))
	v.Check(r.ScopeKind, "scope_kind", validator.Required(), validator.OneOfSlice(coredata.CloudAccountScopeKinds()))
	v.Check(r.ScopeIdentifier, "scope_identifier", validator.Required(), validator.SafeTextNoNewLine(CloudAccountScopeIdentifierMaxLength))
	return v.Error()
}

// Validate on RotateCloudAccountCredentialsRequest enforces the
// (cloudAccountID, provider, kind) shape. Provider/kind must match
// the loaded row -- mismatches surface from the service body, not
// from Validate.
func (r *RotateCloudAccountCredentialsRequest) Validate() error {
	v := validator.New()
	v.Check(r.CloudAccountID, "cloud_account_id", validator.Required(), validator.GID(coredata.CloudAccountEntityType))
	v.Check(r.Provider, "provider", validator.Required(), validator.OneOfSlice(coredata.CloudAccountProviders()))
	v.Check(r.CredentialKind, "credential_kind", validator.Required(), validator.OneOfSlice(coredata.CloudAccountCredentialKinds()))
	return v.Error()
}

// Validate on VerifyCloudAccountRequest enforces the GID type.
func (r *VerifyCloudAccountRequest) Validate() error {
	v := validator.New()
	v.Check(r.CloudAccountID, "cloud_account_id", validator.Required(), validator.GID(coredata.CloudAccountEntityType))
	return v.Error()
}

// Validate on GenerateInstallAssetsRequest enforces the
// (organizationID, provider, scope_kind) shape.
func (r *GenerateInstallAssetsRequest) Validate() error {
	v := validator.New()
	v.Check(r.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(r.Provider, "provider", validator.Required(), validator.OneOfSlice(coredata.CloudAccountProviders()))
	v.Check(r.ScopeKind, "scope_kind", validator.Required(), validator.OneOfSlice(coredata.CloudAccountScopeKinds()))
	return v.Error()
}

// List returns the paginated, filtered cloud accounts in scope for
// the supplied organization id. Credentials are not decrypted --
// callers needing the cleartext envelope must call Get.
func (s *CloudAccountService) List(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.CloudAccountOrderField],
	filter *coredata.CloudAccountFilter,
) (*page.Page[*coredata.CloudAccount, coredata.CloudAccountOrderField], error) {
	var accounts coredata.CloudAccounts

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return accounts.LoadByOrganizationID(ctx, conn, s.svc.scope, organizationID, cursor, filter)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list cloud accounts: %w", err)
	}

	return page.NewPage(accounts, cursor), nil
}

// Get loads a single cloud account by id with decrypted
// credentials. This is the only entry point any package outside
// pkg/probo (notably pkg/probo/cloud_account_worker.go and the
// access-review driver factory) uses to obtain a decrypted record.
func (s *CloudAccountService) Get(
	ctx context.Context,
	cloudAccountID gid.GID,
) (*coredata.CloudAccount, error) {
	account := &coredata.CloudAccount{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return account.LoadByID(ctx, conn, s.svc.scope, cloudAccountID, s.svc.encryptionKey)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get cloud account: %w", err)
	}

	return account, nil
}

// GetMetadata loads a cloud account by id without decrypting the
// credentials envelope. Use when only metadata is required.
func (s *CloudAccountService) GetMetadata(
	ctx context.Context,
	cloudAccountID gid.GID,
) (*coredata.CloudAccount, error) {
	account := &coredata.CloudAccount{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return account.LoadMetadataByID(ctx, conn, s.svc.scope, cloudAccountID)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get cloud account metadata: %w", err)
	}

	return account, nil
}

// Create persists a new cloud account row in PENDING_VERIFICATION
// status. The credential envelope is built from the request fields
// and encrypted in-place. A cloud_account.created webhook is
// emitted in the same transaction. Verification is the resolver's
// responsibility -- callers run Verify after Create returns to
// drive the synchronous probe.
func (s *CloudAccountService) Create(
	ctx context.Context,
	req CreateCloudAccountRequest,
) (*coredata.CloudAccount, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	envelope, err := s.marshalCredentialsForCreate(req)
	if err != nil {
		return nil, fmt.Errorf("cannot build cloud account credentials: %w", err)
	}

	now := time.Now()
	id := gid.New(s.svc.scope.GetTenantID(), coredata.CloudAccountEntityType)

	account := &coredata.CloudAccount{
		ID:                       id,
		OrganizationID:           req.OrganizationID,
		Label:                    req.Label,
		Provider:                 req.Provider,
		CredentialKind:           req.CredentialKind,
		ScopeKind:                req.ScopeKind,
		ScopeIdentifier:          req.ScopeIdentifier,
		EnabledAuditModules:      req.EnabledAuditModules,
		Status:                   coredata.CloudAccountStatusPendingVerification,
		ConsecutiveProbeFailures: 0,
		DecryptedCredentials:     envelope,
		CreatedAt:                now,
		UpdatedAt:                now,
	}

	if req.Provider == coredata.CloudAccountProviderAWS && req.AWSExternalID != "" {
		extID := req.AWSExternalID
		account.ExternalID = &extID
	}

	err = s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, tx, s.svc.scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			if err := account.Insert(ctx, tx, s.svc.scope, s.svc.encryptionKey); err != nil {
				return fmt.Errorf("cannot insert cloud account: %w", err)
			}

			if err := webhook.InsertData(
				ctx,
				tx,
				s.svc.scope,
				req.OrganizationID,
				coredata.WebhookEventTypeCloudAccountCreated,
				webhooktypes.NewCloudAccount(account),
			); err != nil {
				return fmt.Errorf("cannot insert cloud_account.created webhook: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return account, nil
}

// RotateCredentials replaces the encrypted credential envelope on
// an existing row and resets status to PENDING_VERIFICATION. The
// resolver calls Verify next to drive the synchronous probe.
func (s *CloudAccountService) RotateCredentials(
	ctx context.Context,
	req RotateCloudAccountCredentialsRequest,
) (*coredata.CloudAccount, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	account := &coredata.CloudAccount{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := account.LoadByID(ctx, tx, s.svc.scope, req.CloudAccountID, s.svc.encryptionKey); err != nil {
				return fmt.Errorf("cannot load cloud account: %w", err)
			}

			if account.Provider != req.Provider {
				return fmt.Errorf("cannot rotate cloud account credentials: provider mismatch: %w", cloudaccount.ErrCredentialsInvalid)
			}
			if account.CredentialKind != req.CredentialKind {
				return fmt.Errorf("cannot rotate cloud account credentials: credential kind mismatch: %w", cloudaccount.ErrCredentialsInvalid)
			}

			envelope, err := s.marshalCredentialsForRotate(account, req)
			if err != nil {
				return fmt.Errorf("cannot build cloud account credentials: %w", err)
			}

			account.DecryptedCredentials = envelope
			account.Status = coredata.CloudAccountStatusPendingVerification
			account.ConsecutiveProbeFailures = 0
			account.FirstProbeFailureAt = nil
			account.LastProbeError = nil
			account.UpdatedAt = time.Now()

			if req.Provider == coredata.CloudAccountProviderAWS && req.AWSExternalID != "" {
				extID := req.AWSExternalID
				account.ExternalID = &extID
			}

			return account.Update(ctx, tx, s.svc.scope, s.svc.encryptionKey)
		},
	)
	if err != nil {
		return nil, err
	}

	return account, nil
}

// Delete removes the cloud account row in scope. PG 23503 with the
// access_sources_cloud_account_id_fkey constraint name is mapped to
// coredata.ErrResourceInUse so resolvers can render
// gqlutils.Conflict("cloud account is in use").
func (s *CloudAccountService) Delete(
	ctx context.Context,
	cloudAccountID gid.GID,
) error {
	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			account := &coredata.CloudAccount{ID: cloudAccountID}
			return account.Delete(ctx, tx, s.svc.scope)
		},
	)
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" && pgErr.ConstraintName == cloudAccountAccessSourceFKConstraint {
		return coredata.ErrResourceInUse
	}

	return fmt.Errorf("cannot delete cloud account: %w", err)
}

// Verify probes the cloud account out of any DB transaction, then
// writes a short status-transition tx. On ERRORED -> VERIFIED
// recovery emits a cloud_account.verified webhook in the same tx.
// Idempotent.
func (s *CloudAccountService) Verify(
	ctx context.Context,
	cloudAccountID gid.GID,
) (*VerifyCloudAccountResult, error) {
	account, err := s.Get(ctx, cloudAccountID)
	if err != nil {
		return nil, err
	}

	record := mapCloudAccountToRecord(account)
	probeable, err := s.registry.BuildProbeable(record)
	if err != nil {
		return s.commitVerifyFailure(ctx, account, err)
	}

	probeErr := probeable.Probe(ctx)

	if probeErr != nil {
		return s.commitVerifyFailure(ctx, account, probeErr)
	}

	return s.commitVerifySuccess(ctx, account)
}

// GenerateInstallAssets builds the per-provider install payload.
// AWS persists the freshly generated external_id on the
// cloud-account row when one is supplied; GCP / Azure are pure (no
// DB write).
func (s *CloudAccountService) GenerateInstallAssets(
	ctx context.Context,
	req GenerateInstallAssetsRequest,
) (*GeneratedInstallAssets, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	switch req.Provider {
	case coredata.CloudAccountProviderAWS:
		externalID, err := cloudaccount.GenerateAWSExternalID()
		if err != nil {
			return nil, fmt.Errorf("cannot generate aws external id: %w", err)
		}

		quickCreateURL, err := cloudaccount.BuildAWSCloudFormationQuickCreateURL(
			s.awsCfg,
			externalID,
			req.Modules,
			req.AWSRegion,
		)
		if err != nil {
			return nil, fmt.Errorf("cannot build aws install assets: %w", err)
		}

		assets := &cloudaccount.AWSInstallAssets{
			QuickCreateURL:  quickCreateURL,
			ExternalID:      externalID,
			PrincipalARN:    s.awsCfg.PrincipalARN,
			RequiredActions: cloudaccount.AWSActionsForModules(req.Modules),
			Modules:         req.Modules,
			TemplateSHA256:  s.awsCfg.TemplateSHA256,
		}

		return &GeneratedInstallAssets{AWS: assets}, nil

	case coredata.CloudAccountProviderGCP:
		assets, err := cloudaccount.BuildGCPInstallScript(req.Modules, req.ScopeKind, req.ScopeIdentifier)
		if err != nil {
			return nil, fmt.Errorf("cannot build gcp install assets: %w", err)
		}

		return &GeneratedInstallAssets{GCP: &assets}, nil

	case coredata.CloudAccountProviderAzure:
		guide, err := cloudaccount.BuildAzureInstallGuide(req.Modules, req.ScopeKind)
		if err != nil {
			return nil, fmt.Errorf("cannot build azure install assets: %w", err)
		}

		return &GeneratedInstallAssets{Azure: &guide}, nil

	default:
		return nil, fmt.Errorf("cannot generate install assets: unsupported provider %q", req.Provider)
	}
}

// AttachAWSExternalID persists the freshly generated external_id
// on a cloud-account row. Used by the install-assets resolver
// after it commits an AWS asset bundle so the operator's
// CloudFormation stack can be re-bound to the persisted value.
func (s *CloudAccountService) AttachAWSExternalID(
	ctx context.Context,
	cloudAccountID gid.GID,
	externalID string,
) error {
	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			account := &coredata.CloudAccount{}
			if err := account.LoadByID(ctx, tx, s.svc.scope, cloudAccountID, s.svc.encryptionKey); err != nil {
				return fmt.Errorf("cannot load cloud account: %w", err)
			}

			ext := externalID
			account.ExternalID = &ext
			account.UpdatedAt = time.Now()

			return account.Update(ctx, tx, s.svc.scope, s.svc.encryptionKey)
		},
	)
}

func (s *CloudAccountService) commitVerifySuccess(
	ctx context.Context,
	account *coredata.CloudAccount,
) (*VerifyCloudAccountResult, error) {
	wasErrored := account.Status == coredata.CloudAccountStatusErrored ||
		account.Status == coredata.CloudAccountStatusDisconnected

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			fresh := &coredata.CloudAccount{}
			if err := fresh.LoadByID(ctx, tx, s.svc.scope, account.ID, s.svc.encryptionKey); err != nil {
				return fmt.Errorf("cannot reload cloud account: %w", err)
			}

			now := time.Now()
			fresh.Status = coredata.CloudAccountStatusVerified
			fresh.ConsecutiveProbeFailures = 0
			fresh.FirstProbeFailureAt = nil
			fresh.LastProbeError = nil
			fresh.LastProbeAt = &now
			fresh.LastVerifiedAt = &now
			fresh.UpdatedAt = now

			if err := fresh.Update(ctx, tx, s.svc.scope, s.svc.encryptionKey); err != nil {
				return fmt.Errorf("cannot update cloud account: %w", err)
			}

			if wasErrored {
				if err := webhook.InsertData(
					ctx,
					tx,
					s.svc.scope,
					fresh.OrganizationID,
					coredata.WebhookEventTypeCloudAccountVerified,
					webhooktypes.NewCloudAccount(fresh),
				); err != nil {
					return fmt.Errorf("cannot insert cloud_account.verified webhook: %w", err)
				}
			}

			*account = *fresh
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot commit cloud account verification: %w", err)
	}

	return &VerifyCloudAccountResult{
		Status:         account.Status,
		LastProbeError: account.LastProbeError,
	}, nil
}

func (s *CloudAccountService) commitVerifyFailure(
	ctx context.Context,
	account *coredata.CloudAccount,
	probeErr error,
) (*VerifyCloudAccountResult, error) {
	errMsg := probeErr.Error()
	now := time.Now()

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			fresh := &coredata.CloudAccount{}
			if err := fresh.LoadByID(ctx, tx, s.svc.scope, account.ID, s.svc.encryptionKey); err != nil {
				return fmt.Errorf("cannot reload cloud account: %w", err)
			}

			fresh.LastProbeAt = &now
			fresh.LastProbeError = &errMsg
			fresh.UpdatedAt = now

			if fresh.Status == coredata.CloudAccountStatusPendingVerification {
				// Stay PENDING_VERIFICATION; never auto-promote a
				// never-verified row.
				if err := fresh.Update(ctx, tx, s.svc.scope, s.svc.encryptionKey); err != nil {
					return fmt.Errorf("cannot update cloud account: %w", err)
				}
				*account = *fresh
				return nil
			}

			fresh.ConsecutiveProbeFailures++
			if fresh.FirstProbeFailureAt == nil {
				fresh.FirstProbeFailureAt = &now
			}
			fresh.Status = coredata.CloudAccountStatusErrored

			if err := fresh.Update(ctx, tx, s.svc.scope, s.svc.encryptionKey); err != nil {
				return fmt.Errorf("cannot update cloud account: %w", err)
			}

			*account = *fresh
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot commit cloud account verification failure: %w", err)
	}

	return &VerifyCloudAccountResult{
		Status:         account.Status,
		LastProbeError: account.LastProbeError,
	}, nil
}

// marshalCredentialsForCreate builds the cleartext credential
// envelope from the create-request fields. The envelope kind is
// the source of truth; mismatches between request.CredentialKind
// and request.Provider surface here.
func (s *CloudAccountService) marshalCredentialsForCreate(req CreateCloudAccountRequest) ([]byte, error) {
	switch req.Provider {
	case coredata.CloudAccountProviderAWS:
		if req.CredentialKind != coredata.CloudAccountCredentialKindAWSAssumeRole {
			return nil, fmt.Errorf("cannot build aws credentials: unexpected kind %q: %w", req.CredentialKind, cloudaccount.ErrCredentialsInvalid)
		}

		creds := &cloudaccount.AWSCredentials{
			RoleARN:    req.AWSRoleARN,
			ExternalID: req.AWSExternalID,
			ScopeKind:  req.ScopeKind,
		}
		return creds.MarshalJSON()

	case coredata.CloudAccountProviderGCP:
		if req.CredentialKind != coredata.CloudAccountCredentialKindGCPServiceAccountKey {
			return nil, fmt.Errorf("cannot build gcp credentials: unexpected kind %q: %w", req.CredentialKind, cloudaccount.ErrCredentialsInvalid)
		}

		creds := &cloudaccount.GCPCredentials{
			ServiceAccountJSON: req.GCPServiceAccountJSON,
			ScopeKind:          req.ScopeKind,
			ProjectID:          req.GCPProjectID,
			OrganizationID:     req.GCPOrganizationID,
		}
		return creds.MarshalJSON()

	case coredata.CloudAccountProviderAzure:
		if req.CredentialKind != coredata.CloudAccountCredentialKindAzureClientSecret {
			return nil, fmt.Errorf("cannot build azure credentials: unexpected kind %q: %w", req.CredentialKind, cloudaccount.ErrCredentialsInvalid)
		}

		creds := &cloudaccount.AzureCredentials{
			TenantID:          req.AzureTenantID,
			ClientID:          req.AzureClientID,
			ClientSecret:      req.AzureClientSecret,
			ScopeKind:         req.ScopeKind,
			SubscriptionID:    req.AzureSubscriptionID,
			ManagementGroupID: req.AzureManagementGroup,
		}
		return creds.MarshalJSON()

	default:
		return nil, fmt.Errorf("cannot build cloud account credentials: unsupported provider %q", req.Provider)
	}
}

// marshalCredentialsForRotate builds the new envelope from the
// rotate-request fields, preserving immutable scope identifiers
// from the loaded row (the request only carries credentials, not
// scope edits -- those are a separate flow).
func (s *CloudAccountService) marshalCredentialsForRotate(
	account *coredata.CloudAccount,
	req RotateCloudAccountCredentialsRequest,
) ([]byte, error) {
	switch req.Provider {
	case coredata.CloudAccountProviderAWS:
		creds := &cloudaccount.AWSCredentials{
			RoleARN:    req.AWSRoleARN,
			ExternalID: req.AWSExternalID,
			ScopeKind:  account.ScopeKind,
		}
		return creds.MarshalJSON()

	case coredata.CloudAccountProviderGCP:
		creds := &cloudaccount.GCPCredentials{
			ServiceAccountJSON: req.GCPServiceAccountJSON,
			ScopeKind:          account.ScopeKind,
			ProjectID:          extractGCPProjectIDFromAccount(account),
			OrganizationID:     extractGCPOrganizationIDFromAccount(account),
		}
		return creds.MarshalJSON()

	case coredata.CloudAccountProviderAzure:
		creds := &cloudaccount.AzureCredentials{
			TenantID:          req.AzureTenantID,
			ClientID:          req.AzureClientID,
			ClientSecret:      req.AzureClientSecret,
			ScopeKind:         account.ScopeKind,
			SubscriptionID:    extractAzureSubscriptionIDFromAccount(account),
			ManagementGroupID: extractAzureManagementGroupFromAccount(account),
		}
		return creds.MarshalJSON()

	default:
		return nil, fmt.Errorf("cannot build cloud account credentials: unsupported provider %q", req.Provider)
	}
}

// mapCloudAccountToRecord maps a coredata.CloudAccount entity to the
// thin pkg/cloudaccount.CloudAccountRecord value the registry
// operates on. Keeps pkg/cloudaccount free of any coredata entity
// dependency.
func mapCloudAccountToRecord(account *coredata.CloudAccount) cloudaccount.CloudAccountRecord {
	rec := cloudaccount.CloudAccountRecord{
		ID:                   account.ID.String(),
		Provider:             account.Provider,
		Kind:                 account.CredentialKind,
		ScopeKind:            account.ScopeKind,
		ScopeIdentifier:      account.ScopeIdentifier,
		DecryptedCredentials: account.DecryptedCredentials,
	}
	if account.ExternalID != nil {
		rec.ExternalID = *account.ExternalID
	}
	return rec
}

// extractGCPProjectIDFromAccount pulls ProjectID from the loaded
// row's scope when scope_kind=GCP_PROJECT. RotateCredentials
// requires it because the GCP envelope embeds the scope identifier
// in the project_id slot for project-scoped accounts.
func extractGCPProjectIDFromAccount(account *coredata.CloudAccount) string {
	if account.ScopeKind == coredata.CloudAccountScopeKindGCPProject {
		return account.ScopeIdentifier
	}
	return ""
}

func extractGCPOrganizationIDFromAccount(account *coredata.CloudAccount) string {
	if account.ScopeKind == coredata.CloudAccountScopeKindGCPOrganization {
		return account.ScopeIdentifier
	}
	return ""
}

func extractAzureSubscriptionIDFromAccount(account *coredata.CloudAccount) string {
	if account.ScopeKind == coredata.CloudAccountScopeKindAzureSubscription {
		return account.ScopeIdentifier
	}
	return ""
}

func extractAzureManagementGroupFromAccount(account *coredata.CloudAccount) string {
	switch account.ScopeKind {
	case coredata.CloudAccountScopeKindAzureManagementGroup, coredata.CloudAccountScopeKindAzureTenant:
		return account.ScopeIdentifier
	}
	return ""
}
