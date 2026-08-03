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
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	MalaysiaPDPAProfile struct {
		OrganizationID                    gid.GID    `db:"organization_id"`
		TotalDataSubjects                 int64      `db:"total_data_subjects"`
		SensitiveDataSubjects             int64      `db:"sensitive_data_subjects"`
		RegularSystematicMonitoring       bool       `db:"regular_systematic_monitoring"`
		DPORequired                       bool       `db:"dpo_required"`
		DPORequirementReasons             []string   `db:"dpo_requirement_reasons"`
		AssessedByProfileID               *gid.GID   `db:"assessed_by_profile_id"`
		AssessedAt                        *time.Time `db:"assessed_at"`
		DPOProfileID                      *gid.GID   `db:"dpo_profile_id"`
		DPOAppointedAt                    *time.Time `db:"dpo_appointed_at"`
		CommissionerNotifiedAt            *time.Time `db:"commissioner_notified_at"`
		CommissionerNotificationReference *string    `db:"commissioner_notification_reference"`
		CreatedAt                         time.Time  `db:"created_at"`
		UpdatedAt                         time.Time  `db:"updated_at"`
	}
)

func (p *MalaysiaPDPAProfile) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) error {
	q := `
SELECT
    organization_id,
    total_data_subjects,
    sensitive_data_subjects,
    regular_systematic_monitoring,
    dpo_required,
    dpo_requirement_reasons,
    assessed_by_profile_id,
    assessed_at,
    dpo_profile_id,
    dpo_appointed_at,
    commissioner_notified_at,
    commissioner_notification_reference,
    created_at,
    updated_at
FROM
    malaysia_pdpa_profiles
WHERE
    %s
    AND organization_id = @organization_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query Malaysia PDPA profile: %w", err)
	}

	profile, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[MalaysiaPDPAProfile])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect Malaysia PDPA profile: %w", err)
	}

	*p = profile

	return nil
}

func (p *MalaysiaPDPAProfile) Upsert(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `
INSERT INTO malaysia_pdpa_profiles (
    organization_id,
    tenant_id,
    total_data_subjects,
    sensitive_data_subjects,
    regular_systematic_monitoring,
    dpo_required,
    dpo_requirement_reasons,
    assessed_by_profile_id,
    assessed_at,
    dpo_profile_id,
    dpo_appointed_at,
    commissioner_notified_at,
    commissioner_notification_reference,
    created_at,
    updated_at
) VALUES (
    @organization_id,
    @tenant_id,
    @total_data_subjects,
    @sensitive_data_subjects,
    @regular_systematic_monitoring,
    @dpo_required,
    @dpo_requirement_reasons,
    @assessed_by_profile_id,
    @assessed_at,
    @dpo_profile_id,
    @dpo_appointed_at,
    @commissioner_notified_at,
    @commissioner_notification_reference,
    @created_at,
    @updated_at
)
ON CONFLICT (organization_id) DO UPDATE
SET
    total_data_subjects = EXCLUDED.total_data_subjects,
    sensitive_data_subjects = EXCLUDED.sensitive_data_subjects,
    regular_systematic_monitoring = EXCLUDED.regular_systematic_monitoring,
    dpo_required = EXCLUDED.dpo_required,
    dpo_requirement_reasons = EXCLUDED.dpo_requirement_reasons,
    assessed_by_profile_id = EXCLUDED.assessed_by_profile_id,
    assessed_at = EXCLUDED.assessed_at,
    dpo_profile_id = EXCLUDED.dpo_profile_id,
    dpo_appointed_at = EXCLUDED.dpo_appointed_at,
    commissioner_notified_at = EXCLUDED.commissioner_notified_at,
    commissioner_notification_reference = EXCLUDED.commissioner_notification_reference,
    updated_at = EXCLUDED.updated_at
`

	args := pgx.StrictNamedArgs{
		"organization_id":                     p.OrganizationID,
		"tenant_id":                           scope.GetTenantID(),
		"total_data_subjects":                 p.TotalDataSubjects,
		"sensitive_data_subjects":             p.SensitiveDataSubjects,
		"regular_systematic_monitoring":       p.RegularSystematicMonitoring,
		"dpo_required":                        p.DPORequired,
		"dpo_requirement_reasons":             p.DPORequirementReasons,
		"assessed_by_profile_id":              p.AssessedByProfileID,
		"assessed_at":                         p.AssessedAt,
		"dpo_profile_id":                      p.DPOProfileID,
		"dpo_appointed_at":                    p.DPOAppointedAt,
		"commissioner_notified_at":            p.CommissionerNotifiedAt,
		"commissioner_notification_reference": p.CommissionerNotificationReference,
		"created_at":                          p.CreatedAt,
		"updated_at":                          p.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot upsert Malaysia PDPA profile: %w", err)
	}

	return nil
}
