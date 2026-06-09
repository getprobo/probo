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

package vetting

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

const (
	vettingRiskAssessmentValidity = 365 * 24 * time.Hour
	maxVettingNotesGaps           = 5
)

// PersistAssessmentResult writes extracted assessment metadata onto the parent
// third party, links any sub-processors, and stores the risk assessment in one
// short transaction after the long assess phase completes. The assess run
// itself does not touch the database.
func PersistAssessmentResult(
	ctx context.Context,
	pc *PersistenceContext,
	result Result,
) error {
	scope := coredata.NewScopeFromObjectID(pc.ThirdPartyID)

	return pc.PG.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			thirdParty := &coredata.ThirdParty{}

			if err := thirdParty.LoadByID(ctx, conn, scope, pc.ThirdPartyID); err != nil {
				return fmt.Errorf("cannot load third party: %w", err)
			}

			applySaveParams(thirdParty, pc.WebsiteURL, saveParamsFromInfo(result.Info))
			thirdParty.UpdatedAt = time.Now()

			if err := thirdParty.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update third party: %w", err)
			}

			for _, sub := range result.Info.Subprocessors {
				if sub.Name == "" {
					continue
				}

				if err := linkSubThirdParty(
					ctx,
					conn,
					scope,
					pc,
					linkSubThirdPartyParams{
						Name:    sub.Name,
						Country: sub.Country,
						Purpose: sub.Purpose,
					},
				); err != nil {
					return fmt.Errorf("cannot link sub third party %q: %w", sub.Name, err)
				}
			}

			if err := persistVettingRiskAssessment(
				ctx,
				conn,
				scope,
				pc,
				thirdParty,
				result,
			); err != nil {
				return fmt.Errorf("cannot persist vetting risk assessment: %w", err)
			}

			return nil
		},
	)
}

func persistVettingRiskAssessment(
	ctx context.Context,
	conn pg.Tx,
	scope coredata.Scoper,
	pc *PersistenceContext,
	thirdParty *coredata.ThirdParty,
	result Result,
) error {
	if err := thirdParty.ExpireNonExpiredRiskAssessments(ctx, conn, scope); err != nil {
		return fmt.Errorf("cannot expire existing risk assessments: %w", err)
	}

	now := time.Now()
	notes := buildRiskAssessmentNotes(result.Info)

	assessment := &coredata.ThirdPartyRiskAssessment{
		ID:              gid.New(scope.GetTenantID(), coredata.ThirdPartyRiskAssessmentEntityType),
		OrganizationID:  pc.OrganizationID,
		ThirdPartyID:    pc.ThirdPartyID,
		ExpiresAt:       now.Add(vettingRiskAssessmentValidity),
		DataSensitivity: mapVettingDataSensitivity(result.Info),
		BusinessImpact:  mapVettingBusinessImpact(result.Info),
		Notes:           &notes,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := assessment.Insert(ctx, conn, scope); err != nil {
		return fmt.Errorf("cannot insert risk assessment: %w", err)
	}

	return nil
}

func buildRiskAssessmentNotes(info ThirdPartyInfo) string {
	var b strings.Builder

	b.WriteString("Automated vetting\n\n")

	switch {
	case info.OverallRiskRating != "" && info.OverallRiskScore > 0:
		fmt.Fprintf(
			&b,
			"Overall risk: %d/100 (%s)\n",
			info.OverallRiskScore,
			info.OverallRiskRating,
		)
	case info.OverallRiskScore > 0:
		fmt.Fprintf(&b, "Overall risk: %d/100\n", info.OverallRiskScore)
	case info.OverallRiskRating != "":
		fmt.Fprintf(&b, "Overall risk: %s\n", info.OverallRiskRating)
	}

	if info.Recommendation != "" {
		fmt.Fprintf(&b, "Recommendation: %s\n", formatVettingRecommendation(info.Recommendation))
	}

	var scoreParts []string

	if info.SecurityRiskScore > 0 {
		scoreParts = append(scoreParts, fmt.Sprintf("Security %d/100", info.SecurityRiskScore))
	}

	if info.PrivacyRiskScore > 0 {
		scoreParts = append(scoreParts, fmt.Sprintf("Privacy %d/100", info.PrivacyRiskScore))
	}

	if info.InvolvesAI || info.AIRiskScore > 0 {
		scoreParts = append(scoreParts, fmt.Sprintf("AI %d/100", info.AIRiskScore))
	}

	if len(scoreParts) > 0 {
		b.WriteByte('\n')
		b.WriteString(strings.Join(scoreParts, " · "))
		b.WriteByte('\n')
	}

	if len(info.InformationGaps) > 0 {
		b.WriteString("\nGaps\n")

		gaps := info.InformationGaps
		if len(gaps) > maxVettingNotesGaps {
			gaps = gaps[:maxVettingNotesGaps]
		}

		for _, gap := range gaps {
			fmt.Fprintf(&b, "· %s\n", strings.TrimSpace(gap))
		}
	}

	return strings.TrimSpace(b.String())
}

func formatVettingRecommendation(recommendation string) string {
	switch strings.ToUpper(strings.TrimSpace(recommendation)) {
	case "APPROVE":
		return "Approve"
	case "APPROVE_WITH_CONDITIONS":
		return "Approve with conditions"
	case "ESCALATE":
		return "Escalate"
	case "REJECT":
		return "Reject"
	default:
		return recommendation
	}
}

func mapVettingDataSensitivity(info ThirdPartyInfo) coredata.DataSensitivity {
	if !info.ProcessesPII && info.PrivacyRiskScore == 0 {
		return coredata.DataSensitivityNone
	}

	score := info.PrivacyRiskScore
	if score == 0 {
		score = overallScoreFromRating(info.OverallRiskRating)
	}

	return scoreToDataSensitivity(score)
}

func mapVettingBusinessImpact(info ThirdPartyInfo) coredata.BusinessImpact {
	score := info.OverallRiskScore
	if score == 0 {
		score = info.SecurityRiskScore
	}

	if score == 0 {
		score = overallScoreFromRating(info.OverallRiskRating)
	}

	return scoreToBusinessImpact(score)
}

func overallScoreFromRating(rating string) int {
	switch strings.ToLower(strings.TrimSpace(rating)) {
	case "low":
		return 25
	case "medium":
		return 50
	case "high":
		return 75
	default:
		return 0
	}
}

func scoreToDataSensitivity(score int) coredata.DataSensitivity {
	switch {
	case score <= 0:
		return coredata.DataSensitivityNone
	case score <= 25:
		return coredata.DataSensitivityLow
	case score <= 50:
		return coredata.DataSensitivityMedium
	case score <= 75:
		return coredata.DataSensitivityHigh
	default:
		return coredata.DataSensitivityCritical
	}
}

func scoreToBusinessImpact(score int) coredata.BusinessImpact {
	switch {
	case score <= 25:
		return coredata.BusinessImpactLow
	case score <= 50:
		return coredata.BusinessImpactMedium
	case score <= 75:
		return coredata.BusinessImpactHigh
	default:
		return coredata.BusinessImpactCritical
	}
}

func saveParamsFromInfo(info ThirdPartyInfo) saveThirdPartyInfoParams {
	return saveThirdPartyInfoParams{
		saveThirdPartyInfoToolParams: saveThirdPartyInfoToolParams{
			Name:                          info.Name,
			Description:                   info.Description,
			Category:                      info.Category,
			HeadquarterAddress:            info.HeadquarterAddress,
			LegalName:                     info.LegalName,
			PrivacyPolicyURL:              info.PrivacyPolicyURL,
			ServiceLevelAgreementURL:      info.ServiceLevelAgreementURL,
			DataProcessingAgreementURL:    info.DataProcessingAgreementURL,
			BusinessAssociateAgreementURL: info.BusinessAssociateAgreementURL,
			SubprocessorsListURL:          info.SubprocessorsListURL,
			SecurityPageURL:               info.SecurityPageURL,
			TrustPageURL:                  info.TrustPageURL,
			TermsOfServiceURL:             info.TermsOfServiceURL,
			StatusPageURL:                 info.StatusPageURL,
			Certifications:                info.Certifications,
		},
		Countries: countriesFromInfo(info),
	}
}

func applySaveParams(
	thirdParty *coredata.ThirdParty,
	websiteURL string,
	p saveThirdPartyInfoParams,
) {
	if p.Name != "" {
		thirdParty.Name = p.Name
	}

	thirdParty.WebsiteURL = &websiteURL

	if p.Category != "" {
		if category, err := parseThirdPartyCategory(p.Category); err == nil {
			thirdParty.Category = category
		}
	}

	if p.Description != "" {
		thirdParty.Description = &p.Description
	}

	if p.HeadquarterAddress != "" {
		thirdParty.HeadquarterAddress = &p.HeadquarterAddress
	}

	if p.LegalName != "" {
		thirdParty.LegalName = &p.LegalName
	}

	if p.PrivacyPolicyURL != "" {
		thirdParty.PrivacyPolicyURL = &p.PrivacyPolicyURL
	}

	if p.ServiceLevelAgreementURL != "" {
		thirdParty.ServiceLevelAgreementURL = &p.ServiceLevelAgreementURL
	}

	if p.DataProcessingAgreementURL != "" {
		thirdParty.DataProcessingAgreementURL = &p.DataProcessingAgreementURL
	}

	if p.BusinessAssociateAgreementURL != "" {
		thirdParty.BusinessAssociateAgreementURL = &p.BusinessAssociateAgreementURL
	}

	if p.SubprocessorsListURL != "" {
		thirdParty.SubprocessorsListURL = &p.SubprocessorsListURL
	}

	if p.SecurityPageURL != "" {
		thirdParty.SecurityPageURL = &p.SecurityPageURL
	}

	if p.TrustPageURL != "" {
		thirdParty.TrustPageURL = &p.TrustPageURL
	}

	if p.TermsOfServiceURL != "" {
		thirdParty.TermsOfServiceURL = &p.TermsOfServiceURL
	}

	if p.StatusPageURL != "" {
		thirdParty.StatusPageURL = &p.StatusPageURL
	}

	if len(p.Certifications) > 0 {
		thirdParty.Certifications = p.Certifications
	}

	if len(p.Countries) > 0 {
		thirdParty.Countries = p.Countries
	}
}

func linkSubThirdParty(
	ctx context.Context,
	conn pg.Tx,
	scope coredata.Scoper,
	pc *PersistenceContext,
	p linkSubThirdPartyParams,
) error {
	if p.Name == "" {
		return nil
	}

	child := &coredata.ThirdParty{}

	err := child.LoadByNameAndOrganizationID(ctx, conn, scope, p.Name, pc.OrganizationID)
	if err != nil {
		if !errors.Is(err, coredata.ErrResourceNotFound) {
			return fmt.Errorf("cannot find child third party %q: %w", p.Name, err)
		}

		now := time.Now()
		child = &coredata.ThirdParty{
			ID:             gid.New(scope.GetTenantID(), coredata.ThirdPartyEntityType),
			OrganizationID: pc.OrganizationID,
			Name:           p.Name,
			Category:       coredata.ThirdPartyCategoryOther,
			FirstLevel:     false,
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		if p.Description != "" {
			child.Description = &p.Description
		}

		if p.Category != "" {
			if category, err := parseThirdPartyCategory(p.Category); err == nil {
				child.Category = category
			}
		}

		if p.WebsiteURL != "" {
			child.WebsiteURL = &p.WebsiteURL
		}

		if countries := parseOptionalCountryCodes(p.Country); len(countries) > 0 {
			child.Countries = countries
		}

		if err := child.Insert(ctx, conn, scope); err != nil {
			return fmt.Errorf("cannot create child third party %q: %w", p.Name, err)
		}
	} else if countries := parseOptionalCountryCodes(p.Country); len(countries) > 0 && len(child.Countries) == 0 {
		child.Countries = countries
		child.UpdatedAt = time.Now()

		if err := child.Update(ctx, conn, scope); err != nil {
			return fmt.Errorf("cannot update child third party %q countries: %w", p.Name, err)
		}
	}

	if child.ID == pc.ThirdPartyID {
		return nil
	}

	relation := &coredata.ThirdPartyThirdParty{
		ParentThirdPartyID: pc.ThirdPartyID,
		ChildThirdPartyID:  child.ID,
		CreatedAt:          time.Now(),
	}

	if p.Purpose != "" {
		relation.Purpose = &p.Purpose
	}

	if err := relation.Insert(ctx, conn, scope); err != nil {
		return fmt.Errorf("cannot insert third party relation: %w", err)
	}

	return nil
}
