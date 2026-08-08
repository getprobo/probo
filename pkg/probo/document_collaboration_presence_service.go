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

package probo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

const (
	documentCollaborationPresenceLimit = 100
	documentCollaborationPresenceTTL   = 15 * time.Second
)

type (
	DocumentCollaborationPresence struct {
		ConnectionID   string
		IdentityID     gid.GID
		AnchorPosition int
		HeadPosition   int
	}
)

func (s *DocumentService) SaveCollaborationPresence(
	ctx context.Context,
	scope coredata.Scoper,
	documentVersionID gid.GID,
	identityID gid.GID,
	connectionID string,
	anchorPosition int,
	headPosition int,
) error {
	if anchorPosition < 0 || headPosition < 0 {
		return fmt.Errorf("document collaboration presence positions cannot be negative")
	}

	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			version, err := s.loadEditableCollaborationVersion(
				ctx,
				tx,
				scope,
				documentVersionID,
			)
			if err != nil {
				return err
			}

			now := time.Now()

			presence := coredata.DocumentVersionCollaborationPresence{
				ConnectionID:      connectionID,
				DocumentVersionID: documentVersionID,
				OrganizationID:    version.OrganizationID,
				IdentityID:        identityID,
				AnchorPosition:    anchorPosition,
				HeadPosition:      headPosition,
				UpdatedAt:         now,
			}
			if err := presence.Update(ctx, tx, scope); err != nil {
				if !errors.Is(err, coredata.ErrResourceNotFound) {
					return fmt.Errorf("cannot update document collaboration presence: %w", err)
				}

				if err := presence.Insert(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot insert document collaboration presence: %w", err)
				}
			}

			if err := coredata.DeleteExpiredDocumentVersionCollaborationPresences(
				ctx,
				tx,
				scope,
				documentVersionID,
				now.Add(-time.Hour),
			); err != nil {
				return fmt.Errorf("cannot clean document collaboration presences: %w", err)
			}

			return nil
		},
	)
}

func (s *DocumentService) ListCollaborationPresences(
	ctx context.Context,
	scope coredata.Scoper,
	documentVersionID gid.GID,
	excludeConnectionID string,
) ([]DocumentCollaborationPresence, error) {
	var stored coredata.DocumentVersionCollaborationPresences

	if err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return stored.Load(
				ctx,
				conn,
				scope,
				documentVersionID,
				excludeConnectionID,
				time.Now().Add(-documentCollaborationPresenceTTL),
				documentCollaborationPresenceLimit,
			)
		},
	); err != nil {
		return nil, fmt.Errorf("cannot load document collaboration presences: %w", err)
	}

	presences := make([]DocumentCollaborationPresence, len(stored))
	for i, presence := range stored {
		presences[i] = DocumentCollaborationPresence{
			ConnectionID:   presence.ConnectionID,
			IdentityID:     presence.IdentityID,
			AnchorPosition: presence.AnchorPosition,
			HeadPosition:   presence.HeadPosition,
		}
	}

	return presences, nil
}

func (s *DocumentService) DeleteCollaborationPresence(
	ctx context.Context,
	scope coredata.Scoper,
	connectionID string,
) error {
	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			presence := coredata.DocumentVersionCollaborationPresence{
				ConnectionID: connectionID,
			}
			if err := presence.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete document collaboration presence: %w", err)
			}

			return nil
		},
	)
}
