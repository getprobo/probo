// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package github

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func resolveGitHubThirdParty(
	ctx context.Context,
	conn pg.Tx,
	scope coredata.Scoper,
	organizationID gid.GID,
) (*coredata.ThirdParty, error) {
	thirdParty := &coredata.ThirdParty{}

	err := thirdParty.LoadByNameAndOrganizationID(ctx, conn, scope, thirdPartyName, organizationID)
	if err == nil {
		if thirdParty.Level == 1 && thirdParty.ParentThirdPartyID == nil {
			return thirdParty, nil
		}

		return nil, fmt.Errorf("third party %q exists but is not level 1", thirdPartyName)
	}

	if !isNotFound(err) {
		return nil, fmt.Errorf("cannot load github third party: %w", err)
	}

	now := time.Now()

	thirdParty = &coredata.ThirdParty{
		ID:             gid.New(scope.GetTenantID(), coredata.ThirdPartyEntityType),
		OrganizationID: organizationID,
		Name:           thirdPartyName,
		Category:       coredata.ThirdPartyCategoryVersionControl,
		WebsiteURL:     new("https://github.com"),
		Level:          1,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := thirdParty.Insert(ctx, conn, scope); err != nil {
		return nil, fmt.Errorf("cannot insert github third party: %w", err)
	}

	return thirdParty, nil
}

func loadGitHubLinkedMeasures(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	thirdPartyID gid.GID,
) ([]ExistingMeasure, error) {
	var measures coredata.Measures

	if err := measures.LoadByThirdPartyID(
		ctx,
		conn,
		scope,
		thirdPartyID,
		nil,
		coredata.NewMeasureFilter(nil, nil, nil),
	); err != nil {
		return nil, fmt.Errorf("cannot load github-linked measures: %w", err)
	}

	out := make([]ExistingMeasure, 0, len(measures))

	for _, m := range measures {
		out = append(out, ExistingMeasure{
			ID:          m.ID,
			Name:        m.Name,
			Description: m.Description,
			Category:    m.Category,
			State:       m.State,
		})
	}

	return out, nil
}

func isNotFound(err error) bool {
	return errors.Is(err, coredata.ErrResourceNotFound)
}

// EnsureThirdParty resolves or creates the level-1 GitHub third party.
func EnsureThirdParty(
	ctx context.Context,
	pgClient *pg.Client,
	scope coredata.Scoper,
	organizationID gid.GID,
) (*coredata.ThirdParty, error) {
	var thirdParty *coredata.ThirdParty

	err := pgClient.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			tp, err := resolveGitHubThirdParty(ctx, tx, scope, organizationID)
			if err != nil {
				return err
			}

			thirdParty = tp

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot ensure github third party: %w", err)
	}

	return thirdParty, nil
}
