package iam

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
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
	"go.probo.inc/probo/pkg/iam/saml"
	"go.probo.inc/probo/pkg/iam/scim"
	"golang.org/x/sync/errgroup"
)

type (
	Service struct {
		pg                         *pg.Client
		fm                         *filemanager.Service
		hp                         *passwdhash.Profile
		encryptionKey              cipher.EncryptionKey
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
		SCIMService           *scim.Service
		APIKeyService         *APIKeyService
		Authorizer            *Authorizer

		samlDomainVerifier *SAMLDomainVerifier
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

	svc.Authorizer = NewAuthorizer(pgClient)
	svc.Authorizer.RegisterPolicySet(IAMPolicySet())

	samlService, err := saml.NewService(svc.pg, svc.encryptionKey, svc.baseURL, svc.certificate, svc.privateKey, cfg.Logger)
	if err != nil {
		return nil, fmt.Errorf("cannot create SAML service: %w", err)
	}
	svc.SAMLService = samlService

	svc.SCIMService = scim.NewService(svc.pg, cfg.Logger.Named("scim"), scim.ServiceConfig{
		TracerProvider:    cfg.TracerProvider,
		Registerer:        cfg.Registerer,
		EncryptionKey:     cfg.EncryptionKey,
		ConnectorRegistry: cfg.ConnectorRegistry,
		BridgeRunner: scim.BridgeRunnerConfig{
			Interval:     cfg.SCIMBridgeSyncInterval,
			PollInterval: cfg.SCIMBridgePollInterval,
			BaseURL:      cfg.BaseURL,
		},
	})

	svc.samlDomainVerifier = NewSAMLDomainVerifier(
		pgClient,
		cfg.Logger,
		cfg.TracerProvider,
		cfg.DomainVerificationInterval,
		cfg.DomainVerificationResolverAddr,
	)

	return svc, nil
}

func (s *Service) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error { return s.SAMLService.Run(ctx) })
	g.Go(func() error { return s.samlDomainVerifier.Run(ctx) })
	g.Go(func() error { return s.SCIMService.Run(ctx) })

	return g.Wait()
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
