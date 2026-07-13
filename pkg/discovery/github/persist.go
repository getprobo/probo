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
	"fmt"
	"net/url"
	"strings"
	"time"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type (
	persistInput struct {
		plan           *MeasurePlan
		factSheet      *FactSheet
		thirdPartyID   gid.GID
		organizationID gid.GID
		agentRunID     gid.GID
	}

	persistStats struct {
		upserted int
		summary  map[string]int
	}
)

func applyMeasurePlan(
	ctx context.Context,
	pgClient *pg.Client,
	scope coredata.Scoper,
	input persistInput,
) (*persistStats, error) {
	if input.plan == nil {
		return nil, fmt.Errorf("cannot apply measure plan: plan is required")
	}

	factsByName := map[string]Fact{}

	for _, fact := range input.factSheet.Facts {
		factsByName[fact.Name] = fact
	}

	stats := &persistStats{summary: map[string]int{}}

	err := pgClient.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			now := time.Now()

			for _, update := range input.plan.Updates {
				if err := persistMeasureUpdate(
					ctx,
					tx,
					scope,
					input,
					update,
					factsByName,
					now,
				); err != nil {
					return err
				}

				stats.upserted++
				stats.summary[string(update.State)]++
			}

			for _, create := range input.plan.Creates {
				if err := persistMeasureCreate(
					ctx,
					tx,
					scope,
					input,
					create,
					factsByName,
					now,
				); err != nil {
					return err
				}

				stats.upserted++
				stats.summary[string(create.State)]++
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot apply measure plan transaction: %w", err)
	}

	return stats, nil
}

func persistMeasureUpdate(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	input persistInput,
	update MeasurePlanUpdate,
	factsByName map[string]Fact,
	now time.Time,
) error {
	measure := &coredata.Measure{}

	if err := measure.LoadByID(ctx, tx, scope, update.MeasureID); err != nil {
		return fmt.Errorf("cannot load measure %q: %w", update.MeasureID, err)
	}

	measure.State = update.State
	measure.UpdatedAt = now

	if err := measure.Update(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot update measure %q: %w", update.MeasureID, err)
	}

	if err := linkMeasureThirdParty(ctx, tx, scope, measure.ID, input.thirdPartyID, now); err != nil {
		return err
	}

	return insertDiscoveryEvidence(
		ctx,
		tx,
		scope,
		measure.OrganizationID,
		measure.ID,
		input.agentRunID,
		update.EvidenceSummary,
		measure.Name,
		factsByName,
		now,
	)
}

func persistMeasureCreate(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	input persistInput,
	create MeasurePlanCreate,
	factsByName map[string]Fact,
	now time.Time,
) error {
	referenceID, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("cannot generate measure reference id: %w", err)
	}

	measure := &coredata.Measure{
		ID:             gid.New(scope.GetTenantID(), coredata.MeasureEntityType),
		OrganizationID: input.organizationID,
		Name:           create.Name,
		Description:    stringPtr(create.Description),
		Category:       create.Category,
		ReferenceID:    "github-discovery-" + referenceID.String(),
		State:          create.State,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := measure.Insert(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot insert measure: %w", err)
	}

	if err := linkMeasureThirdParty(ctx, tx, scope, measure.ID, input.thirdPartyID, now); err != nil {
		return err
	}

	return insertDiscoveryEvidence(
		ctx,
		tx,
		scope,
		measure.OrganizationID,
		measure.ID,
		input.agentRunID,
		create.EvidenceSummary,
		create.Name,
		factsByName,
		now,
	)
}

func linkMeasureThirdParty(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	measureID gid.GID,
	thirdPartyID gid.GID,
	now time.Time,
) error {
	mapping := coredata.MeasureThirdParty{
		MeasureID:    measureID,
		ThirdPartyID: thirdPartyID,
		CreatedAt:    now,
	}

	if err := mapping.Upsert(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot link measure to github third party: %w", err)
	}

	return nil
}

func insertDiscoveryEvidence(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	organizationID gid.GID,
	measureID gid.GID,
	agentRunID gid.GID,
	summary string,
	measureName string,
	factsByName map[string]Fact,
	now time.Time,
) error {
	referenceID, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("cannot generate evidence reference id: %w", err)
	}

	evidenceURL, err := discoveryEvidenceURL(measureName, factsByName)
	if err != nil {
		return fmt.Errorf("cannot build discovery evidence URL: %w", err)
	}

	description := strings.TrimSpace(summary)
	if description == "" {
		description = "GitHub discovery evidence"
	}

	evidence := &coredata.Evidence{
		ID:                gid.New(scope.GetTenantID(), coredata.EvidenceEntityType),
		OrganizationID:    organizationID,
		MeasureID:         measureID,
		State:             coredata.EvidenceStateFulfilled,
		ReferenceID:       fmt.Sprintf("github-discovery-%s-%s", agentRunID.String(), referenceID.String()),
		Type:              coredata.EvidenceTypeLink,
		URL:               evidenceURL,
		Description:       &description,
		DescriptionStatus: coredata.EvidenceDescriptionStatusCompleted,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := evidence.Insert(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot insert discovery evidence: %w", err)
	}

	return nil
}

func discoveryEvidenceURL(measureName string, factsByName map[string]Fact) (string, error) {
	const fallback = "https://github.com"

	fact, ok := factsByName[measureName]
	if !ok || fact.APIRef == "" {
		return fallback, nil
	}

	rest := strings.TrimSpace(strings.TrimPrefix(fact.APIRef, "GET "))
	if rest == "" {
		return fallback, nil
	}

	u, err := url.Parse("https://api.github.com")
	if err != nil {
		return "", fmt.Errorf("cannot parse github api base URL: %w", err)
	}

	if path, query, ok := strings.Cut(rest, "?"); ok {
		u.Path = path

		q, err := url.ParseQuery(query)
		if err != nil {
			return "", fmt.Errorf("cannot parse github api evidence query: %w", err)
		}

		u.RawQuery = q.Encode()
	} else {
		u.Path = rest
	}

	return u.String(), nil
}

func stringPtr(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	return new(value)
}
