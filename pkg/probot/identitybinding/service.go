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

package identitybinding

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/hash"
	"go.probo.inc/probo/pkg/gid"
)

const (
	challengeExpiry     = 30 * time.Minute
	challengeRetention  = 24 * time.Hour
	challengeBytes      = 32
	displayNameMaxRunes = 255
	listByIdentityLimit = 20
)

var (
	ErrAlreadyBound         = errors.New("external identity already bound")
	ErrChallengeAlreadyUsed = errors.New("identity binding challenge already used")
	ErrChallengeExpired     = errors.New("identity binding challenge expired")
	ErrInvalidSubject       = errors.New("invalid external identity subject")
	ErrOrganizationRequired = errors.New("organization is required")
)

type (
	Subject struct {
		Provider           string
		ExternalTenantID   string
		ExternalUserID     string
		ExternalTenantName string
		ExternalUserName   string
	}

	Binding = coredata.ProbotIdentityBinding

	Gate interface {
		Lookup(ctx context.Context, subject Subject) (*Binding, error)
		BindURL(ctx context.Context, subject Subject, organizationID gid.GID) (string, error)
	}

	Service struct {
		pg        *pg.Client
		baseURL   *baseurl.BaseURL
		now       func() time.Time
		random    io.Reader
		confirmed BindingConfirmedHandler
	}

	BindingConfirmedHandler interface {
		BindingConfirmed(ctx context.Context, subject Subject) error
	}
)

func NewService(pgClient *pg.Client, baseURL *baseurl.BaseURL) *Service {
	return &Service{
		pg:      pgClient,
		baseURL: baseURL,
		now:     time.Now,
		random:  rand.Reader,
	}
}

func (s *Service) SetBindingConfirmedHandler(handler BindingConfirmedHandler) {
	s.confirmed = handler
}

func (s Subject) Validate() error {
	if strings.TrimSpace(s.Provider) == "" ||
		strings.TrimSpace(s.ExternalUserID) == "" {
		return ErrInvalidSubject
	}

	return nil
}

func clipDisplayName(value string) string {
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if len(runes) > displayNameMaxRunes {
		return string(runes[:displayNameMaxRunes])
	}

	return value
}

func (s *Service) Lookup(
	ctx context.Context,
	subject Subject,
) (*Binding, error) {
	if err := subject.Validate(); err != nil {
		return nil, coredata.ErrResourceNotFound
	}

	var binding Binding

	err := s.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return binding.LoadByExternalSubject(
			ctx,
			conn,
			subject.Provider,
			subject.ExternalTenantID,
			subject.ExternalUserID,
		)
	})
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, coredata.ErrResourceNotFound
		}

		return nil, fmt.Errorf("cannot lookup probot identity binding: %w", err)
	}

	return &binding, nil
}

func (s *Service) BindURL(
	ctx context.Context,
	subject Subject,
	organizationID gid.GID,
) (string, error) {
	if err := subject.Validate(); err != nil {
		return "", err
	}

	if organizationID == gid.Nil {
		return "", ErrOrganizationRequired
	}

	tokenBytes := make([]byte, challengeBytes)
	if _, err := io.ReadFull(s.random, tokenBytes); err != nil {
		return "", fmt.Errorf("cannot generate identity binding challenge: %w", err)
	}

	token := base64.RawURLEncoding.EncodeToString(tokenBytes)
	now := s.now()
	challenge := coredata.ProbotIdentityBindingChallenge{
		HashedToken:        hash.SHA256String(token),
		Provider:           subject.Provider,
		ExternalTenantID:   subject.ExternalTenantID,
		ExternalUserID:     subject.ExternalUserID,
		ExternalTenantName: clipDisplayName(subject.ExternalTenantName),
		ExternalUserName:   clipDisplayName(subject.ExternalUserName),
		ExpiresAt:          now.Add(challengeExpiry),
		CreatedAt:          now,
	}

	err := s.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if err := coredata.DeleteExpiredProbotIdentityBindingChallenges(
			ctx,
			tx,
			now.Add(-challengeRetention),
		); err != nil {
			return fmt.Errorf("cannot delete expired identity binding challenges: %w", err)
		}

		return challenge.Insert(ctx, tx)
	})
	if err != nil {
		return "", fmt.Errorf("cannot persist identity binding challenge: %w", err)
	}

	bindPath, err := url.JoinPath(
		"/organizations",
		url.PathEscape(organizationID.String()),
		"employee",
		"bind",
	)
	if err != nil {
		return "", fmt.Errorf("cannot build identity binding path: %w", err)
	}

	bindURL, err := s.baseURL.
		AppendPath(bindPath).
		WithQuery("token", token).
		String()
	if err != nil {
		return "", fmt.Errorf("cannot build identity binding URL: %w", err)
	}

	return bindURL, nil
}

func (s *Service) ListByIdentity(
	ctx context.Context,
	identityID gid.GID,
) ([]*Binding, error) {
	var bindings coredata.ProbotIdentityBindings

	err := s.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return bindings.LoadByIdentityID(
			ctx,
			conn,
			identityID,
			listByIdentityLimit,
		)
	})
	if err != nil {
		return nil, fmt.Errorf("cannot list probot identity bindings: %w", err)
	}

	return bindings, nil
}

func (s *Service) Preview(
	ctx context.Context,
	token string,
) (*Subject, error) {
	challenge := &coredata.ProbotIdentityBindingChallenge{}

	err := s.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return challenge.LoadByHashedToken(
			ctx,
			conn,
			hash.SHA256String(token),
		)
	})
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, coredata.ErrResourceNotFound
		}

		return nil, fmt.Errorf("cannot load identity binding challenge: %w", err)
	}

	if challenge.ConfirmedByID != nil {
		return nil, ErrChallengeAlreadyUsed
	}

	if s.now().After(challenge.ExpiresAt) {
		return nil, ErrChallengeExpired
	}

	return &Subject{
		Provider:           challenge.Provider,
		ExternalTenantID:   challenge.ExternalTenantID,
		ExternalUserID:     challenge.ExternalUserID,
		ExternalTenantName: challenge.ExternalTenantName,
		ExternalUserName:   challenge.ExternalUserName,
	}, nil
}

func (s *Service) Confirm(
	ctx context.Context,
	identityID gid.GID,
	token string,
) (*Binding, error) {
	var binding Binding

	err := s.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		challenge := &coredata.ProbotIdentityBindingChallenge{}
		if err := challenge.LoadByHashedTokenForUpdate(
			ctx,
			tx,
			hash.SHA256String(token),
		); err != nil {
			return fmt.Errorf("cannot load identity binding challenge: %w", err)
		}

		if challenge.ConfirmedByID != nil {
			return ErrChallengeAlreadyUsed
		}

		if s.now().After(challenge.ExpiresAt) {
			return ErrChallengeExpired
		}

		existing := &Binding{}

		err := existing.LoadByExternalSubject(
			ctx,
			tx,
			challenge.Provider,
			challenge.ExternalTenantID,
			challenge.ExternalUserID,
		)
		if err == nil {
			if existing.IdentityID != identityID {
				return ErrAlreadyBound
			}

			binding = *existing

			return s.finishConfirmedChallenge(ctx, tx, challenge, identityID)
		}

		if !errors.Is(err, coredata.ErrResourceNotFound) {
			return fmt.Errorf("cannot load probot identity binding: %w", err)
		}

		existingIdentity := &Binding{}

		err = existingIdentity.LoadByIdentityAndExternalTenant(
			ctx,
			tx,
			identityID,
			challenge.Provider,
			challenge.ExternalTenantID,
		)
		if err == nil {
			return ErrAlreadyBound
		}

		if !errors.Is(err, coredata.ErrResourceNotFound) {
			return fmt.Errorf("cannot load identity probot binding: %w", err)
		}

		now := s.now()

		binding = Binding{
			ID: gid.New(
				gid.NilTenant,
				coredata.ProbotIdentityBindingEntityType,
			),
			Provider:         challenge.Provider,
			ExternalTenantID: challenge.ExternalTenantID,
			ExternalUserID:   challenge.ExternalUserID,
			IdentityID:       identityID,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if err := binding.Insert(ctx, tx); err != nil {
			if errors.Is(err, coredata.ErrResourceAlreadyExists) {
				return ErrAlreadyBound
			}

			return fmt.Errorf("cannot insert probot identity binding: %w", err)
		}

		return s.finishConfirmedChallenge(ctx, tx, challenge, identityID)
	})
	if err != nil {
		switch {
		case errors.Is(err, coredata.ErrResourceNotFound):
			return nil, coredata.ErrResourceNotFound
		default:
			return nil, err
		}
	}

	if s.confirmed != nil {
		_ = s.confirmed.BindingConfirmed(
			ctx,
			Subject{
				Provider:         binding.Provider,
				ExternalTenantID: binding.ExternalTenantID,
				ExternalUserID:   binding.ExternalUserID,
			},
		)
	}

	return &binding, nil
}

func (s *Service) finishConfirmedChallenge(
	ctx context.Context,
	tx pg.Tx,
	challenge *coredata.ProbotIdentityBindingChallenge,
	identityID gid.GID,
) error {
	if err := challenge.MarkConfirmed(ctx, tx, identityID, s.now()); err != nil {
		return fmt.Errorf("cannot mark identity binding challenge confirmed: %w", err)
	}

	return coredata.DeleteUnconfirmedProbotIdentityBindingChallengesBySubject(
		ctx,
		tx,
		challenge.Provider,
		challenge.ExternalTenantID,
		challenge.ExternalUserID,
		challenge.HashedToken,
	)
}

func (s *Service) Delete(
	ctx context.Context,
	identityID gid.GID,
	bindingID gid.GID,
) error {
	return s.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		binding := &Binding{}
		if err := binding.LoadByID(ctx, tx, bindingID); err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return coredata.ErrResourceNotFound
			}

			return fmt.Errorf("cannot load probot identity binding: %w", err)
		}

		if binding.IdentityID != identityID {
			return coredata.ErrResourceNotFound
		}

		if err := binding.Delete(ctx, tx); err != nil {
			return fmt.Errorf("cannot delete probot identity binding: %w", err)
		}

		return nil
	})
}

func DeleteByExternalTenant(
	ctx context.Context,
	conn pg.Querier,
	provider string,
	externalTenantID string,
) error {
	if strings.TrimSpace(provider) == "" ||
		strings.TrimSpace(externalTenantID) == "" {
		return ErrInvalidSubject
	}

	if err := coredata.DeleteProbotIdentityBindingsByProviderAndExternalTenant(
		ctx,
		conn,
		provider,
		externalTenantID,
	); err != nil {
		return fmt.Errorf("cannot delete identity bindings: %w", err)
	}

	if err := coredata.DeleteProbotIdentityBindingChallengesByProviderAndExternalTenant(
		ctx,
		conn,
		provider,
		externalTenantID,
	); err != nil {
		return fmt.Errorf("cannot delete identity binding challenges: %w", err)
	}

	return nil
}
