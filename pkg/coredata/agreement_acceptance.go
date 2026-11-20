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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

type (
	AgreementAcceptance struct {
		ID                    gid.GID         `db:"id"`
		OrganizationID        gid.GID         `db:"organization_id"`
		SiteID                gid.GID         `db:"site_id"`
		AgreementID           gid.GID         `db:"agreement_id"`
		SubjectID             gid.GID         `db:"subject_id"`
		DocumentVersionID     gid.GID         `db:"document_version_id"`
		ContentSHA256         string          `db:"content_sha256"`
		ElectronicSignatureID gid.GID         `db:"electronic_signature_id"`
		IPAddress             *string         `db:"ip_address"`
		UserAgent             *string         `db:"user_agent"`
		SDKVersion            *string         `db:"sdk_version"`
		Presentation          json.RawMessage `db:"presentation"`
		AcceptedAt            time.Time       `db:"accepted_at"`
		CreatedAt             time.Time       `db:"created_at"`
		UpdatedAt             time.Time       `db:"updated_at"`
	}

	AgreementAcceptances []*AgreementAcceptance
)

func (a *AgreementAcceptance) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM agreement_acceptances WHERE id = ANY(@resource_ids::text[])`

	args := pgx.StrictNamedArgs{
		"resource_ids": resourceIDs,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query agreement acceptance authorization attributes: %w", err)
	}
	defer rows.Close()

	attrsByID := make(policy.AttributesByID)

	for rows.Next() {
		var id, organizationID gid.GID
		if err := rows.Scan(&id, &organizationID); err != nil {
			return nil, fmt.Errorf("cannot scan agreement acceptance authorization attributes: %w", err)
		}

		attrsByID[id] = policy.Attributes{
			"organization_id": organizationID.String(),
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate agreement acceptance authorization attributes: %w", err)
	}

	return attrsByID, nil
}

func (a *AgreementAcceptance) CursorKey(orderBy AgreementAcceptanceOrderField) page.CursorKey {
	switch orderBy {
	case AgreementAcceptanceOrderFieldAcceptedAt:
		return page.NewCursorKey(a.ID, a.AcceptedAt)
	case AgreementAcceptanceOrderFieldCreatedAt:
		return page.NewCursorKey(a.ID, a.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (a *AgreementAcceptance) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO agreement_acceptances (
	id, tenant_id, organization_id, site_id, agreement_id, subject_id,
	document_version_id, content_sha256, electronic_signature_id,
	ip_address, user_agent, sdk_version, presentation,
	accepted_at, created_at, updated_at
) VALUES (
	@id, @tenant_id, @organization_id, @site_id, @agreement_id, @subject_id,
	@document_version_id, @content_sha256, @electronic_signature_id,
	@ip_address, @user_agent, @sdk_version, @presentation,
	@accepted_at, @created_at, @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                      a.ID,
		"tenant_id":               scope.GetTenantID(),
		"organization_id":         a.OrganizationID,
		"site_id":                 a.SiteID,
		"agreement_id":            a.AgreementID,
		"subject_id":              a.SubjectID,
		"document_version_id":     a.DocumentVersionID,
		"content_sha256":          a.ContentSHA256,
		"electronic_signature_id": a.ElectronicSignatureID,
		"ip_address":              a.IPAddress,
		"user_agent":              a.UserAgent,
		"sdk_version":             a.SDKVersion,
		"presentation":            a.Presentation,
		"accepted_at":             a.AcceptedAt,
		"created_at":              a.CreatedAt,
		"updated_at":              a.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert agreement acceptance: %w", err)
	}

	return nil
}

func (a *AgreementAcceptance) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
	id, organization_id, site_id, agreement_id, subject_id,
	document_version_id, content_sha256, electronic_signature_id,
	ip_address, user_agent, sdk_version, presentation,
	accepted_at, created_at, updated_at
FROM agreement_acceptances
WHERE %s AND id = @id
LIMIT 1
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query agreement acceptance: %w", err)
	}

	acc, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AgreementAcceptance])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect agreement acceptance: %w", err)
	}

	*a = acc

	return nil
}

// LoadLatestBySubjectAndAgreement returns the most recent acceptance for a
// (subject, agreement) pair, or ErrResourceNotFound if none exists.
func (a *AgreementAcceptance) LoadLatestBySubjectAndAgreement(
	ctx context.Context,
	conn pg.Querier,
	subjectID gid.GID,
	agreementID gid.GID,
) error {
	q := `
SELECT
	id, organization_id, site_id, agreement_id, subject_id,
	document_version_id, content_sha256, electronic_signature_id,
	ip_address, user_agent, sdk_version, presentation,
	accepted_at, created_at, updated_at
FROM agreement_acceptances
WHERE subject_id = @subject_id AND agreement_id = @agreement_id
ORDER BY accepted_at DESC
LIMIT 1
`

	args := pgx.StrictNamedArgs{
		"subject_id":   subjectID,
		"agreement_id": agreementID,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query latest agreement acceptance: %w", err)
	}

	acc, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AgreementAcceptance])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect latest agreement acceptance: %w", err)
	}

	*a = acc

	return nil
}

// LoadByElectronicSignatureID loads the acceptance that owns a given signature.
// Used by the sealing worker to retrieve content_sha256 for CLICKWRAP without
// re-downloading the PDF.
func (a *AgreementAcceptance) LoadByElectronicSignatureID(
	ctx context.Context,
	conn pg.Querier,
	signatureID gid.GID,
) error {
	q := `
SELECT
	id, organization_id, site_id, agreement_id, subject_id,
	document_version_id, content_sha256, electronic_signature_id,
	ip_address, user_agent, sdk_version, presentation,
	accepted_at, created_at, updated_at
FROM agreement_acceptances
WHERE electronic_signature_id = @electronic_signature_id
LIMIT 1
`

	args := pgx.StrictNamedArgs{"electronic_signature_id": signatureID}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query acceptance by signature id: %w", err)
	}

	acc, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AgreementAcceptance])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect acceptance by signature id: %w", err)
	}

	*a = acc

	return nil
}

func (aa *AgreementAcceptances) LoadBySubjectID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	subjectID gid.GID,
	cursor *page.Cursor[AgreementAcceptanceOrderField],
) error {
	q := `
SELECT
	id, organization_id, site_id, agreement_id, subject_id,
	document_version_id, content_sha256, electronic_signature_id,
	ip_address, user_agent, sdk_version, presentation,
	accepted_at, created_at, updated_at
FROM agreement_acceptances
WHERE
	%s
	AND subject_id = @subject_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"subject_id": subjectID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query agreement acceptances: %w", err)
	}

	acceptances, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AgreementAcceptance])
	if err != nil {
		return fmt.Errorf("cannot collect agreement acceptances: %w", err)
	}

	*aa = acceptances

	return nil
}
