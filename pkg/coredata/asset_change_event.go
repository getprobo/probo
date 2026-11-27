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

package coredata

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

const assetChangeEventBatchSize = 500

type (
	AssetChangeEvent struct {
		ID             gid.GID         `db:"id"`
		OrganizationID gid.GID         `db:"organization_id"`
		AssetID        gid.GID         `db:"asset_id"`
		ObservedAt     time.Time       `db:"observed_at"`
		AssetSyncID    *gid.GID        `db:"asset_sync_id"`
		ActorID        *gid.GID        `db:"actor_profile_id"`
		ChangeType     AssetChangeType `db:"change_type"`
		Before         json.RawMessage `db:"before"`
		After          json.RawMessage `db:"after"`
	}

	AssetChangeEvents []*AssetChangeEvent

	AssetLinkChange struct {
		Before []gid.GID
		After  []gid.GID
	}

	AssetChangeInput struct {
		ThirdParties      *AssetLinkChange
		BusinessFunctions *AssetLinkChange
		RecordLastSeen    bool
	}
)

func (e AssetChangeEvent) CursorKey(orderBy AssetChangeEventOrderField) page.CursorKey {
	switch orderBy {
	case AssetChangeEventOrderFieldObservedAt:
		return page.NewCursorKey(e.ID, e.ObservedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (e *AssetChangeEvent) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `
SELECT
	id,
	organization_id
FROM
	asset_change_events
WHERE
	id = ANY(@resource_ids)
`

	args := pgx.StrictNamedArgs{"resource_ids": resourceIDs}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query authorization attributes: %w", err)
	}

	defer rows.Close()

	attrsByID := make(policy.AttributesByID)

	for rows.Next() {
		var id, organizationID gid.GID

		if err := rows.Scan(&id, &organizationID); err != nil {
			return nil, fmt.Errorf("cannot scan authorization attributes: %w", err)
		}

		attrsByID[id] = policy.Attributes{
			"organization_id": organizationID.String(),
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate authorization attributes: %w", err)
	}

	return attrsByID, nil
}

func (e *AssetChangeEvent) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	eventID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	asset_id,
	observed_at,
	asset_sync_id,
	actor_profile_id,
	change_type,
	before,
	after
FROM
	asset_change_events
WHERE
	%s
	AND id = @event_id
LIMIT 1
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"event_id": eventID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query asset change events: %w", err)
	}

	event, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AssetChangeEvent])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect asset change event: %w", err)
	}

	*e = event

	return nil
}

func (e *AssetChangeEvents) LoadByAssetID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	assetID gid.GID,
	cursor *page.Cursor[AssetChangeEventOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	asset_id,
	observed_at,
	asset_sync_id,
	actor_profile_id,
	change_type,
	before,
	after
FROM
	asset_change_events
WHERE
	%s
	AND asset_id = @asset_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"asset_id": assetID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query asset change events: %w", err)
	}

	events, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AssetChangeEvent])
	if err != nil {
		return fmt.Errorf("cannot collect asset change events: %w", err)
	}

	*e = events

	return nil
}

func (e *AssetChangeEvents) CountByAssetID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	assetID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	asset_change_events
WHERE
	%s
	AND asset_id = @asset_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"asset_id": assetID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count asset change events: %w", err)
	}

	return count, nil
}

func (e AssetChangeEvents) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	if len(e) == 0 {
		return nil
	}

	for start := 0; start < len(e); start += assetChangeEventBatchSize {
		end := min(start+assetChangeEventBatchSize, len(e))

		if err := insertAssetChangeEventBatch(ctx, conn, scope, e[start:end]); err != nil {
			return err
		}
	}

	return nil
}

func insertAssetChangeEventBatch(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	events AssetChangeEvents,
) error {
	ids := make([]string, len(events))
	organizationIDs := make([]string, len(events))
	assetIDs := make([]string, len(events))
	observedAts := make([]time.Time, len(events))
	assetSyncIDs := make([]*string, len(events))
	actorIDs := make([]*string, len(events))
	changeTypes := make([]string, len(events))
	befores := make([]*string, len(events))
	afters := make([]*string, len(events))

	for i, event := range events {
		ids[i] = event.ID.String()
		organizationIDs[i] = event.OrganizationID.String()
		assetIDs[i] = event.AssetID.String()
		observedAts[i] = event.ObservedAt
		assetSyncIDs[i] = gidString(event.AssetSyncID)
		actorIDs[i] = gidString(event.ActorID)
		changeTypes[i] = string(event.ChangeType)
		befores[i] = rawJSONString(event.Before)
		afters[i] = rawJSONString(event.After)
	}

	q := `
INSERT INTO asset_change_events (
	id,
	tenant_id,
	organization_id,
	asset_id,
	observed_at,
	asset_sync_id,
	actor_profile_id,
	change_type,
	before,
	after
)
SELECT
	id,
	@tenant_id,
	organization_id,
	asset_id,
	observed_at,
	asset_sync_id,
	actor_profile_id,
	change_type::asset_change_type,
	before::jsonb,
	after::jsonb
FROM unnest(
	@ids::text[],
	@organization_ids::text[],
	@asset_ids::text[],
	@observed_ats::timestamptz[],
	@asset_sync_ids::text[],
	@actor_profile_ids::text[],
	@change_types::text[],
	@befores::text[],
	@afters::text[]
) AS event(
	id,
	organization_id,
	asset_id,
	observed_at,
	asset_sync_id,
	actor_profile_id,
	change_type,
	before,
	after
)
`

	args := pgx.StrictNamedArgs{
		"tenant_id":         scope.GetTenantID(),
		"ids":               ids,
		"organization_ids":  organizationIDs,
		"asset_ids":         assetIDs,
		"observed_ats":      observedAts,
		"asset_sync_ids":    assetSyncIDs,
		"actor_profile_ids": actorIDs,
		"change_types":      changeTypes,
		"befores":           befores,
		"afters":            afters,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot insert asset change events: %w", err)
	}

	return nil
}

func DiscoveredAssetChangeEvents(
	before Assets,
	after Assets,
	syncID gid.GID,
	observedAt time.Time,
) (AssetChangeEvents, error) {
	previous := make(map[string]*Asset, len(before))

	for _, asset := range before {
		key, ok := discoveredIdentityKey(asset)
		if !ok {
			continue
		}

		previous[key] = asset
	}

	events := make(AssetChangeEvents, 0, len(after))

	for _, asset := range after {
		key, ok := discoveredIdentityKey(asset)
		if !ok {
			continue
		}

		changeType, beforeJSON, afterJSON, write, err := DiffAssetChange(
			previous[key],
			asset,
			AssetChangeInput{RecordLastSeen: true},
		)
		if err != nil {
			return nil, fmt.Errorf("cannot diff discovered asset: %w", err)
		}

		if !write {
			continue
		}

		runID := syncID
		events = append(events, &AssetChangeEvent{
			ID:             gid.New(asset.ID.TenantID(), AssetChangeEventEntityType),
			OrganizationID: asset.OrganizationID,
			AssetID:        asset.ID,
			ObservedAt:     observedAt,
			AssetSyncID:    &runID,
			ChangeType:     changeType,
			Before:         beforeJSON,
			After:          afterJSON,
		})
	}

	return events, nil
}

func DiffAssetChange(
	before *Asset,
	after *Asset,
	input AssetChangeInput,
) (AssetChangeType, json.RawMessage, json.RawMessage, bool, error) {
	if after == nil {
		return "", nil, nil, false, fmt.Errorf("cannot diff asset change without the current row")
	}

	if before == nil {
		payload, err := createdAssetChange(after, input)
		if err != nil {
			return "", nil, nil, false, err
		}

		return AssetChangeTypeCreated, nil, payload, true, nil
	}

	beforeMap := map[string]any{}
	afterMap := map[string]any{}
	changed := false

	recordChange(&changed, beforeMap, afterMap, "name", before.Name != after.Name, before.Name, after.Name)
	recordChange(&changed, beforeMap, afterMap, "amount", before.Amount != after.Amount, before.Amount, after.Amount)
	recordChange(
		&changed,
		beforeMap,
		afterMap,
		"asset_type",
		before.Type != after.Type,
		before.Type.String(),
		after.Type.String(),
	)
	recordChange(
		&changed,
		beforeMap,
		afterMap,
		"owner_profile_id",
		!sameGID(before.OwnerID, after.OwnerID),
		gidValue(before.OwnerID),
		gidValue(after.OwnerID),
	)
	recordChange(
		&changed,
		beforeMap,
		afterMap,
		"data_types_stored",
		!sameString(before.DataTypesStored, after.DataTypesStored),
		stringValue(before.DataTypesStored),
		stringValue(after.DataTypesStored),
	)
	recordChange(
		&changed,
		beforeMap,
		afterMap,
		"kind",
		!sameString(before.Kind, after.Kind),
		stringValue(before.Kind),
		stringValue(after.Kind),
	)
	recordChange(
		&changed,
		beforeMap,
		afterMap,
		"external_id",
		!sameString(before.ExternalID, after.ExternalID),
		stringValue(before.ExternalID),
		stringValue(after.ExternalID),
	)
	recordChange(
		&changed,
		beforeMap,
		afterMap,
		"region",
		!sameString(before.Region, after.Region),
		stringValue(before.Region),
		stringValue(after.Region),
	)
	recordChange(
		&changed,
		beforeMap,
		afterMap,
		"configuration",
		!configEqual(before.Configuration, after.Configuration),
		configValue(before.Configuration),
		configValue(after.Configuration),
	)
	recordChange(
		&changed,
		beforeMap,
		afterMap,
		"schema_version",
		!sameInt(before.SchemaVersion, after.SchemaVersion),
		intValue(before.SchemaVersion),
		intValue(after.SchemaVersion),
	)
	recordChange(
		&changed,
		beforeMap,
		afterMap,
		"status",
		before.Status != after.Status,
		before.Status.String(),
		after.Status.String(),
	)
	recordChange(
		&changed,
		beforeMap,
		afterMap,
		"connector_id",
		!sameGID(before.ConnectorID, after.ConnectorID),
		gidValue(before.ConnectorID),
		gidValue(after.ConnectorID),
	)
	recordChange(
		&changed,
		beforeMap,
		afterMap,
		"last_seen_at",
		input.RecordLastSeen || !sameTime(before.LastSeenAt, after.LastSeenAt),
		timeValue(before.LastSeenAt),
		timeValue(after.LastSeenAt),
	)

	if input.ThirdParties != nil && !sameIDSet(input.ThirdParties.Before, input.ThirdParties.After) {
		recordChange(
			&changed,
			beforeMap,
			afterMap,
			"third_party_ids",
			true,
			idList(input.ThirdParties.Before),
			idList(input.ThirdParties.After),
		)
	}

	if input.BusinessFunctions != nil && !sameIDSet(input.BusinessFunctions.Before, input.BusinessFunctions.After) {
		recordChange(
			&changed,
			beforeMap,
			afterMap,
			"business_function_ids",
			true,
			idList(input.BusinessFunctions.Before),
			idList(input.BusinessFunctions.After),
		)
	}

	if !changed {
		return "", nil, nil, false, nil
	}

	beforeJSON, err := json.Marshal(beforeMap)
	if err != nil {
		return "", nil, nil, false, fmt.Errorf("cannot encode asset change before: %w", err)
	}

	afterJSON, err := json.Marshal(afterMap)
	if err != nil {
		return "", nil, nil, false, fmt.Errorf("cannot encode asset change after: %w", err)
	}

	changeType := AssetChangeTypeUpdated
	if after.Status == AssetStatusDeactivated && before.Status != AssetStatusDeactivated {
		changeType = AssetChangeTypeDeactivated
	}

	return changeType, beforeJSON, afterJSON, true, nil
}

func createdAssetChange(asset *Asset, input AssetChangeInput) (json.RawMessage, error) {
	payload := map[string]any{
		"name":       asset.Name,
		"amount":     asset.Amount,
		"asset_type": asset.Type.String(),
		"status":     asset.Status.String(),
	}

	if asset.OwnerID != nil {
		payload["owner_profile_id"] = asset.OwnerID.String()
	}

	if asset.DataTypesStored != nil {
		payload["data_types_stored"] = *asset.DataTypesStored
	}

	if asset.Kind != nil {
		payload["kind"] = *asset.Kind
	}

	if asset.ExternalID != nil {
		payload["external_id"] = *asset.ExternalID
	}

	if asset.Region != nil {
		payload["region"] = *asset.Region
	}

	if len(asset.Configuration) > 0 {
		payload["configuration"] = json.RawMessage(asset.Configuration)
	}

	if asset.SchemaVersion != nil {
		payload["schema_version"] = *asset.SchemaVersion
	}

	if asset.ConnectorID != nil {
		payload["connector_id"] = asset.ConnectorID.String()
	}

	if asset.LastSeenAt != nil {
		payload["last_seen_at"] = *asset.LastSeenAt
	}

	if input.ThirdParties != nil && len(input.ThirdParties.After) > 0 {
		payload["third_party_ids"] = idList(input.ThirdParties.After)
	}

	if input.BusinessFunctions != nil && len(input.BusinessFunctions.After) > 0 {
		payload["business_function_ids"] = idList(input.BusinessFunctions.After)
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("cannot encode created asset change: %w", err)
	}

	return encoded, nil
}

func recordChange(
	changed *bool,
	beforeMap map[string]any,
	afterMap map[string]any,
	key string,
	differ bool,
	beforeValue any,
	afterValue any,
) {
	if !differ {
		return
	}

	beforeMap[key] = beforeValue
	afterMap[key] = afterValue
	*changed = true
}

func gidString(id *gid.GID) *string {
	if id == nil {
		return nil
	}

	value := id.String()

	return &value
}

func rawJSONString(raw json.RawMessage) *string {
	if len(raw) == 0 {
		return nil
	}

	value := string(raw)

	return &value
}

func sameGID(left *gid.GID, right *gid.GID) bool {
	if left == nil || right == nil {
		return left == right
	}

	return *left == *right
}

func gidValue(id *gid.GID) any {
	if id == nil {
		return nil
	}

	return id.String()
}

func sameString(left *string, right *string) bool {
	if left == nil || right == nil {
		return left == right
	}

	return *left == *right
}

func stringValue(value *string) any {
	if value == nil {
		return nil
	}

	return *value
}

func sameInt(left *int, right *int) bool {
	if left == nil || right == nil {
		return left == right
	}

	return *left == *right
}

func intValue(value *int) any {
	if value == nil {
		return nil
	}

	return *value
}

func sameTime(left *time.Time, right *time.Time) bool {
	if left == nil || right == nil {
		return left == right
	}

	return left.Equal(*right)
}

func timeValue(value *time.Time) any {
	if value == nil {
		return nil
	}

	return *value
}

func configEqual(left json.RawMessage, right json.RawMessage) bool {
	if len(left) == 0 && len(right) == 0 {
		return true
	}

	if bytes.Equal(left, right) {
		return true
	}

	var decodedLeft any

	var decodedRight any

	if json.Unmarshal(left, &decodedLeft) != nil || json.Unmarshal(right, &decodedRight) != nil {
		return false
	}

	return reflect.DeepEqual(decodedLeft, decodedRight)
}

func configValue(raw json.RawMessage) any {
	if len(raw) == 0 {
		return nil
	}

	return json.RawMessage(raw)
}

func sameIDSet(left []gid.GID, right []gid.GID) bool {
	return slices.Equal(sortedIDStrings(left), sortedIDStrings(right))
}

func idList(ids []gid.GID) []string {
	return sortedIDStrings(ids)
}

func sortedIDStrings(ids []gid.GID) []string {
	values := make([]string, len(ids))

	for i, id := range ids {
		values[i] = id.String()
	}

	slices.Sort(values)

	return values
}

// DeleteAssetChangeEventsBeforeBatch deletes events observed strictly before
// the cutoff. A name is not part of an event, and nothing else keeps one.
//
// No ORDER BY observed_at. There is no index leading with observed_at, so
// each ordered batch would sort every expired row. Any expired row is
// eligible. SKIP LOCKED lets another replica take a different batch.
func DeleteAssetChangeEventsBeforeBatch(
	ctx context.Context,
	conn pg.Querier,
	before time.Time,
	batchSize int,
) (int64, error) {
	q := `
WITH doomed AS (
	SELECT id
	FROM asset_change_events
	WHERE observed_at < @before
	LIMIT @limit
	FOR UPDATE SKIP LOCKED
)
DELETE FROM asset_change_events
WHERE id IN (SELECT id FROM doomed)
`

	result, err := conn.Exec(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"before": before,
			"limit":  batchSize,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("cannot delete asset change events: %w", err)
	}

	return result.RowsAffected(), nil
}
