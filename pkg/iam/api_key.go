// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
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
		func(ctx context.Context, tx pg.Tx) error {
			if err := apiKey.LoadByID(ctx, tx, keyID); err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewPersonalAPIKeyNotFoundError(keyID)
				}
			}

			if apiKey.ExpireReason != nil {
				return NewPersonalAPIKeyExpiredError(keyID)
			}

			if now.After(apiKey.ExpiresAt) {
				apiKey.ExpireReason = new(coredata.ExpireReasonIdleTimeout)
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
