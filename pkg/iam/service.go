// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

package iam

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.opentelemetry.io/otel/trace"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/crypto/passwdhash"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/oidc"
	"go.probo.inc/probo/pkg/iam/saml"
	"go.probo.inc/probo/pkg/iam/scim"
)

type (
	Service struct {
		pg                         *pg.Client
		fm                         *filemanager.Service
		hp                         *passwdhash.Profile
		baseURL                    string
		tokenSecret                string
		disableSignup              bool
		invitationTokenValidity    time.Duration
		passwordResetTokenValidity time.Duration
		magicLinkTokenValidity     time.Duration
		sessionDuration            time.Duration
		bucket                     string
		certificate                *x509.Certificate
		privateKey                 *rsa.PrivateKey
		logger                     *log.Logger

		AccountService        *AccountService
		OrganizationService   *OrganizationService
		CompliancePageService *CompliancePageService
		SessionService        *SessionService
		AuthService           *AuthService
		SAMLService           *saml.Service
		OIDCService           *oidc.Service
		SCIMService           *scim.Service
		APIKeyService         *APIKeyService
		Authorizer            *Authorizer

		samlDomainVerifier *SAMLDomainVerifier
		logExportWorker    *LogExportWorker
	}

	Config struct {
		DisableSignup                  bool
		InvitationTokenValidity        time.Duration
		PasswordResetTokenValidity     time.Duration
		MagicLinkTokenValidity         time.Duration
		SessionDuration                time.Duration
		Bucket                         string
		TokenSecret                    string
		BaseURL                        *baseurl.BaseURL
		EncryptionKey                  cipher.EncryptionKey
		Certificate                    *x509.Certificate
		PrivateKey                     *rsa.PrivateKey
		Logger                         *log.Logger
		TracerProvider                 trace.TracerProvider
		Registerer                     prometheus.Registerer
		ConnectorRegistry              *connector.ConnectorRegistry
		DomainVerificationInterval     time.Duration
		DomainVerificationResolverAddr string
		SCIMBridgeSyncInterval         time.Duration
		SCIMBridgePollInterval         time.Duration
		GoogleOIDC                     oidc.ProviderConfig
		MicrosoftOIDC                  oidc.ProviderConfig
	}
)

func NewService(
	ctx context.Context,
	pgClient *pg.Client,
	fm *filemanager.Service,
	hp *passwdhash.Profile,
	cfg Config,
) (*Service, error) {
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("bucket is required")
	}

	if cfg.TokenSecret == "" {
		return nil, fmt.Errorf("token secret is required")
	}

	if cfg.BaseURL == nil {
		return nil, fmt.Errorf("base URL is required")
	}

	if len(cfg.EncryptionKey) == 0 {
		return nil, fmt.Errorf("encryption key is required")
	}

	svc := &Service{
		pg:                         pgClient,
		fm:                         fm,
		hp:                         hp,
		baseURL:                    cfg.BaseURL.String(),
		tokenSecret:                cfg.TokenSecret,
		disableSignup:              cfg.DisableSignup,
		invitationTokenValidity:    cfg.InvitationTokenValidity,
		passwordResetTokenValidity: cfg.PasswordResetTokenValidity,
		magicLinkTokenValidity:     cfg.MagicLinkTokenValidity,
		sessionDuration:            cfg.SessionDuration,
		bucket:                     cfg.Bucket,
		certificate:                cfg.Certificate,
		privateKey:                 cfg.PrivateKey,
		logger:                     cfg.Logger,
	}

	svc.AccountService = NewAccountService(svc)
	svc.OrganizationService = NewOrganizationService(svc)
	svc.CompliancePageService = NewCompliancePageService(svc)
	svc.SessionService = NewSessionService(svc)
	svc.AuthService = NewAuthService(svc)
	svc.APIKeyService = NewAPIKeyService(svc)

	svc.Authorizer = NewAuthorizer(
		pgClient,
		cfg.Logger.Named("authorizer"),
	)
	svc.Authorizer.RegisterPolicySet(IAMPolicySet())

	samlService, err := saml.NewService(svc.pg, svc.baseURL, svc.certificate, svc.privateKey, cfg.Logger)
	if err != nil {
		return nil, fmt.Errorf("cannot create SAML service: %w", err)
	}
	svc.SAMLService = samlService

	svc.OIDCService = oidc.NewService(
		svc.pg,
		svc.baseURL,
		cfg.GoogleOIDC,
		cfg.MicrosoftOIDC,
		cfg.Logger,
	)

	svc.SCIMService = scim.NewService(
		svc.pg,
		cfg.Logger.Named("scim"),
		scim.ServiceConfig{
			TracerProvider:    cfg.TracerProvider,
			Registerer:        cfg.Registerer,
			EncryptionKey:     cfg.EncryptionKey,
			ConnectorRegistry: cfg.ConnectorRegistry,
			BridgeRunner: scim.BridgeRunnerConfig{
				Interval:     cfg.SCIMBridgeSyncInterval,
				PollInterval: cfg.SCIMBridgePollInterval,
				BaseURL:      cfg.BaseURL,
			},
		},
	)

	svc.samlDomainVerifier = NewSAMLDomainVerifier(
		pgClient,
		cfg.Logger,
		cfg.TracerProvider,
		cfg.DomainVerificationInterval,
		cfg.DomainVerificationResolverAddr,
	)

	svc.logExportWorker = NewLogExportWorker(
		pgClient,
		fm,
		cfg.Bucket,
		cfg.BaseURL.String(),
		cfg.Logger.Named("log-export-worker"),
	)

	return svc, nil
}

func (s *Service) Run(ctx context.Context) error {
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(context.Canceled)

	samlCtx, stopSAML := context.WithCancel(context.WithoutCancel(ctx))
	wg.Go(
		func() {
			if err := s.SAMLService.Run(samlCtx); err != nil {
				cancel(fmt.Errorf("saml service crashed: %w", err))
			}
		},
	)

	oidcCtx, stopOIDC := context.WithCancel(context.WithoutCancel(ctx))
	wg.Go(
		func() {
			if err := s.OIDCService.Run(oidcCtx); err != nil {
				cancel(fmt.Errorf("oidc service crashed: %w", err))
			}
		},
	)

	domainVerifierCtx, stopDomainVerifier := context.WithCancel(context.WithoutCancel(ctx))
	wg.Go(
		func() {
			if err := s.samlDomainVerifier.Run(domainVerifierCtx); err != nil {
				cancel(fmt.Errorf("saml domain verifier crashed: %w", err))
			}
		},
	)

	scimCtx, stopSCIM := context.WithCancel(context.WithoutCancel(ctx))
	wg.Go(
		func() {
			if err := s.SCIMService.Run(scimCtx); err != nil {
				cancel(fmt.Errorf("scim service crashed: %w", err))
			}
		},
	)

	logExportCtx, stopLogExport := context.WithCancel(context.WithoutCancel(ctx))
	wg.Go(
		func() {
			if err := s.logExportWorker.Run(logExportCtx); err != nil {
				cancel(fmt.Errorf("log export worker crashed: %w", err))
			}
		},
	)

	<-ctx.Done()

	stopSAML()
	stopOIDC()
	stopDomainVerifier()
	stopSCIM()
	stopLogExport()

	wg.Wait()

	return context.Cause(ctx)
}

func (s *Service) GetMembership(ctx context.Context, membershipID gid.GID) (*coredata.Membership, error) {
	var (
		scope      = coredata.NewScopeFromObjectID(membershipID)
		membership = &coredata.Membership{}
	)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := membership.LoadByID(ctx, conn, scope, membershipID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewMembershipNotFoundError(membershipID)
				}

				return fmt.Errorf("cannot load membership: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return membership, nil
}

func (s *Service) GetInvitation(ctx context.Context, invitationID gid.GID) (*coredata.Invitation, error) {
	var (
		scope      = coredata.NewScopeFromObjectID(invitationID)
		invitation = &coredata.Invitation{}
	)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := invitation.LoadByID(ctx, conn, scope, invitationID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewInvitationNotFoundError(invitationID)
				}

				return fmt.Errorf("cannot load invitation: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return invitation, nil
}

func (s *Service) GetSession(ctx context.Context, sessionID gid.GID) (*coredata.Session, error) {
	session := &coredata.Session{}

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := session.LoadByID(ctx, conn, sessionID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewSessionNotFoundError(sessionID)
				}

				return fmt.Errorf("cannot load session: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (s *Service) GetSAMLconfiguration(ctx context.Context, samlConfigurationID gid.GID) (*coredata.SAMLConfiguration, error) {
	var (
		scope             = coredata.NewScopeFromObjectID(samlConfigurationID)
		samlConfiguration = &coredata.SAMLConfiguration{}
	)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := samlConfiguration.LoadByID(ctx, conn, scope, samlConfigurationID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return saml.NewSAMLConfigurationNotFoundError(samlConfigurationID)
				}

				return fmt.Errorf("cannot load SAML configuration: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return samlConfiguration, nil
}

func (s *Service) GetPersonalAPIKey(ctx context.Context, personalAPIKeyID gid.GID) (*coredata.PersonalAPIKey, error) {
	personalAPIKey := &coredata.PersonalAPIKey{}

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := personalAPIKey.LoadByID(ctx, conn, personalAPIKeyID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewPersonalAPIKeyNotFoundError(personalAPIKeyID)
				}

				return fmt.Errorf("cannot load personal API key: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return personalAPIKey, nil
}

func (s *Service) GetSCIMConfiguration(ctx context.Context, scimConfigurationID gid.GID) (*coredata.SCIMConfiguration, error) {
	var (
		scope             = coredata.NewScopeFromObjectID(scimConfigurationID)
		scimConfiguration = &coredata.SCIMConfiguration{}
	)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := scimConfiguration.LoadByID(ctx, conn, scope, scimConfigurationID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return scim.NewSCIMConfigurationNotFoundError(scimConfigurationID)
				}

				return fmt.Errorf("cannot load SCIM configuration: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return scimConfiguration, nil
}

func (s *Service) GetSCIMEvent(ctx context.Context, scimEventID gid.GID) (*coredata.SCIMEvent, error) {
	var (
		scope     = coredata.NewScopeFromObjectID(scimEventID)
		scimEvent = &coredata.SCIMEvent{}
	)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := scimEvent.LoadByID(ctx, conn, scope, scimEventID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return fmt.Errorf("SCIM event not found: %s", scimEventID)
				}

				return fmt.Errorf("cannot load SCIM event: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return scimEvent, nil
}
