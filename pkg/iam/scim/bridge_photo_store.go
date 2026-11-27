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

package scim

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type bridgePhotoStore struct {
	pg             *pg.Client
	scope          coredata.Scoper
	organizationID gid.GID
}

func (s *bridgePhotoStore) PhotoETags(ctx context.Context) (map[string]string, error) {
	var photos coredata.SCIMProfilePhotos

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return photos.LoadByOrganizationID(ctx, conn, s.scope, s.organizationID)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load scim profile photos: %w", err)
	}

	etags := make(map[string]string, len(photos))
	for _, photo := range photos {
		etags[photo.MembershipProfileID.String()] = photo.ETag
	}

	return etags, nil
}

func (s *bridgePhotoStore) SavePhotoETags(ctx context.Context, etags map[string]string) error {
	var errs []error

	for profileID, etag := range etags {
		id, err := gid.ParseGID(profileID)
		if err != nil {
			errs = append(errs, fmt.Errorf("cannot parse profile id: %w", err))

			continue
		}

		err = s.pg.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				photo := &coredata.SCIMProfilePhoto{}
				if err := photo.LoadByMembershipProfileID(ctx, tx, s.scope, id); err != nil {
					if !errors.Is(err, coredata.ErrResourceNotFound) {
						return fmt.Errorf("cannot load scim profile photo: %w", err)
					}

					now := time.Now()
					photo = &coredata.SCIMProfilePhoto{
						MembershipProfileID: id,
						OrganizationID:      s.organizationID,
						CreatedAt:           now,
						UpdatedAt:           now,
					}
				}

				photo.ETag = etag
				photo.UpdatedAt = time.Now()

				return photo.Upsert(ctx, tx, s.scope)
			},
		)
		if err != nil {
			errs = append(errs, fmt.Errorf("cannot save photo version for profile %q: %w", profileID, err))
		}
	}

	return errors.Join(errs...)
}
