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

package probo

import (
	"context"
	"fmt"
	"time"

	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/probo/coredata"
	"go.gearno.de/kit/pg"
)

type (
	CreateFrameworkRequest struct {
		OrganizationID gid.GID
		Name           string
		Description    string
		ContentRef     string
	}
)

func (s Service) CreateFramework(
	ctx context.Context,
	req CreateFrameworkRequest,
) (*coredata.Framework, error) {
	now := time.Now()
	frameworkID, err := gid.NewGID(s.scope.GetTenantID(), coredata.FrameworkEntityType)
	if err != nil {
		return nil, fmt.Errorf("cannot create global id: %w", err)
	}

	framework := &coredata.Framework{
		ID:             frameworkID,
		OrganizationID: req.OrganizationID,
		Name:           req.Name,
		Description:    req.Description,
		ContentRef:     req.ContentRef,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err = s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return framework.Insert(ctx, conn, s.scope)
		},
	)

	if err != nil {
		return nil, err
	}

	return framework, nil
}
