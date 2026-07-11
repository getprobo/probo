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

package coredata

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
)

type (
	DocumentVersionSignature struct {
		ID                    gid.GID                       `json:"id" db:"id"`
		OrganizationID        gid.GID                       `json:"-" db:"organization_id"`
		DocumentVersionID     gid.GID                       `json:"document_version_id" db:"document_version_id"`
		State                 DocumentVersionSignatureState `json:"state" db:"state"`
		SignedBy              gid.GID                       `json:"signed_by" db:"signed_by_profile_id"`
		SignedAt              *time.Time                    `json:"signed_at" db:"signed_at"`
		RequestedAt           time.Time                     `json:"requested_at" db:"requested_at"`
		ElectronicSignatureID *gid.GID                      `json:"-" db:"electronic_signature_id"`
		CreatedAt             time.Time                     `json:"created_at" db:"created_at"`
		UpdatedAt             time.Time                     `json:"updated_at" db:"updated_at"`
	}

	DocumentVersionSignatures []*DocumentVersionSignature

	DocumentVersionSignatureWithPeople struct {
		DocumentVersionSignature
		SignedByFullName string `db:"signed_by_full_name"`
	}

	DocumentVersionSignaturesWithPeople []*DocumentVersionSignatureWithPeople
)

func (pvs DocumentVersionSignature) CursorKey(orderBy DocumentVersionSignatureOrderField) page.CursorKey {
	switch orderBy {
	case DocumentVersionSignatureOrderFieldCreatedAt:
		return page.NewCursorKey(pvs.ID, pvs.CreatedAt)
	case DocumentVersionSignatureOrderFieldSignedAt:
		return page.NewCursorKey(pvs.ID, pvs.SignedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

// AuthorizationAttributes returns the authorization attributes for policy evaluation.
func (dvs *DocumentVersionSignature) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM document_version_signatures WHERE id = ANY(@resource_ids::text[])`

	args := pgx.StrictNamedArgs{
		"resource_ids": resourceIDs,
	}

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

func (pvs *DocumentVersionSignature) LoadByDocumentVersionIDAndSignatory(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentVersionID gid.GID,
	signatory gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	document_version_id,
	state,
	signed_by_profile_id,
	signed_at,
	requested_at,
	electronic_signature_id,
	created_at,
	updated_at
FROM
	document_version_signatures
WHERE
	%s
	AND document_version_id = @document_version_id
	AND signed_by_profile_id = @signatory
LIMIT 1
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"document_version_id": documentVersionID, "signatory": signatory}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document version signature: %w", err)
	}

	documentVersionSignature, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[DocumentVersionSignature])
	if err != nil {
		return fmt.Errorf("cannot collect document version signature: %w", err)
	}

	*pvs = documentVersionSignature

	return nil
}

// LoadByDocumentMajorAndSignatory loads the signatory's existing signature for
// the whole major that owns documentVersionID, scanning across every minor
// version of that major. A signed signature is preferred over a still pending
// one, then the most recent. It returns ErrResourceNotFound when the signatory
// has no signature anywhere in the major.
func (pvs *DocumentVersionSignature) LoadByDocumentMajorAndSignatory(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentVersionID gid.GID,
	signatory gid.GID,
) error {
	q := `
WITH source_version AS (
	SELECT document_id, major FROM document_versions WHERE id = @document_version_id
),
major_versions AS (
	SELECT dv.id FROM document_versions dv
	INNER JOIN source_version sv ON dv.document_id = sv.document_id AND dv.major = sv.major
),
major_signatures AS (
	SELECT
		dvs.id,
		dvs.organization_id,
		dvs.tenant_id,
		dvs.document_version_id,
		dvs.state,
		dvs.signed_by_profile_id,
		dvs.signed_at,
		dvs.requested_at,
		dvs.electronic_signature_id,
		dvs.created_at,
		dvs.updated_at
	FROM document_version_signatures dvs
	INNER JOIN major_versions mv ON dvs.document_version_id = mv.id
	WHERE dvs.signed_by_profile_id = @signatory
)
SELECT
	id,
	organization_id,
	document_version_id,
	state,
	signed_by_profile_id,
	signed_at,
	requested_at,
	electronic_signature_id,
	created_at,
	updated_at
FROM
	major_signatures
WHERE
	%s
ORDER BY
	CASE state WHEN 'SIGNED' THEN 0 ELSE 1 END,
	created_at DESC
LIMIT 1
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"document_version_id": documentVersionID, "signatory": signatory}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document version signature by major: %w", err)
	}

	documentVersionSignature, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[DocumentVersionSignature])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect document version signature by major: %w", err)
	}

	*pvs = documentVersionSignature

	return nil
}

func (pvs *DocumentVersionSignature) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	signatureID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	document_version_id,
	state,
	signed_by_profile_id,
	signed_at,
	requested_at,
	electronic_signature_id,
	created_at,
	updated_at
FROM
	document_version_signatures
WHERE
	id = @document_version_signature_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"document_version_signature_id": signatureID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document version signature: %w", err)
	}

	documentVersionSignature, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[DocumentVersionSignature])
	if err != nil {
		return fmt.Errorf("cannot collect document version signature: %w", err)
	}

	*pvs = documentVersionSignature

	return nil
}

func (pvs DocumentVersionSignature) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO document_version_signatures (
	id,
	tenant_id,
	organization_id,
	document_version_id,
	state,
	signed_by_profile_id,
	signed_at,
	requested_at,
	electronic_signature_id,
	created_at,
	updated_at,
	notification_count
) VALUES (
 	@id,
	@tenant_id,
	@organization_id,
	@document_version_id,
	@state,
	@signed_by_profile_id,
	@signed_at,
	@requested_at,
	@electronic_signature_id,
	@created_at,
	@updated_at,
	@notification_count
)
`

	args := pgx.StrictNamedArgs{
		"id":                      pvs.ID,
		"tenant_id":               scope.GetTenantID(),
		"organization_id":         pvs.OrganizationID,
		"document_version_id":     pvs.DocumentVersionID,
		"state":                   pvs.State,
		"signed_by_profile_id":    pvs.SignedBy,
		"signed_at":               pvs.SignedAt,
		"requested_at":            pvs.RequestedAt,
		"electronic_signature_id": pvs.ElectronicSignatureID,
		"created_at":              pvs.CreatedAt,
		"updated_at":              pvs.UpdatedAt,
		"notification_count":      0,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "policy_version_signatures_policy_version_id_signed_by_key" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert document version signature: %w", err)
	}

	return nil
}

func (pvss *DocumentVersionSignatures) LoadByDocumentVersionID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentVersionID gid.GID,
	cursor *page.Cursor[DocumentVersionSignatureOrderField],
	filter *DocumentVersionSignatureFilter,
) error {
	q := `
WITH source_version AS (
	SELECT document_id, major FROM document_versions WHERE id = @document_version_id
),
major_versions AS (
	SELECT dv.id FROM document_versions dv
	INNER JOIN source_version sv ON dv.document_id = sv.document_id AND dv.major = sv.major
)
SELECT
	document_version_signatures.id,
	document_version_signatures.organization_id,
	document_version_signatures.document_version_id,
	document_version_signatures.state,
	document_version_signatures.signed_by_profile_id,
	document_version_signatures.signed_at,
	document_version_signatures.requested_at,
	document_version_signatures.electronic_signature_id,
	document_version_signatures.created_at,
	document_version_signatures.updated_at
FROM
	document_version_signatures
INNER JOIN major_versions mv ON document_version_signatures.document_version_id = mv.id
WHERE
	%s
	AND %s
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"document_version_id": documentVersionID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document version signatures: %w", err)
	}

	documentVersionSignatures, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[DocumentVersionSignature])
	if err != nil {
		return fmt.Errorf("cannot collect document version signatures: %w", err)
	}

	*pvss = documentVersionSignatures

	return nil
}

// documentNotificationMaxCount caps how many emails a signature/approval
// request gets: the first notice plus three reminders. A request is due for its
// next email once it is past its scheduled offset — the first email after the
// debounce delay, then reminders at 1x, 2x and 3x the reminder interval after
// the previous email — and stops once it reaches this cap.
const documentNotificationMaxCount = 4

func remainingNotificationIDs(all []gid.GID, claimed []gid.GID) []gid.GID {
	claimedSet := make(map[gid.GID]struct{}, len(claimed))
	for _, id := range claimed {
		claimedSet[id] = struct{}{}
	}

	rest := make([]gid.GID, 0, len(all))
	for _, id := range all {
		if _, ok := claimedSet[id]; ok {
			continue
		}

		rest = append(rest, id)
	}

	return rest
}

// LoadNextDueGroupForNotification loads every still-REQUESTED signature for the
// next (organization, signatory) group that has at least one request due for a
// notification, so the whole group can be emailed together. A request is due
// once it is past its scheduled offset (see documentNotificationMaxCount) and
// has not reached the cap. The receiver is left empty when no group is due.
func (pvss *DocumentVersionSignatures) LoadNextDueGroupForNotification(
	ctx context.Context,
	conn pg.Querier,
	now time.Time,
	debounceBefore time.Time,
	reminderInterval time.Duration,
) error {
	q := `
WITH next_group AS (
	SELECT
		organization_id,
		signed_by_profile_id
	FROM
		document_version_signatures
	WHERE
		state = @state
		AND notification_count < @max_notifications
		AND (
			(notification_count = 0 AND requested_at < @debounce_before)
			OR (notification_count > 0 AND last_notified_at < @now::timestamptz - make_interval(secs => @reminder_interval_seconds * notification_count))
		)
		AND EXISTS (
			SELECT 1
			FROM document_versions dv
			JOIN documents doc ON doc.id = dv.document_id
			WHERE dv.id = document_version_signatures.document_version_id
				AND doc.deleted_at IS NULL
				AND doc.archived_at IS NULL
		)
		AND EXISTS (
			SELECT 1
			FROM iam_membership_profiles p
			WHERE p.id = document_version_signatures.signed_by_profile_id
				AND p.state = @recipient_state::membership_state
				AND (p.contract_end_date IS NULL OR p.contract_end_date >= @now::date)
		)
	GROUP BY
		organization_id,
		signed_by_profile_id
	ORDER BY
		organization_id,
		signed_by_profile_id
	LIMIT 1
)
SELECT
	s.id,
	s.organization_id,
	s.document_version_id,
	s.state,
	s.signed_by_profile_id,
	s.signed_at,
	s.requested_at,
	s.electronic_signature_id,
	s.created_at,
	s.updated_at
FROM
	document_version_signatures s
INNER JOIN next_group g
	ON g.organization_id = s.organization_id
	AND g.signed_by_profile_id = s.signed_by_profile_id
WHERE
	s.state = @state
	AND s.notification_count < @max_notifications
	AND EXISTS (
		SELECT 1
		FROM document_versions dv
		JOIN documents doc ON doc.id = dv.document_id
		WHERE dv.id = s.document_version_id
			AND doc.deleted_at IS NULL
			AND doc.archived_at IS NULL
	)
ORDER BY
	s.document_version_id
`

	args := pgx.StrictNamedArgs{
		"state":                     DocumentVersionSignatureStateRequested,
		"recipient_state":           ProfileStateActive,
		"max_notifications":         documentNotificationMaxCount,
		"now":                       now,
		"debounce_before":           debounceBefore,
		"reminder_interval_seconds": reminderInterval.Seconds(),
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query due signatures for notification: %w", err)
	}

	signatures, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[DocumentVersionSignature])
	if err != nil {
		return fmt.Errorf("cannot collect due signatures for notification: %w", err)
	}

	*pvss = signatures

	return nil
}

// ClaimForNotification claims the receiver's signatures that are individually
// due for a notification and returns their ids. The conditional update doubles
// as the claim, so concurrent workers never email the same group twice. Callers
// advance the rest of the group with BumpRemainingForNotification in the same
// transaction.
func (pvss DocumentVersionSignatures) ClaimForNotification(
	ctx context.Context,
	conn pg.Tx,
	now time.Time,
	debounceBefore time.Time,
	reminderInterval time.Duration,
) ([]gid.GID, error) {
	ids := make([]gid.GID, len(pvss))
	for i, signature := range pvss {
		ids[i] = signature.ID
	}

	q := `
UPDATE document_version_signatures
SET
	notification_count = notification_count + 1,
	last_notified_at = @now
WHERE
	id = ANY(@ids::text[])
	AND state = @state
	AND notification_count < @max_notifications
	AND (
		(notification_count = 0 AND requested_at < @debounce_before)
		OR (notification_count > 0 AND last_notified_at < @now::timestamptz - make_interval(secs => @reminder_interval_seconds * notification_count))
	)
RETURNING id
`

	rows, err := conn.Query(ctx, q, pgx.StrictNamedArgs{
		"ids":                       ids,
		"state":                     DocumentVersionSignatureStateRequested,
		"max_notifications":         documentNotificationMaxCount,
		"now":                       now,
		"debounce_before":           debounceBefore,
		"reminder_interval_seconds": reminderInterval.Seconds(),
	})
	if err != nil {
		return nil, fmt.Errorf("cannot claim due signatures for notification: %w", err)
	}

	claimed, err := pgx.CollectRows(rows, pgx.RowTo[gid.GID])
	if err != nil {
		return nil, fmt.Errorf("cannot collect claimed signatures for notification: %w", err)
	}

	return claimed, nil
}

// BumpRemainingForNotification advances the notification schedule for the
// still-pending signatures in the group that were not individually claimed, so
// the whole emailed list moves forward together. It must run in the same
// transaction as ClaimForNotification.
func (pvss DocumentVersionSignatures) BumpRemainingForNotification(
	ctx context.Context,
	conn pg.Tx,
	claimed []gid.GID,
	now time.Time,
) ([]gid.GID, error) {
	ids := make([]gid.GID, len(pvss))
	for i, signature := range pvss {
		ids[i] = signature.ID
	}

	rest := remainingNotificationIDs(ids, claimed)
	if len(rest) == 0 {
		return nil, nil
	}

	q := `
UPDATE document_version_signatures
SET
	notification_count = notification_count + 1,
	last_notified_at = @now
WHERE
	id = ANY(@ids::text[])
	AND state = @state
	AND notification_count < @max_notifications
RETURNING id
`

	rows, err := conn.Query(ctx, q, pgx.StrictNamedArgs{
		"ids":               rest,
		"state":             DocumentVersionSignatureStateRequested,
		"max_notifications": documentNotificationMaxCount,
		"now":               now,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot bump remaining signatures for notification: %w", err)
	}

	bumped, err := pgx.CollectRows(rows, pgx.RowTo[gid.GID])
	if err != nil {
		return nil, fmt.Errorf("cannot collect bumped signatures for notification: %w", err)
	}

	return bumped, nil
}

func (pvs *DocumentVersionSignature) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE document_version_signatures
SET
	state = @state,
	signed_by_profile_id = @signed_by_profile_id,
	signed_at = @signed_at,
	electronic_signature_id = @electronic_signature_id,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                      pvs.ID,
		"state":                   pvs.State,
		"signed_by_profile_id":    pvs.SignedBy,
		"signed_at":               pvs.SignedAt,
		"electronic_signature_id": pvs.ElectronicSignatureID,
		"updated_at":              pvs.UpdatedAt,
	}

	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update document version signature: %w", err)
	}

	return nil
}

func (pvs *DocumentVersionSignature) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	documentVersionSignatureID gid.GID,
) error {
	q := `
DELETE FROM document_version_signatures
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": documentVersionSignatureID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete document version signature: %w", err)
	}

	return nil
}

func (pvss *DocumentVersionSignatures) DeleteRequestedBySignatory(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	signatoryID gid.GID,
) error {
	q := `
DELETE FROM document_version_signatures
WHERE
	%s
	AND signed_by_profile_id = @signatory_id
	AND state = @state
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"signatory_id": signatoryID,
		"state":        DocumentVersionSignatureStateRequested,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot delete requested document version signatures: %w", err)
	}

	return nil
}

func (pvss *DocumentVersionSignatures) MoveRequestedToVersionWithinMajor(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	targetVersionID gid.GID,
) error {
	q := `
UPDATE document_version_signatures
SET
	document_version_id = @target_version_id,
	updated_at = @now
WHERE
	%s
	AND state = @state
	AND document_version_id <> @target_version_id
	AND document_version_id IN (
		SELECT dv.id
		FROM document_versions dv
		INNER JOIN document_versions target
			ON target.document_id = dv.document_id
			AND target.major = dv.major
		WHERE target.id = @target_version_id
	)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"target_version_id": targetVersionID,
		"state":             DocumentVersionSignatureStateRequested,
		"now":               time.Now(),
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot move requested document version signatures to the newly published version: %w", err)
	}

	return nil
}

func (pvss *DocumentVersionSignatures) DeleteRequestedByDocumentIDBelowMajor(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	documentID gid.GID,
	major int,
) error {
	q := `
DELETE FROM document_version_signatures
WHERE
	%s
	AND state = @state
	AND document_version_id IN (
		SELECT id
		FROM document_versions
		WHERE document_id = @document_id
			AND major < @major
	)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"document_id": documentID,
		"major":       major,
		"state":       DocumentVersionSignatureStateRequested,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot delete requested document version signatures from previous major versions: %w", err)
	}

	return nil
}

func (pvss *DocumentVersionSignatures) LoadRequestedByDocumentID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentID gid.GID,
) error {
	q := `
SELECT
	document_version_signatures.id,
	document_version_signatures.organization_id,
	document_version_signatures.document_version_id,
	document_version_signatures.state,
	document_version_signatures.signed_by_profile_id,
	document_version_signatures.signed_at,
	document_version_signatures.requested_at,
	document_version_signatures.electronic_signature_id,
	document_version_signatures.created_at,
	document_version_signatures.updated_at
FROM
	document_version_signatures
INNER JOIN document_versions ON document_versions.id = document_version_signatures.document_version_id
WHERE
	%s
	AND document_versions.document_id = @document_id
	AND document_version_signatures.state = @state
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"document_id": documentID,
		"state":       DocumentVersionSignatureStateRequested,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query requested document version signatures: %w", err)
	}

	documentVersionSignatures, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[DocumentVersionSignature])
	if err != nil {
		return fmt.Errorf("cannot collect requested document version signatures: %w", err)
	}

	*pvss = documentVersionSignatures

	return nil
}

func (pvss *DocumentVersionSignatures) DeleteRequestedByDocumentID(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	documentID gid.GID,
) error {
	q := `
DELETE FROM document_version_signatures
WHERE
	%s
	AND state = @state
	AND document_version_id IN (
		SELECT id
		FROM document_versions
		WHERE document_id = @document_id
	)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"document_id": documentID,
		"state":       DocumentVersionSignatureStateRequested,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot delete requested document version signatures: %w", err)
	}

	return nil
}

func (pvss *DocumentVersionSignaturesWithPeople) LoadByDocumentVersionIDWithPeople(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentVersionID gid.GID,
	limit int,
) error {
	q := `
WITH source_version AS (
	SELECT document_id, major FROM document_versions WHERE id = @document_version_id
),
major_versions AS (
	SELECT dv.id FROM document_versions dv
	INNER JOIN source_version sv ON dv.document_id = sv.document_id AND dv.major = sv.major
),
signatures_with_people AS (
	SELECT
		dvs.id,
		dvs.organization_id,
		dvs.tenant_id,
		dvs.document_version_id,
		dvs.state,
		dvs.signed_by_profile_id,
		dvs.signed_at,
		dvs.requested_at,
		dvs.electronic_signature_id,
		dvs.created_at,
		dvs.updated_at,
		p.full_name AS signed_by_full_name
	FROM document_version_signatures dvs
	INNER JOIN major_versions mv ON dvs.document_version_id = mv.id
	INNER JOIN iam_membership_profiles p ON dvs.signed_by_profile_id = p.id
	WHERE
		dvs.state = 'SIGNED'
		OR (
			p.state = 'ACTIVE'
			AND (p.contract_end_date IS NULL OR p.contract_end_date >= CURRENT_DATE)
			AND EXISTS (
				SELECT 1
				FROM iam_memberships m
				WHERE m.identity_id = p.identity_id
					AND m.organization_id = p.organization_id
			)
		)
)
SELECT
	id,
	organization_id,
	document_version_id,
	state,
	signed_by_profile_id,
	signed_at,
	requested_at,
	electronic_signature_id,
	created_at,
	updated_at,
	signed_by_full_name
FROM
	signatures_with_people
WHERE
	%s
ORDER BY
	signed_by_full_name ASC
LIMIT @limit
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"document_version_id": documentVersionID,
		"limit":               limit,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document version signatures with people: %w", err)
	}

	signatures, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[DocumentVersionSignatureWithPeople])
	if err != nil {
		return fmt.Errorf("cannot collect document version signatures with people: %w", err)
	}

	*pvss = signatures

	return nil
}

func (pvs *DocumentVersionSignature) IsSignedByUserEmail(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentVersionID gid.GID,
	userEmail mail.Addr,
) (bool, error) {
	q := `
WITH source_version AS (
	SELECT document_id, major FROM document_versions WHERE id = @document_version_id
),
major_versions AS (
	SELECT dv.id FROM document_versions dv
	INNER JOIN source_version sv ON dv.document_id = sv.document_id AND dv.major = sv.major
),
signed_emails AS (
	SELECT dvs.id, dvs.tenant_id
	FROM document_version_signatures dvs
	INNER JOIN major_versions mv ON dvs.document_version_id = mv.id
	INNER JOIN iam_membership_profiles p ON dvs.signed_by_profile_id = p.id
	INNER JOIN identities i ON p.identity_id = i.id
	WHERE i.email_address = @user_email::CITEXT
		AND dvs.state = 'SIGNED'
)
SELECT EXISTS (
	SELECT 1
	FROM signed_emails
	WHERE %s
) AS signed
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"document_version_id": documentVersionID,
		"user_email":          userEmail,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot query document version signature: %w", err)
	}

	signed, err := pgx.CollectOneRow(rows, pgx.RowTo[bool])
	if err != nil {
		return false, fmt.Errorf("cannot collect signed status: %w", err)
	}

	return signed, nil
}

func (dvs *DocumentVersionSignatures) CountByDocumentVersionID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentVersionID gid.GID,
	filter *DocumentVersionSignatureFilter,
) (int, error) {
	q := `
WITH source_version AS (
	SELECT document_id, major FROM document_versions WHERE id = @document_version_id
),
major_versions AS (
	SELECT dv.id FROM document_versions dv
	INNER JOIN source_version sv ON dv.document_id = sv.document_id AND dv.major = sv.major
)
SELECT
	COUNT(document_version_signatures.id)
FROM
	document_version_signatures
INNER JOIN major_versions mv ON document_version_signatures.document_version_id = mv.id
WHERE
	%s
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.NamedArgs{"document_version_id": documentVersionID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}
