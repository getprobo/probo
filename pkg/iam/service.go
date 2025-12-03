package iam

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/crypto/passwdhash"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/iam/saml"
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
	}

	Config struct {
		DisableSignup              bool
		InvitationTokenValidity    time.Duration
		PasswordResetTokenValidity time.Duration
		SessionDuration            time.Duration
		Bucket                     string
		TokenSecret                string
		BaseURL                    string
		EncryptionKey              cipher.EncryptionKey
		Certificate                *x509.Certificate
		PrivateKey                 *rsa.PrivateKey
		Logger                     *log.Logger
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
	samlService, err := saml.NewService(svc.pg, svc.encryptionKey, svc.baseURL, svc.certificate, svc.privateKey, cfg.Logger)
	if err != nil {
		return nil, fmt.Errorf("cannot create SAML service: %w", err)
	}
	svc.SAMLService = samlService

	return svc, nil
}
