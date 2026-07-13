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
	"strings"

	"go.probo.inc/probo/pkg/coredata"
)

type (
	measurePlanRule struct {
		name        string
		description string
		category    string
		evaluate    func(Fact) coredata.MeasureState
	}

	gidKey string
)

// buildMeasurePlanFromFacts maps collected facts to creates/updates without an LLM.
func buildMeasurePlanFromFacts(sheet *FactSheet, existing []ExistingMeasure) (*MeasurePlan, error) {
	rules := defaultMeasurePlanRules()
	byName := map[string]Fact{}

	for _, fact := range sheet.Facts {
		byName[fact.Name] = fact
	}

	plan := &MeasurePlan{
		Unchanged: []MeasurePlanUnchanged{},
	}

	used := map[gidKey]struct{}{}

	for _, rule := range rules {
		fact, ok := byName[rule.name]
		if !ok {
			continue
		}

		state := rule.evaluate(fact)
		summary := rule.description

		if match := findMeasureByName(existing, rule.name); match != nil {
			plan.Updates = append(plan.Updates, MeasurePlanUpdate{
				MeasureID:       match.ID,
				State:           state,
				EvidenceSummary: summary,
			})
			used[gidKey(match.ID.String())] = struct{}{}

			continue
		}

		if len(plan.Creates) >= maxMeasureCreatesPerRun {
			continue
		}

		plan.Creates = append(plan.Creates, MeasurePlanCreate{
			Name:            rule.name,
			Description:     rule.description,
			Category:        rule.category,
			State:           state,
			EvidenceSummary: summary,
		})
	}

	for _, m := range existing {
		if _, ok := used[gidKey(m.ID.String())]; ok {
			continue
		}

		plan.Unchanged = append(plan.Unchanged, MeasurePlanUnchanged{
			MeasureID: m.ID,
			Reason:    "No matching facts in this run.",
		})
	}

	return plan, nil
}

func findMeasureByName(existing []ExistingMeasure, name string) *ExistingMeasure {
	target := strings.ToLower(strings.TrimSpace(name))

	for i := range existing {
		if strings.ToLower(strings.TrimSpace(existing[i].Name)) == target {
			return &existing[i]
		}
	}

	return nil
}

func defaultMeasurePlanRules() []measurePlanRule {
	rules := []measurePlanRule{
		{
			name:        MeasureOrgMFARequired,
			description: "Organization requires two-factor authentication for all members.",
			category:    "access",
			evaluate: func(f Fact) coredata.MeasureState {
				if b, ok := f.Value.(bool); ok && b {
					return coredata.MeasureStateImplemented
				}

				return coredata.MeasureStateNotImplemented
			},
		},
		{
			name:        MeasureOrgNo2FAMembers,
			description: "No organization members lack two-factor authentication.",
			category:    "access",
			evaluate: func(f Fact) coredata.MeasureState {
				count, ok := factIntValue(f.Value)
				if !ok {
					return coredata.MeasureStateUnknown
				}

				if count == 0 {
					return coredata.MeasureStateImplemented
				}

				return coredata.MeasureStateNotImplemented
			},
		},
		{
			name:        MeasureOrgBasePermissions,
			description: "Default repository permission is read or none.",
			category:    "access",
			evaluate: func(f Fact) coredata.MeasureState {
				perm, ok := f.Value.(string)
				if !ok {
					return coredata.MeasureStateUnknown
				}

				switch strings.ToLower(perm) {
				case "none", "read":
					return coredata.MeasureStateImplemented
				default:
					return coredata.MeasureStateNotImplemented
				}
			},
		},
		{
			name:        MeasureOrgNoPublicRepoCreation,
			description: "Members cannot create public repositories.",
			category:    "exposure",
			evaluate: func(f Fact) coredata.MeasureState {
				if b, ok := f.Value.(bool); ok && b {
					return coredata.MeasureStateImplemented
				}

				return coredata.MeasureStateNotImplemented
			},
		},
		{
			name:        MeasureOrgAdminMinimization,
			description: "Organization admin accounts are limited.",
			category:    "access",
			evaluate: func(f Fact) coredata.MeasureState {
				return evaluateAdminMinimization(f.Value)
			},
		},
		{
			name:        MeasureOrgPublicRepos,
			description: "Unexpected public repositories are controlled.",
			category:    "exposure",
			evaluate: func(f Fact) coredata.MeasureState {
				public, total, ok := factCountPair(f.Value)
				if !ok {
					return coredata.MeasureStateUnknown
				}

				if total == 0 || public == 0 {
					return coredata.MeasureStateImplemented
				}

				return coredata.MeasureStateNotImplemented
			},
		},
		{
			name:        MeasureOrgNoVisibilityChange,
			description: "Members cannot change repository visibility.",
			category:    "exposure",
			evaluate: func(f Fact) coredata.MeasureState {
				if b, ok := f.Value.(bool); ok && b {
					return coredata.MeasureStateImplemented
				}

				return coredata.MeasureStateNotImplemented
			},
		},
		{
			name:        MeasureOrgOutsideCollaborators,
			description: "Outside collaborators are inventoried for review.",
			category:    "access",
			evaluate: func(f Fact) coredata.MeasureState {
				_, ok := factCountValue(f.Value, "count")
				if !ok {
					return coredata.MeasureStateUnknown
				}

				return coredata.MeasureStateImplemented
			},
		},
		{
			name:        MeasureOrgActionsRestricted,
			description: "Organization restricts which actions and workflows may run.",
			category:    "ci_cd",
			evaluate: func(f Fact) coredata.MeasureState {
				restricted, ok := factBoolField(f.Value, "restricted")
				if !ok {
					return coredata.MeasureStateUnknown
				}

				if restricted {
					return coredata.MeasureStateImplemented
				}

				return coredata.MeasureStateNotImplemented
			},
		},
		{
			name:        MeasureOrgGitHubApps,
			description: "Installed GitHub Apps are inventoried for review.",
			category:    "integrations",
			evaluate: func(f Fact) coredata.MeasureState {
				_, ok := factCountValue(f.Value, "installations")
				if !ok {
					return coredata.MeasureStateUnknown
				}

				return coredata.MeasureStateImplemented
			},
		},
		{
			name:        MeasureOrgAuditLogAccessible,
			description: "Organization audit log is accessible for review.",
			category:    "audit",
			evaluate: func(f Fact) coredata.MeasureState {
				if b, ok := f.Value.(bool); ok && b {
					return coredata.MeasureStateImplemented
				}

				if b, ok := f.Value.(bool); ok && !b {
					return coredata.MeasureStateNotApplicable
				}

				return coredata.MeasureStateUnknown
			},
		},
		{
			name:        MeasureRepoBranchProtectionCoverage,
			description: "Default branches are protected across scanned repositories.",
			category:    "code_review",
			evaluate:    evaluateFullCoverage,
		},
		{
			name:        MeasureRepoPRReviewsRequiredCoverage,
			description: "Default branches require pull request reviews.",
			category:    "code_review",
			evaluate:    evaluateFullCoverage,
		},
		{
			name:        MeasureRepoSignedCommitsRequiredCoverage,
			description: "Default branches require signed commits.",
			category:    "code_integrity",
			evaluate:    evaluateFullCoverage,
		},
		{
			name:        MeasureRepoWorkflowCoverage,
			description: "Repositories run automated workflows.",
			category:    "ci_cd",
			evaluate:    evaluateAnyCoverage,
		},
		{
			name:        MeasureRepoSecurityMDCoverage,
			description: "Repositories publish a SECURITY.md disclosure policy.",
			category:    "documentation",
			evaluate:    evaluateAnyCoverage,
		},
		{
			name:        MeasureRepoContributingMDCoverage,
			description: "Repositories publish CONTRIBUTING.md guidance.",
			category:    "documentation",
			evaluate:    evaluateAnyCoverage,
		},
		{
			name:        MeasureRepoDependabotConfigCoverage,
			description: "Repositories configure Dependabot update automation.",
			category:    "dependencies",
			evaluate:    evaluateAnyCoverage,
		},
		{
			name:        MeasureRepoDependabotCriticalOpen,
			description: "No open critical Dependabot alerts in scanned repositories.",
			category:    "dependencies",
			evaluate: func(f Fact) coredata.MeasureState {
				count, ok := factCountValue(f.Value, "open_critical")
				if !ok {
					return coredata.MeasureStateUnknown
				}

				if count == 0 {
					return coredata.MeasureStateImplemented
				}

				return coredata.MeasureStateNotImplemented
			},
		},
		{
			name:        MeasureRepoSecretScanningAlertsOpen,
			description: "No open secret scanning alerts in scanned repositories.",
			category:    "secrets",
			evaluate: func(f Fact) coredata.MeasureState {
				count, ok := factCountValue(f.Value, "open")
				if !ok {
					return coredata.MeasureStateUnknown
				}

				if count == 0 {
					return coredata.MeasureStateImplemented
				}

				return coredata.MeasureStateNotImplemented
			},
		},
		{
			name:        MeasureRepoCodeScanningCriticalOpen,
			description: "No open critical code scanning alerts in scanned repositories.",
			category:    "code_scanning",
			evaluate: func(f Fact) coredata.MeasureState {
				count, ok := factCountValue(f.Value, "open_critical")
				if !ok {
					return coredata.MeasureStateUnknown
				}

				if count == 0 {
					return coredata.MeasureStateImplemented
				}

				return coredata.MeasureStateNotImplemented
			},
		},
		{
			name:        MeasureOrgForkPRApprovalRequired,
			description: "Workflows from fork pull requests require approval before running.",
			category:    "ci_cd",
			evaluate: func(f Fact) coredata.MeasureState {
				if b, ok := f.Value.(bool); ok && b {
					return coredata.MeasureStateImplemented
				}

				return coredata.MeasureStateNotImplemented
			},
		},
		{
			name:        MeasureOrgEnterpriseAccessible,
			description: "Enterprise configuration is accessible for governance review.",
			category:    "audit",
			evaluate: func(f Fact) coredata.MeasureState {
				if b, ok := f.Value.(bool); ok && b {
					return coredata.MeasureStateImplemented
				}

				return coredata.MeasureStateNotApplicable
			},
		},
		{
			name:        MeasureRepoProductionClassification,
			description: "Likely production repositories are identified for deeper checks.",
			category:    "governance",
			evaluate:    evaluateAnyCoverage,
		},
		{
			name:        MeasureRepoSignedCommitsPracticeCoverage,
			description: "Recent commits on default branches are cryptographically signed.",
			category:    "code_integrity",
			evaluate:    evaluateAnyCoverage,
		},
		{
			name:        MeasureRepoForcePushDisabledCoverage,
			description: "Default branches disallow force pushes.",
			category:    "code_review",
			evaluate:    evaluateFullCoverage,
		},
		{
			name:        MeasureRepoRequiredStatusChecksCoverage,
			description: "Default branches require status checks before merge.",
			category:    "ci_cd",
			evaluate:    evaluateAnyCoverage,
		},
		{
			name:        MeasureRepoBypassActorRestrictionsCoverage,
			description: "Branch protection limits who can bypass required checks.",
			category:    "code_review",
			evaluate:    evaluateAnyCoverage,
		},
		{
			name:        MeasureRepoPRCICoverage,
			description: "CI results are observed on recent merged pull requests via commit statuses or check runs.",
			category:    "ci_cd",
			evaluate:    evaluateAnyCoverage,
		},
		{
			name:        MeasureRepoPullRequestTargetRisk,
			description: "Scanned repositories avoid dangerous pull_request_target workflows.",
			category:    "ci_cd",
			evaluate:    evaluateCoverageRiskAbsent,
		},
		{
			name:        MeasureRepoCodeQLEnabledCoverage,
			description: "Repositories show CodeQL or equivalent code scanning in commit statuses or check runs.",
			category:    "code_scanning",
			evaluate:    evaluateAnyCoverage,
		},
		{
			name:        MeasureRepoCodeQLDefaultSetupCoverage,
			description: "Repositories enable GitHub code scanning default setup.",
			category:    "code_scanning",
			evaluate:    evaluateAnyCoverage,
		},
		{
			name:        MeasureRepoDependencyReviewCoverage,
			description: "Repositories show dependency review in commit statuses or check runs.",
			category:    "dependencies",
			evaluate:    evaluateAnyCoverage,
		},
		{
			name:        MeasureRepoSASTInCICoverage,
			description: "Repositories show static analysis security testing in commit statuses or check runs.",
			category:    "code_scanning",
			evaluate:    evaluateAnyCoverage,
		},
		{
			name:        MeasureRepoDepScanInCICoverage,
			description: "Repositories show dependency scanning in commit statuses or check runs.",
			category:    "dependencies",
			evaluate:    evaluateAnyCoverage,
		},
		{
			name:        MeasureRepoDevelopmentGuideCoverage,
			description: "Repositories publish engineering development guidance.",
			category:    "documentation",
			evaluate:    evaluateAnyCoverage,
		},
		{
			name:        MeasureRepoCodeReviewGuideCoverage,
			description: "Repositories publish code review guidance.",
			category:    "documentation",
			evaluate:    evaluateAnyCoverage,
		},
		{
			name:        MeasureRepoRenovateConfigCoverage,
			description: "Repositories configure Renovate or equivalent update automation.",
			category:    "dependencies",
			evaluate:    evaluateAnyCoverage,
		},
		{
			name:        MeasureRepoLockfileCoverage,
			description: "Repositories maintain dependency lock files.",
			category:    "dependencies",
			evaluate:    evaluateAnyCoverage,
		},
		{
			name:        MeasureRepoSecretScanningPushProtection,
			description: "Repositories enable secret scanning push protection.",
			category:    "secrets",
			evaluate:    evaluateAnyCoverage,
		},
		{
			name:        MeasureRepoEnvOnDefaultBranch,
			description: "Default branches do not contain .env files.",
			category:    "secrets",
			evaluate: func(f Fact) coredata.MeasureState {
				return evaluateCountRiskAbsent(f.Value, "repos_with_env")
			},
		},
		{
			name:        MeasureRepoDeployKeysWriteAccess,
			description: "Write-capable deploy keys are limited across scanned repositories.",
			category:    "secrets",
			evaluate: func(f Fact) coredata.MeasureState {
				return evaluateCountRiskAbsent(f.Value, "write_keys")
			},
		},
		{
			name:        MeasureRepoCommitStatusCICoverage,
			description: "Repositories report CI results via commit statuses or check runs.",
			category:    "ci_cd",
			evaluate:    evaluateAnyCoverage,
		},
		{
			name:        MeasureRepoExternalCICoverage,
			description: "Repositories use external CI providers such as CircleCI or Jenkins.",
			category:    "ci_cd",
			evaluate:    evaluateAnyCoverage,
		},
		{
			name:        MeasureRepoCIProviders,
			description: "CI providers are inventoried from commit statuses and check runs.",
			category:    "ci_cd",
			evaluate: func(f Fact) coredata.MeasureState {
				providers, ok := factProviderMap(f.Value)
				if !ok || len(providers) == 0 {
					return coredata.MeasureStateNotImplemented
				}

				return coredata.MeasureStateImplemented
			},
		},
		{
			name:        MeasureRepoSecurityContactCoverage,
			description: "Repositories publish a reachable security contact in SECURITY.md.",
			category:    "documentation",
			evaluate:    evaluateAnyCoverage,
		},
		{
			name:        MeasureRepoIncidentResponseDocCoverage,
			description: "Repositories document incident response procedures.",
			category:    "documentation",
			evaluate:    evaluateAnyCoverage,
		},
		{
			name:        MeasureRepoIssueTemplatesCoverage,
			description: "Repositories provide GitHub issue templates.",
			category:    "documentation",
			evaluate:    evaluateAnyCoverage,
		},
		{
			name:        MeasureRepoDeFactoPRReviewCoverage,
			description: "Merged pull requests receive approvals in practice.",
			category:    "code_review",
			evaluate:    evaluateAnyCoverage,
		},
		{
			name:        MeasureRepoPRApprovalRate,
			description: "Merged pull requests are approved before merge.",
			category:    "code_review",
			evaluate: func(f Fact) coredata.MeasureState {
				return evaluatePRApprovalRate(f.Value)
			},
		},
		{
			name:        MeasureOrgProfileSecurityMD,
			description: "The organization profile repository publishes SECURITY.md.",
			category:    "documentation",
			evaluate: func(f Fact) coredata.MeasureState {
				return evaluateOrgProfileSecurityMD(f.Value)
			},
		},
		{
			name:        MeasureOrgProfileContributingMD,
			description: "The organization profile repository publishes CONTRIBUTING.md.",
			category:    "documentation",
			evaluate: func(f Fact) coredata.MeasureState {
				if b, ok := f.Value.(bool); ok && b {
					return coredata.MeasureStateImplemented
				}

				return coredata.MeasureStateNotImplemented
			},
		},
	}

	return rules
}

func factIntValue(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

func factCountPair(value any) (int, int, bool) {
	switch m := value.(type) {
	case map[string]int:
		return m["public"], m["total"], true
	case map[string]any:
		public, ok1 := toInt(m["public"])
		total, ok2 := toInt(m["total"])

		return public, total, ok1 && ok2
	default:
		return 0, 0, false
	}
}

func factCoveragePair(value any) (int, int, bool) {
	switch m := value.(type) {
	case map[string]int:
		return m["matched"], m["total"], m["total"] > 0
	case map[string]any:
		matched, ok1 := toInt(m["matched"])
		total, ok2 := toInt(m["total"])

		return matched, total, ok1 && ok2 && total > 0
	default:
		return 0, 0, false
	}
}

func factCountValue(value any, key string) (int, bool) {
	switch m := value.(type) {
	case map[string]int:
		v, ok := m[key]

		return v, ok
	case map[string]any:
		return toInt(m[key])
	default:
		return 0, false
	}
}

func factBoolField(value any, key string) (bool, bool) {
	m, ok := value.(map[string]any)
	if !ok {
		return false, false
	}

	v, ok := m[key].(bool)

	return v, ok
}

func evaluateFullCoverage(f Fact) coredata.MeasureState {
	matched, total, ok := factCoveragePair(f.Value)
	if !ok {
		return coredata.MeasureStateUnknown
	}

	if matched == total {
		return coredata.MeasureStateImplemented
	}

	if matched == 0 {
		return coredata.MeasureStateNotImplemented
	}

	return coredata.MeasureStateNotImplemented
}

func evaluateAnyCoverage(f Fact) coredata.MeasureState {
	matched, _, ok := factCoveragePair(f.Value)
	if !ok {
		return coredata.MeasureStateUnknown
	}

	if matched > 0 {
		return coredata.MeasureStateImplemented
	}

	return coredata.MeasureStateNotImplemented
}

func factProviderMap(value any) (map[string]int, bool) {
	m, ok := value.(map[string]any)
	if !ok {
		return nil, false
	}

	raw, ok := m["providers"].(map[string]int)
	if ok {
		return raw, true
	}

	rawAny, ok := m["providers"].(map[string]any)
	if !ok {
		return nil, false
	}

	out := make(map[string]int, len(rawAny))
	for key, val := range rawAny {
		count, ok := toInt(val)
		if !ok {
			continue
		}

		out[key] = count
	}

	return out, len(out) > 0
}

func evaluatePRApprovalRate(value any) coredata.MeasureState {
	reviewed, ok1 := factCountValue(value, "reviewed")

	sampled, ok2 := factCountValue(value, "sampled")
	if !ok1 || !ok2 || sampled == 0 {
		return coredata.MeasureStateUnknown
	}

	if reviewed*100/sampled >= 80 {
		return coredata.MeasureStateImplemented
	}

	return coredata.MeasureStateNotImplemented
}

func evaluateOrgProfileSecurityMD(value any) coredata.MeasureState {
	m, ok := value.(map[string]any)
	if !ok {
		return coredata.MeasureStateUnknown
	}

	present, ok := m["present"].(bool)
	if !ok || !present {
		return coredata.MeasureStateNotImplemented
	}

	contact, ok := m["security_contact"].(bool)
	if ok && contact {
		return coredata.MeasureStateImplemented
	}

	return coredata.MeasureStateNotImplemented
}

func evaluateCoverageRiskAbsent(f Fact) coredata.MeasureState {
	matched, _, ok := factCoveragePair(f.Value)
	if !ok {
		return coredata.MeasureStateUnknown
	}

	if matched == 0 {
		return coredata.MeasureStateImplemented
	}

	return coredata.MeasureStateNotImplemented
}

func evaluateCountRiskAbsent(value any, key string) coredata.MeasureState {
	count, ok := factCountValue(value, key)
	if !ok {
		return coredata.MeasureStateUnknown
	}

	if count == 0 {
		return coredata.MeasureStateImplemented
	}

	return coredata.MeasureStateNotImplemented
}
