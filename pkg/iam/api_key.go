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

package iam

import (
	"context"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.gearno.de/x/ref"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type (
	APIKeyService struct {
		*Service
	}
)

func NewAPIKeyService(svc *Service) *APIKeyService {
	return &APIKeyService{Service: svc}
}

func (s *APIKeyService) GetAPIKey(ctx context.Context, keyID gid.GID) (*coredata.PersonalAPIKey, error) {
	var (
		apiKey = &coredata.PersonalAPIKey{}
		now    = time.Now()
	)

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			if err := apiKey.LoadByID(ctx, tx, keyID); err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewPersonalAPIKeyNotFoundError(keyID)
				}
			}

			if apiKey.ExpireReason != nil {
				return NewPersonalAPIKeyExpiredError(keyID)
			}

			if now.After(apiKey.ExpiresAt) {
				apiKey.ExpireReason = ref.Ref(coredata.ExpireReasonIdleTimeout)
				apiKey.ExpiresAt = now
				apiKey.UpdatedAt = now

				if err := apiKey.Update(ctx, tx); err != nil {
					return fmt.Errorf("cannot update personal api key: %w", err)
				}

				return NewPersonalAPIKeyExpiredError(keyID)
			}

			apiKey.LastUsedAt = &now
			apiKey.UpdatedAt = now

			if err := apiKey.Update(ctx, tx); err != nil {
				return fmt.Errorf("cannot update personal api key last used at: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return apiKey, nil
}
