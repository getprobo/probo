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

package slackbot

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/hash"
	"go.probo.inc/probo/pkg/gid"
)

var (
	ErrSlackIdentityAlreadyBound = errors.New("slack identity already bound")
	ErrBindTokenAlreadyUsed      = errors.New("slack bind token already used")
)

type BindingService struct {
	pg          *pg.Client
	tokenSecret string
	baseURL     *baseurl.BaseURL
}

func NewBindingService(
	pgClient *pg.Client,
	tokenSecret string,
	baseURL *baseurl.BaseURL,
) *BindingService {
	return &BindingService{
		pg:          pgClient,
		tokenSecret: tokenSecret,
		baseURL:     baseURL,
	}
}

func (s *BindingService) Lookup(
	ctx context.Context,
	teamID, slackUserID string,
) (*coredata.SlackIdentityBinding, error) {
	if teamID == "" || slackUserID == "" {
		return nil, coredata.ErrResourceNotFound
	}

	var binding coredata.SlackIdentityBinding

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := binding.LoadBySlackUser(ctx, conn, teamID, slackUserID); err != nil {
				return err
			}

			return nil
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, coredata.ErrResourceNotFound
		}

		return nil, fmt.Errorf("cannot lookup slack identity binding: %w", err)
	}

	return &binding, nil
}

func (s *BindingService) BindURL(teamID, slackUserID string) (string, error) {
	token, err := newBindToken(s.tokenSecret, teamID, slackUserID)
	if err != nil {
		return "", fmt.Errorf("cannot generate bind token: %w", err)
	}

	return s.baseURL.
		AppendPath("/me/slack/bind").
		WithQuery("token", token).
		MustString(), nil
}

func (s *BindingService) Preview(token string) (*BindTokenData, error) {
	payload, err := validateBindToken(s.tokenSecret, token)
	if err != nil {
		return nil, err
	}

	return &payload.Data, nil
}

func (s *BindingService) Confirm(
	ctx context.Context,
	identityID gid.GID,
	token string,
) (*coredata.SlackIdentityBinding, error) {
	payload, err := validateBindToken(s.tokenSecret, token)
	if err != nil {
		return nil, err
	}

	hashedToken := hash.SHA256String(token)
	var binding coredata.SlackIdentityBinding

	err = s.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		now := time.Now()
		tokenUse := coredata.SlackIdentityBindTokenUse{
			HashedToken: hashedToken,
			IdentityID:  identityID,
			UsedAt:      now,
			ExpiresAt:   payload.ExpiresAt,
		}

		// Insert inside a savepoint so a unique violation does not abort the
		// outer transaction before we can load the existing token use.
		err := tx.Savepoint(ctx, func(ctx context.Context, sp pg.Tx) error {
			return tokenUse.Insert(ctx, sp)
		})
		if errors.Is(err, coredata.ErrResourceAlreadyExists) {
			existingUse := &coredata.SlackIdentityBindTokenUse{}
			if loadErr := existingUse.LoadByHashedToken(ctx, tx, hashedToken); loadErr != nil {
				return fmt.Errorf("cannot load slack bind token use: %w", loadErr)
			}
			if existingUse.IdentityID != identityID {
				return ErrBindTokenAlreadyUsed
			}

			existing := &coredata.SlackIdentityBinding{}
			if loadErr := existing.LoadBySlackUser(
				ctx,
				tx,
				payload.Data.TeamID,
				payload.Data.SlackUserID,
			); loadErr != nil {
				return fmt.Errorf("cannot load slack identity binding: %w", loadErr)
			}
			if existing.IdentityID != identityID {
				return ErrSlackIdentityAlreadyBound
			}

			binding = *existing

			return nil
		}
		if err != nil {
			return fmt.Errorf("cannot consume slack bind token: %w", err)
		}

		existing := &coredata.SlackIdentityBinding{}
		err = existing.LoadBySlackUser(ctx, tx, payload.Data.TeamID, payload.Data.SlackUserID)
		if err == nil {
			if existing.IdentityID == identityID {
				binding = *existing

				return nil
			}

			return ErrSlackIdentityAlreadyBound
		}
		if !errors.Is(err, coredata.ErrResourceNotFound) {
			return fmt.Errorf("cannot load slack identity binding: %w", err)
		}

		existingIdentity := &coredata.SlackIdentityBinding{}
		err = existingIdentity.LoadByIdentityAndTeam(ctx, tx, identityID, payload.Data.TeamID)
		if err == nil {
			return ErrSlackIdentityAlreadyBound
		}
		if !errors.Is(err, coredata.ErrResourceNotFound) {
			return fmt.Errorf("cannot load identity slack binding: %w", err)
		}

		binding = coredata.SlackIdentityBinding{
			ID:          gid.New(gid.NilTenant, coredata.SlackIdentityBindingEntityType),
			TeamID:      payload.Data.TeamID,
			SlackUserID: payload.Data.SlackUserID,
			IdentityID:  identityID,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		if err := binding.Insert(ctx, tx); err != nil {
			if errors.Is(err, coredata.ErrResourceAlreadyExists) {
				return ErrSlackIdentityAlreadyBound
			}

			return fmt.Errorf("cannot insert slack identity binding: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &binding, nil
}

func (s *BindingService) Delete(
	ctx context.Context,
	identityID gid.GID,
	bindingID gid.GID,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			binding := &coredata.SlackIdentityBinding{}

			err := binding.LoadByID(ctx, tx, bindingID)
			if err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return coredata.ErrResourceNotFound
				}

				return fmt.Errorf("cannot load slack identity binding: %w", err)
			}

			if binding.IdentityID != identityID {
				return coredata.ErrResourceNotFound
			}

			if err := binding.Delete(ctx, tx); err != nil {
				return fmt.Errorf("cannot delete slack identity binding: %w", err)
			}

			return nil
		},
	)
}
