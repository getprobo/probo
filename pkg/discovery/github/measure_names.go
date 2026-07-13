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

// Measure titles shared by scanners, synthesis, and measure planning.
const (
	MeasureOrgMFARequired                      = "Org-wide MFA enforcement"
	MeasureOrgNo2FAMembers                     = "Members without 2FA"
	MeasureOrgBasePermissions                  = "Minimal default repository permissions"
	MeasureOrgNoPublicRepoCreation             = "Restrict public repository creation"
	MeasureOrgAdminMinimization                = "Admin account minimization"
	MeasureOrgPublicRepos                      = "Public repository exposure"
	MeasureOrgNoVisibilityChange               = "Restrict repository visibility changes"
	MeasureOrgOutsideCollaborators             = "Outside collaborator inventory"
	MeasureOrgActionsRestricted                = "GitHub Actions usage restricted"
	MeasureOrgGitHubApps                       = "GitHub App inventory"
	MeasureOrgAuditLogAccessible               = "Audit log accessible"
	MeasureRepoBranchProtectionCoverage        = "Default branch protection"
	MeasureRepoPRReviewsRequiredCoverage       = "Pull request reviews required"
	MeasureRepoSignedCommitsRequiredCoverage   = "Signed commits required"
	MeasureRepoWorkflowCoverage                = "CI/CD workflows present"
	MeasureRepoSecurityMDCoverage              = "Security disclosure policy"
	MeasureRepoContributingMDCoverage          = "Contributing guidelines documented"
	MeasureRepoDependabotConfigCoverage        = "Dependabot configuration"
	MeasureRepoDependabotCriticalOpen          = "Critical Dependabot alerts resolved"
	MeasureRepoSecretScanningAlertsOpen        = "Secret scanning alerts resolved"
	MeasureRepoCodeScanningCriticalOpen        = "Critical code scanning alerts resolved"
	MeasureOrgForkPRApprovalRequired           = "Fork pull request approval required"
	MeasureOrgEnterpriseAccessible             = "Enterprise settings accessible"
	MeasureRepoProductionClassification        = "Production repository classification"
	MeasureRepoSignedCommitsPracticeCoverage   = "Signed commits in practice"
	MeasureRepoForcePushDisabledCoverage       = "Force push disabled on default branch"
	MeasureRepoRequiredStatusChecksCoverage    = "Required status checks on default branch"
	MeasureRepoBypassActorRestrictionsCoverage = "Branch protection bypass restrictions"
	MeasureRepoPRCICoverage                    = "CI runs on pull requests"
	MeasureRepoPullRequestTargetRisk           = "No pull_request_target workflow risk"
	MeasureRepoCodeQLEnabledCoverage           = "CodeQL analysis in CI"
	MeasureRepoCodeQLDefaultSetupCoverage      = "CodeQL default setup enabled"
	MeasureRepoDependencyReviewCoverage        = "Dependency review in CI"
	MeasureRepoSASTInCICoverage                = "SAST in CI"
	MeasureRepoDepScanInCICoverage             = "Dependency scanning in CI"
	MeasureRepoDevelopmentGuideCoverage        = "Development guide documented"
	MeasureRepoCodeReviewGuideCoverage         = "Code review guide documented"
	MeasureRepoRenovateConfigCoverage          = "Renovate dependency automation"
	MeasureRepoLockfileCoverage                = "Dependency lock files maintained"
	MeasureRepoSecretScanningPushProtection    = "Secret scanning push protection"
	MeasureRepoEnvOnDefaultBranch              = "No secrets committed to default branch"
	MeasureRepoDeployKeysWriteAccess           = "Deploy keys with write access controlled"
	MeasureRepoCommitStatusCICoverage          = "CI detected via commit statuses"
	MeasureRepoExternalCICoverage              = "External CI providers detected"
	MeasureRepoCIProviders                     = "CI provider inventory"
	MeasureRepoSecurityContactCoverage         = "Security contact in SECURITY.md"
	MeasureRepoIncidentResponseDocCoverage     = "Incident response documentation"
	MeasureRepoIssueTemplatesCoverage          = "Issue templates configured"
	MeasureRepoDeFactoPRReviewCoverage         = "Pull requests reviewed in practice"
	MeasureRepoPRApprovalRate                  = "Pull request approval rate"
	MeasureOrgProfileSecurityMD                = "Organization security disclosure policy"
	MeasureOrgProfileContributingMD            = "Organization contributing guidelines"
)
