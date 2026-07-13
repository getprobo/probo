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

// Check identifies a deterministic discovery observation emitted by scanners.
type Check string

const (
	CheckOrgMFARequired                      Check = "org_mfa_required"
	CheckOrgNo2FAMembers                     Check = "org_no_2fa_members"
	CheckOrgAdminMinimization                Check = "org_admin_minimization"
	CheckOrgBasePermissions                  Check = "org_base_permissions"
	CheckOrgNoPublicRepoCreation             Check = "org_no_public_repo_creation"
	CheckOrgNoVisibilityChange               Check = "org_no_visibility_change"
	CheckOrgPublicRepos                      Check = "org_public_repos"
	CheckOrgOutsideCollaborators             Check = "org_outside_collaborators"
	CheckOrgActionsRestricted                Check = "org_actions_restricted"
	CheckOrgForkPRApprovalRequired           Check = "org_fork_pr_approval_required"
	CheckOrgGitHubApps                       Check = "org_github_apps"
	CheckOrgAuditLogAccessible               Check = "org_audit_log_accessible"
	CheckOrgEnterpriseAccessible             Check = "org_enterprise_accessible"
	CheckOrgProfileSecurityMD                Check = "org_profile_security_md"
	CheckOrgProfileContributingMD            Check = "org_profile_contributing_md"
	CheckRepoProductionClassification        Check = "repo_production_classification"
	CheckRepoBranchProtectionCoverage        Check = "repo_branch_protection_coverage"
	CheckRepoPRReviewsRequiredCoverage       Check = "repo_pr_reviews_required_coverage"
	CheckRepoSignedCommitsRequiredCoverage   Check = "repo_signed_commits_required_coverage"
	CheckRepoSignedCommitsPracticeCoverage   Check = "repo_signed_commits_practice_coverage"
	CheckRepoForcePushDisabledCoverage       Check = "repo_force_push_disabled_coverage"
	CheckRepoRequiredStatusChecksCoverage    Check = "repo_required_status_checks_coverage"
	CheckRepoBypassActorRestrictionsCoverage Check = "repo_bypass_actor_restrictions_coverage"
	CheckRepoWorkflowCoverage                Check = "repo_workflow_coverage"
	CheckRepoPRCICoverage                    Check = "repo_pr_ci_coverage"
	CheckRepoPullRequestTargetRisk           Check = "repo_pull_request_target_risk"
	CheckRepoCodeQLEnabledCoverage           Check = "repo_codeql_enabled_coverage"
	CheckRepoCodeQLDefaultSetupCoverage      Check = "repo_codeql_default_setup_coverage"
	CheckRepoDependencyReviewCoverage        Check = "repo_dependency_review_coverage"
	CheckRepoSASTInCICoverage                Check = "repo_sast_in_ci_coverage"
	CheckRepoDepScanInCICoverage             Check = "repo_dep_scan_in_ci_coverage"
	CheckRepoWorkflowSecretsUsage            Check = "repo_workflow_secrets_usage"
	CheckRepoSecurityMDCoverage              Check = "repo_security_md_coverage"
	CheckRepoContributingMDCoverage          Check = "repo_contributing_md_coverage"
	CheckRepoDevelopmentGuideCoverage        Check = "repo_development_guide_coverage"
	CheckRepoCodeReviewGuideCoverage         Check = "repo_code_review_guide_coverage"
	CheckRepoDependabotConfigCoverage        Check = "repo_dependabot_config_coverage"
	CheckRepoRenovateConfigCoverage          Check = "repo_renovate_config_coverage"
	CheckRepoLockfileCoverage                Check = "repo_lockfile_coverage"
	CheckRepoSecretScanningPushProtection    Check = "repo_secret_scanning_push_protection_coverage"
	CheckRepoEnvOnDefaultBranch              Check = "repo_env_on_default_branch"
	CheckRepoDeployKeysWriteAccess           Check = "repo_deploy_keys_write_access"
	CheckRepoDependabotCriticalOpen          Check = "repo_dependabot_critical_open"
	CheckRepoSecretScanningAlertsOpen        Check = "repo_secret_scanning_alerts_open"
	CheckRepoCodeScanningCriticalOpen        Check = "repo_code_scanning_critical_open"
	CheckRepoCommitStatusCICoverage          Check = "repo_commit_status_ci_coverage"
	CheckRepoExternalCICoverage              Check = "repo_external_ci_coverage"
	CheckRepoCIProviders                     Check = "repo_ci_providers"
	CheckRepoSecurityContactCoverage         Check = "repo_security_contact_coverage"
	CheckRepoIncidentResponseDocCoverage     Check = "repo_incident_response_doc_coverage"
	CheckRepoIssueTemplatesCoverage          Check = "repo_issue_templates_coverage"
	CheckRepoDeFactoPRReviewCoverage         Check = "repo_de_facto_pr_review_coverage"
	CheckRepoPRApprovalRate                  Check = "repo_pr_approval_rate"
)

func newFact(check Check, scope string, value any, apiRef string) Fact {
	return Fact{
		Check:  check,
		Scope:  scope,
		Value:  value,
		APIRef: apiRef,
	}
}

func coverageFact(check Check, matched, total int, apiRef string) Fact {
	return newFact(check, "org", map[string]int{
		"matched": matched,
		"total":   total,
	}, apiRef)
}
