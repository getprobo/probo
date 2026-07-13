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

import "go.probo.inc/probo/pkg/coredata"

func p0MeasurePlanRules() []measurePlanRule {
	return []measurePlanRule{
		{
			factKey:     "org_fork_pr_approval_required",
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
			factKey:     "org_enterprise_accessible",
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
			factKey:     "repo_production_classification",
			name:        "Production repository classification",
			description: "Likely production repositories are identified for deeper checks.",
			category:    "governance",
			evaluate:    evaluateAnyCoverage,
		},
		{
			factKey:     "repo_signed_commits_practice_coverage",
			name:        "Signed commits in practice",
			description: "Recent commits on default branches are cryptographically signed.",
			category:    "code_integrity",
			evaluate:    evaluateAnyCoverage,
		},
		{
			factKey:     "repo_force_push_disabled_coverage",
			name:        "Force push disabled on default branch",
			description: "Default branches disallow force pushes.",
			category:    "code_review",
			evaluate:    evaluateFullCoverage,
		},
		{
			factKey:     "repo_required_status_checks_coverage",
			name:        "Required status checks on default branch",
			description: "Default branches require status checks before merge.",
			category:    "ci_cd",
			evaluate:    evaluateAnyCoverage,
		},
		{
			factKey:     "repo_bypass_actor_restrictions_coverage",
			name:        "Branch protection bypass restrictions",
			description: "Branch protection limits who can bypass required checks.",
			category:    "code_review",
			evaluate:    evaluateAnyCoverage,
		},
		{
			factKey:     "repo_pr_ci_coverage",
			name:        "CI runs on pull requests",
			description: "Workflows run on pull request events.",
			category:    "ci_cd",
			evaluate:    evaluateAnyCoverage,
		},
		{
			factKey:     "repo_pull_request_target_risk",
			name:        "No pull_request_target workflow risk",
			description: "Scanned repositories avoid dangerous pull_request_target workflows.",
			category:    "ci_cd",
			evaluate:    evaluateCoverageRiskAbsent,
		},
		{
			factKey:     "repo_codeql_enabled_coverage",
			name:        "CodeQL analysis in CI",
			description: "Repositories run CodeQL or equivalent code scanning in CI.",
			category:    "code_scanning",
			evaluate:    evaluateAnyCoverage,
		},
		{
			factKey:     "repo_codeql_default_setup_coverage",
			name:        "CodeQL default setup enabled",
			description: "Repositories enable GitHub code scanning default setup.",
			category:    "code_scanning",
			evaluate:    evaluateAnyCoverage,
		},
		{
			factKey:     "repo_dependency_review_coverage",
			name:        "Dependency review in CI",
			description: "Repositories run dependency review on pull requests.",
			category:    "dependencies",
			evaluate:    evaluateAnyCoverage,
		},
		{
			factKey:     "repo_sast_in_ci_coverage",
			name:        "SAST in CI",
			description: "Repositories run static analysis security testing in CI.",
			category:    "code_scanning",
			evaluate:    evaluateAnyCoverage,
		},
		{
			factKey:     "repo_dep_scan_in_ci_coverage",
			name:        "Dependency scanning in CI",
			description: "Repositories scan dependencies in CI pipelines.",
			category:    "dependencies",
			evaluate:    evaluateAnyCoverage,
		},
		{
			factKey:     "repo_development_guide_coverage",
			name:        "Development guide documented",
			description: "Repositories publish engineering development guidance.",
			category:    "documentation",
			evaluate:    evaluateAnyCoverage,
		},
		{
			factKey:     "repo_code_review_guide_coverage",
			name:        "Code review guide documented",
			description: "Repositories publish code review guidance.",
			category:    "documentation",
			evaluate:    evaluateAnyCoverage,
		},
		{
			factKey:     "repo_renovate_config_coverage",
			name:        "Renovate dependency automation",
			description: "Repositories configure Renovate or equivalent update automation.",
			category:    "dependencies",
			evaluate:    evaluateAnyCoverage,
		},
		{
			factKey:     "repo_lockfile_coverage",
			name:        "Dependency lock files maintained",
			description: "Repositories maintain dependency lock files.",
			category:    "dependencies",
			evaluate:    evaluateAnyCoverage,
		},
		{
			factKey:     "repo_secret_scanning_push_protection_coverage",
			name:        "Secret scanning push protection",
			description: "Repositories enable secret scanning push protection.",
			category:    "secrets",
			evaluate:    evaluateAnyCoverage,
		},
		{
			factKey:     "repo_env_on_default_branch",
			name:        "No secrets committed to default branch",
			description: "Default branches do not contain .env files.",
			category:    "secrets",
			evaluate: func(f Fact) coredata.MeasureState {
				return evaluateCountRiskAbsent(f.Value, "repos_with_env")
			},
		},
		{
			factKey:     "repo_deploy_keys_write_access",
			name:        "Deploy keys with write access controlled",
			description: "Write-capable deploy keys are limited across scanned repositories.",
			category:    "secrets",
			evaluate: func(f Fact) coredata.MeasureState {
				return evaluateCountRiskAbsent(f.Value, "write_keys")
			},
		},
		{
			factKey:     "repo_commit_status_ci_coverage",
			name:        "CI detected via commit statuses",
			description: "Repositories report CI results via commit statuses or check runs.",
			category:    "ci_cd",
			evaluate:    evaluateAnyCoverage,
		},
		{
			factKey:     "repo_external_ci_coverage",
			name:        "External CI providers detected",
			description: "Repositories use external CI providers such as CircleCI or Jenkins.",
			category:    "ci_cd",
			evaluate:    evaluateAnyCoverage,
		},
		{
			factKey:     "repo_ci_providers",
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
			factKey:     "repo_security_contact_coverage",
			name:        "Security contact in SECURITY.md",
			description: "Repositories publish a reachable security contact in SECURITY.md.",
			category:    "documentation",
			evaluate:    evaluateAnyCoverage,
		},
		{
			factKey:     "repo_incident_response_doc_coverage",
			name:        "Incident response documentation",
			description: "Repositories document incident response procedures.",
			category:    "documentation",
			evaluate:    evaluateAnyCoverage,
		},
		{
			factKey:     "repo_issue_templates_coverage",
			name:        "Issue templates configured",
			description: "Repositories provide GitHub issue templates.",
			category:    "documentation",
			evaluate:    evaluateAnyCoverage,
		},
		{
			factKey:     "repo_de_facto_pr_review_coverage",
			name:        "Pull requests reviewed in practice",
			description: "Merged pull requests receive approvals in practice.",
			category:    "code_review",
			evaluate:    evaluateAnyCoverage,
		},
		{
			factKey:     "repo_pr_approval_rate",
			name:        "Pull request approval rate",
			description: "Merged pull requests are approved before merge.",
			category:    "code_review",
			evaluate: func(f Fact) coredata.MeasureState {
				return evaluatePRApprovalRate(f.Value)
			},
		},
		{
			factKey:     "org_profile_security_md",
			name:        "Organization security disclosure policy",
			description: "The organization profile repository publishes SECURITY.md.",
			category:    "documentation",
			evaluate: func(f Fact) coredata.MeasureState {
				return evaluateOrgProfileSecurityMD(f.Value)
			},
		},
		{
			factKey:     "org_profile_contributing_md",
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
