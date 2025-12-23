package iam

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.opentelemetry.io/otel/trace"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/crypto/passwdhash"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/saml"
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
		sessionDuration            time.Duration
		bucket                     string
		certificate                *x509.Certificate
		privateKey                 *rsa.PrivateKey
		logger                     *log.Logger

		AccountService      *AccountService
		OrganizationService *OrganizationService
		SessionService      *SessionService
		AuthService         *AuthService
		SAMLService         *saml.Service
		APIKeyService       *APIKeyService
		Authorizer          *Authorizer

		samlDomainVerifier *SAMLDomainVerifier
	}

	Config struct {
		DisableSignup                  bool
		InvitationTokenValidity        time.Duration
		PasswordResetTokenValidity     time.Duration
		SessionDuration                time.Duration
		Bucket                         string
		TokenSecret                    string
		BaseURL                        string
		EncryptionKey                  cipher.EncryptionKey
		Certificate                    *x509.Certificate
		PrivateKey                     *rsa.PrivateKey
		Logger                         *log.Logger
		TracerProvider                 trace.TracerProvider
		DomainVerificationInterval     time.Duration
		DomainVerificationResolverAddr string
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

	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}

	if len(cfg.EncryptionKey) == 0 {
		return nil, fmt.Errorf("encryption key is required")
	}

	svc := &Service{
		pg:                         pgClient,
		fm:                         fm,
		hp:                         hp,
		baseURL:                    cfg.BaseURL,
		tokenSecret:                cfg.TokenSecret,
		disableSignup:              cfg.DisableSignup,
		invitationTokenValidity:    cfg.InvitationTokenValidity,
		passwordResetTokenValidity: cfg.PasswordResetTokenValidity,
		sessionDuration:            cfg.SessionDuration,
		bucket:                     cfg.Bucket,
		certificate:                cfg.Certificate,
		privateKey:                 cfg.PrivateKey,
		logger:                     cfg.Logger,
	}

	svc.AccountService = NewAccountService(svc)
	svc.OrganizationService = NewOrganizationService(svc)
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
