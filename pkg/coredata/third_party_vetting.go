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
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

func (v *ThirdParty) LoadNextPendingVettingForUpdateSkipLocked(
	ctx context.Context,
	tx pg.Tx,
) error {
	q := `
SELECT
    id,
    organization_id,
    parent_third_party_id,
    common_third_party_id,
    name,
    description,
    category,
    headquarter_address,
    legal_name,
    website_url,
    privacy_policy_url,
    service_level_agreement_url,
    data_processing_agreement_url,
    business_associate_agreement_url,
    subprocessors_list_url,
    certifications,
    countries,
    business_owner_profile_id,
    security_owner_profile_id,
    status_page_url,
    terms_of_service_url,
    security_page_url,
    trust_page_url,
    show_on_trust_center,
    level,
    vetting_status,
    vetting_website_url,
    vetting_procedure,
    vetting_processing_started_at,
    vetting_error_message,
    created_at,
    updated_at
FROM
    third_parties
WHERE
    vetting_status = @vetting_status
    AND vetting_website_url IS NOT NULL
ORDER BY
    created_at ASC
LIMIT 1
FOR UPDATE SKIP LOCKED;
`

	rows, err := tx.Query(
		ctx,
		q,
		pgx.StrictNamedArgs{"vetting_status": ThirdPartyVettingStatusPending},
	)
	if err != nil {
		return fmt.Errorf("cannot query third party vetting queue: %w", err)
	}

	thirdParty, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ThirdParty])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect third party: %w", err)
	}

	*v = thirdParty

	return nil
}

func ResetStaleVettingProcessing(
	ctx context.Context,
	conn pg.Querier,
	staleAfter time.Duration,
) error {
	q := `
UPDATE third_parties
SET
    vetting_status = @pending_status,
    vetting_processing_started_at = NULL,
    updated_at = @now
WHERE
    vetting_status = @processing_status
    AND vetting_processing_started_at < @stale_before;
`

	_, err := conn.Exec(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"pending_status":    ThirdPartyVettingStatusPending,
			"processing_status": ThirdPartyVettingStatusProcessing,
			"now":               time.Now(),
			"stale_before":      time.Now().Add(-staleAfter),
		},
	)

	return err
}
