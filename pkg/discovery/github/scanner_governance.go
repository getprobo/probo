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
	"net/url"
	"strings"
)

type (
	outsideCollaborator struct {
		Login string `json:"login"`
	}

	orgActionsPermissions struct {
		EnabledRepositories          string `json:"enabled_repositories"`
		AllowedActions               string `json:"allowed_actions"`
		SHAPinningRequired           bool   `json:"sha_pinning_required"`
		CanApprovePullRequestReviews bool   `json:"can_approve_pull_request_reviews"`
	}

	orgInstallation struct {
		ID int64 `json:"id"`
	}

	orgInstallationsPage struct {
		TotalCount    int               `json:"total_count"`
		Installations []orgInstallation `json:"installations"`
	}
)

func (s *discoveryScanner) scanGovernance(ctx context.Context, sheet *FactSheet) {
	s.scanOutsideCollaborators(ctx, sheet)
	s.scanActionsPermissions(ctx, sheet)
	s.scanGitHubApps(ctx, sheet)
	s.scanAuditLogAccess(ctx, sheet)
	s.scanEnterpriseAccess(ctx, sheet)
}

func (s *discoveryScanner) scanOutsideCollaborators(ctx context.Context, sheet *FactSheet) {
	endpoint, err := s.api.orgEndpoint(s.org, "outside_collaborators")
	if err != nil {
		sheet.Limitations = append(sheet.Limitations, "cannot build outside collaborators URL")

		return
	}

	endpoint, err = withPerPage(endpoint, 100)
	if err != nil {
		sheet.Limitations = append(sheet.Limitations, "cannot build outside collaborators URL")

		return
	}

	var collaborators []outsideCollaborator

	if _, err := s.api.getPaginated(ctx, endpoint, &collaborators); err != nil {
		sheet.Limitations = append(sheet.Limitations, "cannot list outside collaborators")

		return
	}

	sheet.Facts = append(sheet.Facts, Fact{
		FactID:  "f-outside-collaborators",
		FactKey: "org_outside_collaborators",
		Scope:   "org",
		Value: map[string]int{
			"count": len(collaborators),
		},
		APIRef: "GET /orgs/{org}/outside_collaborators",
	})
}

func (s *discoveryScanner) scanActionsPermissions(ctx context.Context, sheet *FactSheet) {
	endpoint, err := s.api.orgEndpoint(s.org, "actions", "permissions")
	if err != nil {
		sheet.Limitations = append(sheet.Limitations, "cannot build actions permissions URL")

		return
	}

	var perms orgActionsPermissions

	if _, err := s.api.getJSON(ctx, endpoint, &perms); err != nil {
		sheet.Limitations = append(sheet.Limitations, "cannot read org actions permissions")

		return
	}

	restricted := strings.EqualFold(perms.AllowedActions, "selected") ||
		strings.EqualFold(perms.AllowedActions, "local_only")

	sheet.Facts = append(sheet.Facts, Fact{
		FactID:  "f-actions-restricted",
		FactKey: "org_actions_restricted",
		Scope:   "org",
		Value: map[string]any{
			"restricted":                       restricted,
			"allowed_actions":                  perms.AllowedActions,
			"enabled_repositories":             perms.EnabledRepositories,
			"sha_pinning_required":             perms.SHAPinningRequired,
			"can_approve_pull_request_reviews": perms.CanApprovePullRequestReviews,
		},
		APIRef: "GET /orgs/{org}/actions/permissions",
	})

	sheet.Facts = append(sheet.Facts, Fact{
		FactID:  "f-fork-pr-approval",
		FactKey: "org_fork_pr_approval_required",
		Scope:   "org",
		Value:   !perms.CanApprovePullRequestReviews,
		APIRef:  "GET /orgs/{org}/actions/permissions",
	})
}

func (s *discoveryScanner) scanGitHubApps(ctx context.Context, sheet *FactSheet) {
	endpoint, err := s.api.orgEndpoint(s.org, "installations")
	if err != nil {
		sheet.Limitations = append(sheet.Limitations, "cannot build github app installations URL")

		return
	}

	var page orgInstallationsPage

	if _, err := s.api.getJSON(ctx, endpoint, &page); err != nil {
		sheet.Limitations = append(sheet.Limitations, "cannot list github app installations")

		return
	}

	sheet.Facts = append(sheet.Facts, Fact{
		FactID:  "f-github-apps",
		FactKey: "org_github_apps",
		Scope:   "org",
		Value: map[string]int{
			"installations": page.TotalCount,
		},
		APIRef: "GET /orgs/{org}/installations",
	})
}

func (s *discoveryScanner) scanAuditLogAccess(ctx context.Context, sheet *FactSheet) {
	if !s.scopes.hasAuditLogRead() {
		sheet.Limitations = append(
			sheet.Limitations,
			"read:audit_log scope not granted; skipping audit log probe",
		)

		return
	}

	endpoint, err := url.JoinPath(githubAPIBase, "organizations", url.PathEscape(s.org), "audit-log")
	if err != nil {
		sheet.Limitations = append(sheet.Limitations, "cannot build audit log URL")

		return
	}

	endpoint, err = withPerPage(endpoint, 1)
	if err != nil {
		sheet.Limitations = append(sheet.Limitations, "cannot build audit log URL")

		return
	}

	var events []map[string]any

	if _, err := s.api.getJSON(ctx, endpoint, &events); err != nil {
		sheet.Facts = append(sheet.Facts, Fact{
			FactID:  "f-audit-log",
			FactKey: "org_audit_log_accessible",
			Scope:   "org",
			Value:   false,
			APIRef:  "GET /organizations/{org}/audit-log",
		})

		return
	}

	sheet.Facts = append(sheet.Facts, Fact{
		FactID:  "f-audit-log",
		FactKey: "org_audit_log_accessible",
		Scope:   "org",
		Value:   true,
		APIRef:  "GET /organizations/{org}/audit-log",
	})
}

func (s *discoveryScanner) scanEnterpriseAccess(ctx context.Context, sheet *FactSheet) {
	if !s.scopes.hasEnterpriseRead() {
		sheet.Limitations = append(
			sheet.Limitations,
			"read:enterprise scope not granted; skipping enterprise probe",
		)

		return
	}

	endpoint, err := url.JoinPath(githubAPIBase, "enterprise")
	if err != nil {
		sheet.Limitations = append(sheet.Limitations, "cannot build enterprise URL")

		return
	}

	var enterprises []map[string]any

	if _, err := s.api.getJSON(ctx, endpoint, &enterprises); err != nil {
		sheet.Facts = append(sheet.Facts, Fact{
			FactID:  "f-enterprise",
			FactKey: "org_enterprise_accessible",
			Scope:   "org",
			Value:   false,
			APIRef:  "GET /enterprise",
		})

		return
	}

	sheet.Facts = append(sheet.Facts, Fact{
		FactID:  "f-enterprise",
		FactKey: "org_enterprise_accessible",
		Scope:   "org",
		Value:   len(enterprises) > 0,
		APIRef:  "GET /enterprise",
	})
}
