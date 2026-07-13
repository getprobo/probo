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
	"fmt"
	"strings"

	"go.probo.inc/probo/pkg/coredata"
)

type (
	measurePlanRule struct {
		check       Check
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
	byCheck := map[Check]Fact{}

	for _, fact := range sheet.Facts {
		byCheck[fact.Check] = fact
	}

	plan := &MeasurePlan{
		Unchanged: []MeasurePlanUnchanged{},
	}

	used := map[gidKey]struct{}{}

	for _, rule := range rules {
		fact, ok := byCheck[rule.check]
		if !ok {
			continue
		}

		state := rule.evaluate(fact)
		summary := fmt.Sprintf("%s (check %s)", rule.description, fact.Check)

		if match := findMeasureByName(existing, rule.name); match != nil {
			plan.Updates = append(plan.Updates, MeasurePlanUpdate{
				MeasureID:       match.ID,
				State:           state,
				EvidenceSummary: summary,
				CheckRefs:       []Check{fact.Check},
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
			CheckRefs:       []Check{fact.Check},
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
			check:       CheckOrgMFARequired,
			name:        "Org-wide MFA enforcement",
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
			check:       CheckOrgNo2FAMembers,
			name:        "Members without 2FA",
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
			check:       CheckOrgBasePermissions,
			name:        "Minimal default repository permissions",
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
			check:       CheckOrgNoPublicRepoCreation,
			name:        "Restrict public repository creation",
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
			check:       CheckOrgAdminMinimization,
			name:        "Admin account minimization",
			description: "Organization admin accounts are limited.",
			category:    "access",
			evaluate: func(f Fact) coredata.MeasureState {
				return evaluateAdminMinimization(f.Value)
			},
		},
		{
			check:       CheckOrgPublicRepos,
			name:        "Public repository exposure",
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
			check:       CheckOrgNoVisibilityChange,
			name:        "Restrict repository visibility changes",
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
			check:       CheckOrgOutsideCollaborators,
			name:        "Outside collaborator inventory",
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
			check:       CheckOrgActionsRestricted,
			name:        "GitHub Actions usage restricted",
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
			check:       CheckOrgGitHubApps,
			name:        "GitHub App inventory",
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
			check:       CheckOrgAuditLogAccessible,
			name:        "Audit log accessible",
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
			check:       CheckRepoBranchProtectionCoverage,
			name:        "Default branch protection",
			description: "Default branches are protected across scanned repositories.",
			category:    "code_review",
			evaluate:    evaluateFullCoverage,
		},
		{
			check:       CheckRepoPRReviewsRequiredCoverage,
			name:        "Pull request reviews required",
			description: "Default branches require pull request reviews.",
			category:    "code_review",
			evaluate:    evaluateFullCoverage,
		},
		{
			check:       CheckRepoSignedCommitsRequiredCoverage,
			name:        "Signed commits required",
			description: "Default branches require signed commits.",
			category:    "code_integrity",
			evaluate:    evaluateFullCoverage,
		},
		{
			check:       CheckRepoWorkflowCoverage,
			name:        "CI/CD workflows present",
			description: "Repositories run automated workflows.",
			category:    "ci_cd",
			evaluate:    evaluateAnyCoverage,
		},
		{
			check:       CheckRepoSecurityMDCoverage,
			name:        "Security disclosure policy",
			description: "Repositories publish a SECURITY.md disclosure policy.",
			category:    "documentation",
			evaluate:    evaluateAnyCoverage,
		},
		{
			check:       CheckRepoContributingMDCoverage,
			name:        "Contributing guidelines documented",
			description: "Repositories publish CONTRIBUTING.md guidance.",
			category:    "documentation",
			evaluate:    evaluateAnyCoverage,
		},
		{
			check:       CheckRepoDependabotConfigCoverage,
			name:        "Dependabot configuration",
			description: "Repositories configure Dependabot update automation.",
			category:    "dependencies",
			evaluate:    evaluateAnyCoverage,
		},
		{
			check:       CheckRepoDependabotCriticalOpen,
			name:        "Critical Dependabot alerts resolved",
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
			check:       CheckRepoSecretScanningAlertsOpen,
			name:        "Secret scanning alerts resolved",
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
			check:       CheckRepoCodeScanningCriticalOpen,
			name:        "Critical code scanning alerts resolved",
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
			check:       CheckOrgForkPRApprovalRequired,
			name:        "Fork pull request approval required",
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
			check:       CheckOrgEnterpriseAccessible,
			name:        "Enterprise settings accessible",
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
			check:       CheckRepoProductionClassification,
			name:        "Production repository classification",
			description: "Likely production repositories are identified for deeper checks.",
			category:    "governance",
			evaluate:    evaluateAnyCoverage,
		},
		{
			check:       CheckRepoSignedCommitsPracticeCoverage,
			name:        "Signed commits in practice",
			description: "Recent commits on default branches are cryptographically signed.",
			category:    "code_integrity",
			evaluate:    evaluateAnyCoverage,
		},
		{
			check:       CheckRepoForcePushDisabledCoverage,
			name:        "Force push disabled on default branch",
			description: "Default branches disallow force pushes.",
			category:    "code_review",
			evaluate:    evaluateFullCoverage,
		},
		{
			check:       CheckRepoRequiredStatusChecksCoverage,
			name:        "Required status checks on default branch",
			description: "Default branches require status checks before merge.",
			category:    "ci_cd",
			evaluate:    evaluateAnyCoverage,
		},
		{
			check:       CheckRepoBypassActorRestrictionsCoverage,
			name:        "Branch protection bypass restrictions",
			description: "Branch protection limits who can bypass required checks.",
			category:    "code_review",
			evaluate:    evaluateAnyCoverage,
		},
		{
			check:       CheckRepoPRCICoverage,
			name:        "CI runs on pull requests",
			description: "Workflows run on pull request events.",
			category:    "ci_cd",
			evaluate:    evaluateAnyCoverage,
		},
		{
			check:       CheckRepoPullRequestTargetRisk,
			name:        "No pull_request_target workflow risk",
			description: "Scanned repositories avoid dangerous pull_request_target workflows.",
			category:    "ci_cd",
			evaluate:    evaluateCoverageRiskAbsent,
		},
		{
			check:       CheckRepoCodeQLEnabledCoverage,
			name:        "CodeQL analysis in CI",
			description: "Repositories run CodeQL or equivalent code scanning in CI.",
			category:    "code_scanning",
			evaluate:    evaluateAnyCoverage,
		},
		{
			check:       CheckRepoCodeQLDefaultSetupCoverage,
			name:        "CodeQL default setup enabled",
			description: "Repositories enable GitHub code scanning default setup.",
			category:    "code_scanning",
			evaluate:    evaluateAnyCoverage,
		},
		{
			check:       CheckRepoDependencyReviewCoverage,
			name:        "Dependency review in CI",
			description: "Repositories run dependency review on pull requests.",
			category:    "dependencies",
			evaluate:    evaluateAnyCoverage,
		},
		{
			check:       CheckRepoSASTInCICoverage,
			name:        "SAST in CI",
			description: "Repositories run static analysis security testing in CI.",
			category:    "code_scanning",
			evaluate:    evaluateAnyCoverage,
		},
		{
			check:       CheckRepoDepScanInCICoverage,
			name:        "Dependency scanning in CI",
			description: "Repositories scan dependencies in CI pipelines.",
			category:    "dependencies",
			evaluate:    evaluateAnyCoverage,
		},
		{
			check:       CheckRepoDevelopmentGuideCoverage,
			name:        "Development guide documented",
			description: "Repositories publish engineering development guidance.",
			category:    "documentation",
			evaluate:    evaluateAnyCoverage,
		},
		{
			check:       CheckRepoCodeReviewGuideCoverage,
			name:        "Code review guide documented",
			description: "Repositories publish code review guidance.",
			category:    "documentation",
			evaluate:    evaluateAnyCoverage,
		},
		{
			check:       CheckRepoRenovateConfigCoverage,
			name:        "Renovate dependency automation",
			description: "Repositories configure Renovate or equivalent update automation.",
			category:    "dependencies",
			evaluate:    evaluateAnyCoverage,
		},
		{
			check:       CheckRepoLockfileCoverage,
			name:        "Dependency lock files maintained",
			description: "Repositories maintain dependency lock files.",
			category:    "dependencies",
			evaluate:    evaluateAnyCoverage,
		},
		{
			check:       CheckRepoSecretScanningPushProtection,
			name:        "Secret scanning push protection",
			description: "Repositories enable secret scanning push protection.",
			category:    "secrets",
			evaluate:    evaluateAnyCoverage,
		},
		{
			check:       CheckRepoEnvOnDefaultBranch,
			name:        "No secrets committed to default branch",
			description: "Default branches do not contain .env files.",
			category:    "secrets",
			evaluate: func(f Fact) coredata.MeasureState {
				return evaluateCountRiskAbsent(f.Value, "repos_with_env")
			},
		},
		{
			check:       CheckRepoDeployKeysWriteAccess,
			name:        "Deploy keys with write access controlled",
			description: "Write-capable deploy keys are limited across scanned repositories.",
			category:    "secrets",
			evaluate: func(f Fact) coredata.MeasureState {
				return evaluateCountRiskAbsent(f.Value, "write_keys")
			},
		},
		{
			check:       CheckRepoCommitStatusCICoverage,
			name:        "CI detected via commit statuses",
			description: "Repositories report CI results via commit statuses or check runs.",
			category:    "ci_cd",
			evaluate:    evaluateAnyCoverage,
		},
		{
			check:       CheckRepoExternalCICoverage,
			name:        "External CI providers detected",
			description: "Repositories use external CI providers such as CircleCI or Jenkins.",
			category:    "ci_cd",
			evaluate:    evaluateAnyCoverage,
		},
		{
			check:       CheckRepoCIProviders,
			name:        "CI provider inventory",
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
			check:       CheckRepoSecurityContactCoverage,
			name:        "Security contact in SECURITY.md",
			description: "Repositories publish a reachable security contact in SECURITY.md.",
			category:    "documentation",
			evaluate:    evaluateAnyCoverage,
		},
		{
			check:       CheckRepoIncidentResponseDocCoverage,
			name:        "Incident response documentation",
			description: "Repositories document incident response procedures.",
			category:    "documentation",
			evaluate:    evaluateAnyCoverage,
		},
		{
			check:       CheckRepoIssueTemplatesCoverage,
			name:        "Issue templates configured",
			description: "Repositories provide GitHub issue templates.",
			category:    "documentation",
			evaluate:    evaluateAnyCoverage,
		},
		{
			check:       CheckRepoDeFactoPRReviewCoverage,
			name:        "Pull requests reviewed in practice",
			description: "Merged pull requests receive approvals in practice.",
			category:    "code_review",
			evaluate:    evaluateAnyCoverage,
		},
		{
			check:       CheckRepoPRApprovalRate,
			name:        "Pull request approval rate",
			description: "Merged pull requests are approved before merge.",
			category:    "code_review",
			evaluate: func(f Fact) coredata.MeasureState {
				return evaluatePRApprovalRate(f.Value)
			},
		},
		{
			check:       CheckOrgProfileSecurityMD,
			name:        "Organization security disclosure policy",
			description: "The organization profile repository publishes SECURITY.md.",
			category:    "documentation",
			evaluate: func(f Fact) coredata.MeasureState {
				return evaluateOrgProfileSecurityMD(f.Value)
			},
		},
		{
			check:       CheckOrgProfileContributingMD,
			name:        "Organization contributing guidelines",
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
