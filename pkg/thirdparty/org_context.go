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

package thirdparty

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/x/ref"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

// maxOrgContextListItems caps how many items each org-context list
// renders, keeping the prompt bounded for large organizations.
const maxOrgContextListItems = 50

// buildOrganizationContext assembles the org-specific context block fed to
// the assessment orchestrator: the org's frameworks, the vendor's prior
// risk rating, and the data/assets/processing routed through the vendor.
// Best-effort: a failed load is logged and omitted. Returns an empty
// string when nothing org-specific is known.
func (h *vettingHandler) buildOrganizationContext(
	ctx context.Context,
	thirdParty *coredata.ThirdParty,
) string {
	scope := coredata.NewScopeFromObjectID(thirdParty.ID)

	var (
		frameworks   coredata.Frameworks
		linkedData   coredata.Data
		linkedAssets coredata.Assets
		activities   coredata.ProcessingActivities
		risk         coredata.ThirdPartyRiskAssessment
		hasRisk      bool
	)

	err := h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			frameworks = h.loadOrgFrameworks(ctx, conn, scope, thirdParty.OrganizationID)
			linkedData = h.loadLinkedData(ctx, conn, scope, thirdParty.ID)
			linkedAssets = h.loadLinkedAssets(ctx, conn, scope, thirdParty.ID)
			activities = h.loadLinkedProcessingActivities(ctx, conn, scope, thirdParty.ID)

			if err := risk.LoadLatestByThirdPartyID(ctx, conn, scope, thirdParty.ID); err != nil {
				if !errors.Is(err, coredata.ErrResourceNotFound) {
					h.logger.WarnCtx(
						ctx,
						"cannot load third party risk assessment for org context",
						log.Error(err),
						log.String("third_party_id", thirdParty.ID.String()),
					)
				}
			} else if risk.ExpiresAt.After(time.Now()) {
				// Skip expired ratings: they would read as current.
				hasRisk = true
			}

			return nil
		},
	)
	if err != nil {
		h.logger.WarnCtx(
			ctx,
			"cannot open connection for organization context",
			log.Error(err),
			log.String("third_party_id", thirdParty.ID.String()),
		)

		return ""
	}

	return renderOrganizationContext(
		thirdParty,
		frameworks,
		risk,
		hasRisk,
		linkedData,
		linkedAssets,
		activities,
	)
}

func (h *vettingHandler) loadOrgFrameworks(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	organizationID gid.GID,
) coredata.Frameworks {
	frameworks, err := page.LoadAll(
		ctx,
		page.OrderBy[coredata.FrameworkOrderField]{
			Field:     coredata.FrameworkOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
		func(ctx context.Context, cursor *page.Cursor[coredata.FrameworkOrderField]) ([]*coredata.Framework, error) {
			var batch coredata.Frameworks
			if err := batch.LoadByOrganizationID(ctx, conn, scope, organizationID, cursor); err != nil {
				return nil, err
			}

			return batch, nil
		},
	)
	if err != nil {
		h.logger.WarnCtx(ctx, "cannot load organization frameworks for org context", log.Error(err))

		return nil
	}

	return frameworks
}

func (h *vettingHandler) loadLinkedData(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	thirdPartyID gid.GID,
) coredata.Data {
	data, err := page.LoadAll(
		ctx,
		page.OrderBy[coredata.DatumOrderField]{
			Field:     coredata.DatumOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
		func(ctx context.Context, cursor *page.Cursor[coredata.DatumOrderField]) ([]*coredata.Datum, error) {
			var batch coredata.Data
			if err := batch.LoadByThirdPartyID(ctx, conn, scope, thirdPartyID, cursor); err != nil {
				return nil, err
			}

			return batch, nil
		},
	)
	if err != nil {
		h.logger.WarnCtx(ctx, "cannot load linked data for org context", log.Error(err))

		return nil
	}

	return data
}

func (h *vettingHandler) loadLinkedAssets(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	thirdPartyID gid.GID,
) coredata.Assets {
	assets, err := page.LoadAll(
		ctx,
		page.OrderBy[coredata.AssetOrderField]{
			Field:     coredata.AssetOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
		func(ctx context.Context, cursor *page.Cursor[coredata.AssetOrderField]) ([]*coredata.Asset, error) {
			var batch coredata.Assets
			if err := batch.LoadByThirdPartyID(ctx, conn, scope, thirdPartyID, cursor); err != nil {
				return nil, err
			}

			return batch, nil
		},
	)
	if err != nil {
		h.logger.WarnCtx(ctx, "cannot load linked assets for org context", log.Error(err))

		return nil
	}

	return assets
}

func (h *vettingHandler) loadLinkedProcessingActivities(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	thirdPartyID gid.GID,
) coredata.ProcessingActivities {
	activities, err := page.LoadAll(
		ctx,
		page.OrderBy[coredata.ProcessingActivityOrderField]{
			Field:     coredata.ProcessingActivityOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
		func(ctx context.Context, cursor *page.Cursor[coredata.ProcessingActivityOrderField]) ([]*coredata.ProcessingActivity, error) {
			var batch coredata.ProcessingActivities
			if err := batch.LoadByThirdPartyID(ctx, conn, scope, thirdPartyID, cursor); err != nil {
				return nil, err
			}

			return batch, nil
		},
	)
	if err != nil {
		h.logger.WarnCtx(ctx, "cannot load linked processing activities for org context", log.Error(err))

		return nil
	}

	return activities
}

// renderOrganizationContext formats the assembled data into the
// <organization_context> block. Empty sections are omitted; an empty
// string is returned when nothing org-specific is known.
func renderOrganizationContext(
	thirdParty *coredata.ThirdParty,
	frameworks coredata.Frameworks,
	risk coredata.ThirdPartyRiskAssessment,
	hasRisk bool,
	linkedData coredata.Data,
	linkedAssets coredata.Assets,
	activities coredata.ProcessingActivities,
) string {
	countries := countryStrings(thirdParty.Countries)

	hasContent := len(frameworks) > 0 ||
		hasRisk ||
		len(linkedData) > 0 ||
		len(linkedAssets) > 0 ||
		len(activities) > 0 ||
		len(countries) > 0

	if !hasContent {
		return ""
	}

	var b strings.Builder

	b.WriteString("<organization_context>\n")
	b.WriteString("Assess this vendor for the specific organization described below. Weight risk by how this organization uses the vendor and the obligations it must meet. Treat requirements implied by the organization's frameworks and the sensitivity of the shared data as acceptance criteria: unmet hard requirements should lower the recommendation (escalate or reject) and be called out as conditions.\n")

	if len(frameworks) > 0 {
		b.WriteString("\n<frameworks>\n")
		b.WriteString("Compliance frameworks the organization maintains; assess the vendor against their relevant requirements:\n")
		shown := frameworks[:min(len(frameworks), maxOrgContextListItems)]
		for _, f := range shown {
			fmt.Fprintf(&b, "- %s\n", strings.TrimSpace(f.Name))
		}
		writeMoreItems(&b, len(frameworks), len(shown))
		b.WriteString("</frameworks>\n")
	}

	relationship := make([]string, 0, 4)
	if category := strings.TrimSpace(thirdParty.Category.String()); category != "" {
		relationship = append(relationship, "- vendor_category: "+category)
	}
	if len(countries) > 0 {
		relationship = append(relationship, "- vendor_operating_countries: "+strings.Join(countries, ", "))
	}
	if hasRisk {
		relationship = append(relationship, fmt.Sprintf("- data_sensitivity_rating: %s (from prior risk assessment)", risk.DataSensitivity.String()))
		relationship = append(relationship, fmt.Sprintf("- business_impact_rating: %s (from prior risk assessment)", risk.BusinessImpact.String()))
	}
	if len(relationship) > 0 {
		b.WriteString("\n<relationship>\n")
		for _, line := range relationship {
			b.WriteString(line + "\n")
		}
		b.WriteString("</relationship>\n")
	}

	if len(linkedData) > 0 {
		b.WriteString("\n<shared_data>\n")
		b.WriteString("Data the organization shares with or exposes to this vendor:\n")
		shown := linkedData[:min(len(linkedData), maxOrgContextListItems)]
		for _, d := range shown {
			fmt.Fprintf(&b, "- %s (classification: %s)\n", strings.TrimSpace(d.Name), d.DataClassification.String())
		}
		writeMoreItems(&b, len(linkedData), len(shown))
		b.WriteString("</shared_data>\n")
	}

	if len(linkedAssets) > 0 {
		b.WriteString("\n<linked_assets>\n")
		b.WriteString("Organization assets that involve this vendor:\n")
		shown := linkedAssets[:min(len(linkedAssets), maxOrgContextListItems)]
		for _, a := range shown {
			dataTypes := strings.TrimSpace(a.DataTypesStored)
			if dataTypes == "" {
				fmt.Fprintf(&b, "- %s\n", strings.TrimSpace(a.Name))
				continue
			}

			fmt.Fprintf(&b, "- %s (data stored: %s)\n", strings.TrimSpace(a.Name), dataTypes)
		}
		writeMoreItems(&b, len(linkedAssets), len(shown))
		b.WriteString("</linked_assets>\n")
	}

	if len(activities) > 0 {
		b.WriteString("\n<processing_activities>\n")
		b.WriteString("Processing activities that route data through this vendor:\n")
		shown := activities[:min(len(activities), maxOrgContextListItems)]
		for _, pa := range shown {
			b.WriteString(renderProcessingActivity(pa))
		}
		writeMoreItems(&b, len(activities), len(shown))
		b.WriteString("</processing_activities>\n")
	}

	b.WriteString("</organization_context>")

	return b.String()
}

// writeMoreItems appends a truncation marker when a list was capped.
func writeMoreItems(b *strings.Builder, total, shown int) {
	if total > shown {
		fmt.Fprintf(b, "- ... and %d more\n", total-shown)
	}
}

func renderProcessingActivity(pa *coredata.ProcessingActivity) string {
	parts := make([]string, 0, 6)

	if purpose := ref.UnrefOrZero(pa.Purpose); strings.TrimSpace(purpose) != "" {
		parts = append(parts, "purpose: "+strings.TrimSpace(purpose))
	}

	if category := ref.UnrefOrZero(pa.PersonalDataCategory); strings.TrimSpace(category) != "" {
		parts = append(parts, "personal data: "+strings.TrimSpace(category))
	}

	if location := ref.UnrefOrZero(pa.Location); strings.TrimSpace(location) != "" {
		parts = append(parts, "location: "+strings.TrimSpace(location))
	}

	if pa.InternationalTransfers {
		parts = append(parts, "international transfers: yes")

		if pa.TransferSafeguard != nil {
			parts = append(parts, "safeguard: "+pa.TransferSafeguard.String())
		}
	}

	if basis := strings.TrimSpace(pa.LawfulBasis.String()); basis != "" {
		parts = append(parts, "lawful basis: "+basis)
	}

	if retention := ref.UnrefOrZero(pa.RetentionPeriod); strings.TrimSpace(retention) != "" {
		parts = append(parts, "retention: "+strings.TrimSpace(retention))
	}

	name := strings.TrimSpace(pa.Name)
	if name == "" {
		return fmt.Sprintf("- %s\n", strings.Join(parts, " | "))
	}

	return fmt.Sprintf("- %s — %s\n", name, strings.Join(parts, " | "))
}

func countryStrings(countries coredata.CountryCodes) []string {
	out := make([]string, 0, len(countries))
	for _, c := range countries {
		code := strings.TrimSpace(c.String())
		if code == "" {
			continue
		}

		out = append(out, code)
	}

	return out
}
