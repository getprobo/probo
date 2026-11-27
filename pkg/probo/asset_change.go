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

func actorProfileID(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	identityID *gid.GID,
	organizationID gid.GID,
) (*gid.GID, error) {
	if identityID == nil {
		return nil, nil
	}

	profile := &coredata.MembershipProfile{}
	if err := profile.LoadByIdentityIDAndOrganizationID(
		ctx,
		conn,
		scope,
		*identityID,
		organizationID,
	); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("cannot load actor profile: %w", err)
	}

	return &profile.ID, nil
}

func insertAssetChange(
	ctx context.Context,
	conn pg.Tx,
	scope coredata.Scoper,
	before *coredata.Asset,
	after *coredata.Asset,
	input coredata.AssetChangeInput,
	identityID *gid.GID,
	observedAt time.Time,
) error {
	changeType, beforeJSON, afterJSON, write, err := coredata.DiffAssetChange(before, after, input)
	if err != nil {
		return fmt.Errorf("cannot diff asset: %w", err)
	}

	if !write {
		return nil
	}

	actorID, err := actorProfileID(ctx, conn, scope, identityID, after.OrganizationID)
	if err != nil {
		return err
	}

	event := coredata.AssetChangeEvents{
		&coredata.AssetChangeEvent{
			ID:             gid.New(scope.GetTenantID(), coredata.AssetChangeEventEntityType),
			OrganizationID: after.OrganizationID,
			AssetID:        after.ID,
			ObservedAt:     observedAt,
			ActorID:        actorID,
			ChangeType:     changeType,
			Before:         beforeJSON,
			After:          afterJSON,
		},
	}

	if err := event.Insert(ctx, conn, scope); err != nil {
		return fmt.Errorf("cannot insert asset change event: %w", err)
	}

	return nil
}

func recordBusinessFunctionAssetChanges(
	ctx context.Context,
	conn pg.Tx,
	scope coredata.Scoper,
	organizationID gid.GID,
	businessFunctionID gid.GID,
	previousAssetIDs []gid.GID,
	nextAssetIDs []gid.GID,
	identityID *gid.GID,
	observedAt time.Time,
) error {
	previous := gidSet(previousAssetIDs)
	next := gidSet(nextAssetIDs)
	affected := make([]gid.GID, 0)

	for id := range previous {
		if _, ok := next[id]; !ok {
			affected = append(affected, id)
		}
	}

	for id := range next {
		if _, ok := previous[id]; !ok {
			affected = append(affected, id)
		}
	}

	if len(affected) == 0 {
		return nil
	}

	links := &coredata.BusinessFunctionAssets{}
	current, err := links.LoadBusinessFunctionIDsByAssetIDs(ctx, conn, scope, affected)
	if err != nil {
		return fmt.Errorf("cannot load asset business functions: %w", err)
	}

	actorID, err := actorProfileID(ctx, conn, scope, identityID, organizationID)
	if err != nil {
		return err
	}

	events := make(coredata.AssetChangeEvents, 0, len(affected))

	for _, assetID := range affected {
		beforeIDs := current[assetID]

		var afterIDs []gid.GID
		if _, linked := next[assetID]; linked {
			afterIDs = appendGID(beforeIDs, businessFunctionID)
		} else {
			afterIDs = removeGID(beforeIDs, businessFunctionID)
		}

		asset := &coredata.Asset{
			ID:             assetID,
			OrganizationID: organizationID,
		}
		changeType, beforeJSON, afterJSON, write, err := coredata.DiffAssetChange(
			asset,
			asset,
			coredata.AssetChangeInput{
				BusinessFunctions: &coredata.AssetLinkChange{
					Before: beforeIDs,
					After:  afterIDs,
				},
			},
		)
		if err != nil {
			return fmt.Errorf("cannot diff asset business functions: %w", err)
		}

		if !write {
			continue
		}

		events = append(events, &coredata.AssetChangeEvent{
			ID:             gid.New(scope.GetTenantID(), coredata.AssetChangeEventEntityType),
			OrganizationID: organizationID,
			AssetID:        assetID,
			ObservedAt:     observedAt,
			ActorID:        actorID,
			ChangeType:     changeType,
			Before:         beforeJSON,
			After:          afterJSON,
		})
	}

	if err := events.Insert(ctx, conn, scope); err != nil {
		return fmt.Errorf("cannot insert asset change events: %w", err)
	}

	return nil
}

func gidSet(ids []gid.GID) map[gid.GID]struct{} {
	set := make(map[gid.GID]struct{}, len(ids))

	for _, id := range ids {
		set[id] = struct{}{}
	}

	return set
}

func appendGID(ids []gid.GID, id gid.GID) []gid.GID {
	for _, existing := range ids {
		if existing == id {
			return ids
		}
	}

	next := make([]gid.GID, len(ids)+1)
	copy(next, ids)
	next[len(ids)] = id

	return next
}

func removeGID(ids []gid.GID, id gid.GID) []gid.GID {
	next := make([]gid.GID, 0, len(ids))

	for _, existing := range ids {
		if existing != id {
			next = append(next, existing)
		}
	}

	return next
}
