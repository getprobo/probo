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

package vetting

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

// nameSuffixPattern matches a trailing " (path)" suffix on a stored third party
// name. It mirrors the console UI regex so backend and frontend agree on how a
// hierarchy-qualified name is split back into its bare base.
var nameSuffixPattern = regexp.MustCompile(`\s*\([^)]*\)\s*$`)

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

			// Sub third parties store hierarchy-qualified names ("aws (Probo)").
			// Load the ancestor chain so the vetted third party and any
			// discovered sub-processors are named consistently with the console.
			ancestorBaseNames, err := loadAncestorBaseNames(ctx, conn, scope, thirdParty.ID)
			if err != nil {
				return err
			}

			applySaveParams(thirdParty, pc.WebsiteURL, saveParamsFromInfo(result.Info), ancestorBaseNames)
			thirdParty.UpdatedAt = time.Now()

			if err := thirdParty.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update third party: %w", err)
			}

			// The suffix for children of this third party is the ancestor path
			// plus the third party itself (computed after applySaveParams so a
			// freshly canonicalized name is reflected).
			childNamePath := make([]string, 0, len(ancestorBaseNames)+1)
			childNamePath = append(childNamePath, ancestorBaseNames...)
			childNamePath = append(childNamePath, baseThirdPartyName(thirdParty.Name))

			for _, sub := range result.Info.Subprocessors {
				if sub.Name == "" {
					continue
				}

				if err := linkSubThirdParty(
					ctx,
					conn,
					scope,
					pc,
					thirdParty.Level,
					childNamePath,
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
	sections := []string{"Automated vetting"}

	appendSection := func(section string) {
		if strings.TrimSpace(section) != "" {
			sections = append(sections, section)
		}
	}

	appendSection(vettingOverviewSection(info))
	appendSection(vettingPillarSection(info))
	appendSection(vettingClassificationSection(info))
	appendSection(vettingRiskBreakdownSection(info))
	appendSection(vettingPrivacySection(info))
	appendSection(vettingAIGovernanceSection(info))
	appendSection(vettingClausesSection(info))
	appendSection(vettingProfessionalStandingSection(info))
	appendSection(vettingBaselineSection(info))
	appendSection(vettingGapsSection(info))

	return strings.Join(sections, "\n\n")
}

func vettingBulletSection(title string, lines []string) string {
	if len(lines) == 0 {
		return ""
	}

	var b strings.Builder

	b.WriteString(title)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		b.WriteString("\n· ")
		b.WriteString(line)
	}

	return b.String()
}

func vettingOverviewSection(info ThirdPartyInfo) string {
	var lines []string

	switch {
	case info.OverallRiskRating != "" && info.OverallRiskScore > 0:
		lines = append(lines, fmt.Sprintf("Overall risk: %d/100 (%s)", info.OverallRiskScore, info.OverallRiskRating))
	case info.OverallRiskScore > 0:
		lines = append(lines, fmt.Sprintf("Overall risk: %d/100", info.OverallRiskScore))
	case info.OverallRiskRating != "":
		lines = append(lines, fmt.Sprintf("Overall risk: %s", info.OverallRiskRating))
	}

	if info.Recommendation != "" {
		lines = append(lines, fmt.Sprintf("Recommendation: %s", formatVettingRecommendation(info.Recommendation)))
	}

	return strings.Join(lines, "\n")
}

func vettingPillarSection(info ThirdPartyInfo) string {
	var parts []string

	if info.SecurityRiskScore > 0 {
		parts = append(parts, fmt.Sprintf("Security %d/100", info.SecurityRiskScore))
	}

	if info.PrivacyRiskScore > 0 {
		parts = append(parts, fmt.Sprintf("Privacy %d/100", info.PrivacyRiskScore))
	}

	if info.InvolvesAI || info.AIRiskScore > 0 {
		parts = append(parts, fmt.Sprintf("AI %d/100", info.AIRiskScore))
	}

	return strings.Join(parts, " · ")
}

func vettingClassificationSection(info ThirdPartyInfo) string {
	var lines []string

	if info.ThirdPartyType != "" {
		lines = append(lines, "Type: "+info.ThirdPartyType)
	}

	if info.PrivacyRole != "" {
		lines = append(lines, "Privacy role: "+info.PrivacyRole)
	}

	lines = append(lines, "Processes PII: "+vettingYesNo(info.ProcessesPII))

	if info.CrossBorderTransfer {
		lines = append(lines, "Cross-border transfers: yes")
	}

	if info.InvolvesAI {
		ai := "AI involvement: yes"
		if useCases := nonEmptyStrings(info.AIUseCases); len(useCases) > 0 {
			ai += " (" + strings.Join(useCases, ", ") + ")"
		}

		lines = append(lines, ai)
	}

	return vettingBulletSection("Classification", lines)
}

func vettingRiskBreakdownSection(info ThirdPartyInfo) string {
	var lines []string

	for _, score := range info.RiskScores {
		if score.Category == "" {
			continue
		}

		line := score.Category
		if score.Rating != "" {
			line += " — " + score.Rating
		}

		if score.Notes != "" {
			line += ": " + score.Notes
		}

		lines = append(lines, line)
	}

	return vettingBulletSection("Risk breakdown", lines)
}

func vettingPrivacySection(info ThirdPartyInfo) string {
	var lines []string

	if info.DPAStatus != "" {
		lines = append(lines, "DPA: "+info.DPAStatus)
	}

	if info.DSARCapability != "" {
		lines = append(lines, "DSAR: "+info.DSARCapability)
	}

	if info.RetentionPolicy != "" {
		lines = append(lines, "Retention: "+info.RetentionPolicy)
	}

	if info.DeletionPolicy != "" {
		lines = append(lines, "Deletion: "+info.DeletionPolicy)
	}

	if info.DataMinimization != "" {
		lines = append(lines, "Data minimization: "+info.DataMinimization)
	}

	if info.PurposeLimitation != "" {
		lines = append(lines, "Purpose limitation: "+info.PurposeLimitation)
	}

	return vettingBulletSection("Privacy & data processing", lines)
}

func vettingAIGovernanceSection(info ThirdPartyInfo) string {
	if !info.InvolvesAI {
		return ""
	}

	var lines []string

	if info.AITransparency != "" {
		lines = append(lines, "Transparency: "+info.AITransparency)
	}

	if info.BiasControls != "" {
		lines = append(lines, "Bias controls: "+info.BiasControls)
	}

	if info.HumanOversight != "" {
		lines = append(lines, "Human oversight: "+info.HumanOversight)
	}

	if info.TrainingDataGovernance != "" {
		lines = append(lines, "Training data governance: "+info.TrainingDataGovernance)
	}

	if info.AIGovernanceDocURL != "" {
		lines = append(lines, "Governance doc: "+info.AIGovernanceDocURL)
	}

	return vettingBulletSection("AI governance", lines)
}

func vettingClausesSection(info ThirdPartyInfo) string {
	var lines []string

	for _, clause := range info.PrivacyClauses {
		if strings.TrimSpace(clause) != "" {
			lines = append(lines, "Privacy: "+clause)
		}
	}

	for _, clause := range info.AIClauses {
		if strings.TrimSpace(clause) != "" {
			lines = append(lines, "AI: "+clause)
		}
	}

	return vettingBulletSection("Contractual clauses", lines)
}

func vettingProfessionalStandingSection(info ThirdPartyInfo) string {
	var lines []string

	for _, license := range info.ProfessionalLicenses {
		if strings.TrimSpace(license) != "" {
			lines = append(lines, "License: "+license)
		}
	}

	for _, membership := range info.IndustryMemberships {
		if strings.TrimSpace(membership) != "" {
			lines = append(lines, "Membership: "+membership)
		}
	}

	if info.InsuranceCoverage != "" {
		lines = append(lines, "Insurance: "+info.InsuranceCoverage)
	}

	return vettingBulletSection("Professional standing", lines)
}

func vettingBaselineSection(info ThirdPartyInfo) string {
	failures := nonEmptyStrings(info.BaselineFailures)
	if len(failures) == 0 {
		return ""
	}

	return vettingBulletSection("Minimum baseline not met", failures)
}

func vettingGapsSection(info ThirdPartyInfo) string {
	gaps := nonEmptyStrings(info.InformationGaps)
	if len(gaps) > maxVettingNotesGaps {
		gaps = gaps[:maxVettingNotesGaps]
	}

	return vettingBulletSection("Gaps", gaps)
}

func vettingYesNo(value bool) string {
	if value {
		return "yes"
	}

	return "no"
}

func nonEmptyStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			out = append(out, strings.TrimSpace(value))
		}
	}

	return out
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
	nameSuffixPath []string,
) {
	if p.Name != "" {
		// Keep the name hierarchy-qualified for sub third parties; a top-level
		// third party (empty suffix path) keeps the bare name.
		thirdParty.Name = qualifyThirdPartyName(p.Name, nameSuffixPath)
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
	parentLevel int,
	parentNamePath []string,
	p linkSubThirdPartyParams,
) error {
	if p.Name == "" {
		return nil
	}

	// Auto-discovered subprocessors must not nest beyond the maximum level.
	// Stop descending here rather than creating an invalid child.
	if parentLevel+1 > coredata.MaxThirdPartyLevel {
		return nil
	}

	// Store and match the child under its hierarchy-qualified name so vetting
	// agrees with names created from the console (e.g. "aws (Probo)").
	qualifiedName := qualifyThirdPartyName(p.Name, parentNamePath)

	child := &coredata.ThirdParty{}

	// Sub-third-parties are scoped per parent, so a child is matched by name
	// within this parent only — a same-named third party under a different
	// parent is an independent entity and must be created here too.
	err := child.LoadByNameAndParentThirdPartyID(ctx, conn, scope, qualifiedName, pc.ThirdPartyID)
	switch {
	case err == nil:
		if countries := parseOptionalCountryCodes(p.Country); len(countries) > 0 && len(child.Countries) == 0 {
			child.Countries = countries
			child.UpdatedAt = time.Now()

			if err := child.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update child third party %q countries: %w", p.Name, err)
			}
		}

		return nil
	case !errors.Is(err, coredata.ErrResourceNotFound):
		return fmt.Errorf("cannot find child third party %q: %w", p.Name, err)
	}

	now := time.Now()
	parentID := pc.ThirdPartyID
	child = &coredata.ThirdParty{
		ID:                 gid.New(scope.GetTenantID(), coredata.ThirdPartyEntityType),
		OrganizationID:     pc.OrganizationID,
		ParentThirdPartyID: &parentID,
		Name:               qualifiedName,
		Category:           coredata.ThirdPartyCategoryOther,
		Level:              parentLevel + 1,
		CreatedAt:          now,
		UpdatedAt:          now,
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

	return nil
}

// baseThirdPartyName strips a trailing " (path)" suffix so a stored,
// hierarchy-qualified name is reduced to its bare base, mirroring the console
// UI convention.
func baseThirdPartyName(name string) string {
	return strings.TrimSpace(nameSuffixPattern.ReplaceAllString(name, ""))
}

// qualifyThirdPartyName appends the parent path as a parenthesized suffix, e.g.
// ("aws", ["Probo", "Acme"]) → "aws (Probo/Acme)". An empty path leaves the
// name unchanged, so top-level third parties are never suffixed.
func qualifyThirdPartyName(base string, path []string) string {
	if len(path) == 0 {
		return base
	}

	return fmt.Sprintf("%s (%s)", base, strings.Join(path, "/"))
}

// loadAncestorBaseNames returns the base names of a third party's ancestors,
// ordered root → immediate parent. It is the suffix path used to qualify the
// third party's own name; append the third party's own base name to it to get
// the suffix path for its children.
func loadAncestorBaseNames(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	thirdPartyID gid.GID,
) ([]string, error) {
	var ancestors coredata.ThirdParties

	if err := ancestors.LoadAllAncestorsByThirdPartyID(ctx, conn, scope, thirdPartyID); err != nil {
		return nil, fmt.Errorf("cannot load ancestors: %w", err)
	}

	names := make([]string, len(ancestors))
	for i, ancestor := range ancestors {
		names[i] = baseThirdPartyName(ancestor.Name)
	}

	return names, nil
}
