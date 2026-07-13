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

	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/coredata"
)

type githubOrganization struct {
	TwoFactorRequirementEnabled        *bool  `json:"two_factor_requirement_enabled"`
	DefaultRepositoryPermission        string `json:"default_repository_permission"`
	MembersCanCreatePublicRepositories *bool  `json:"members_can_create_public_repositories"`
	MembersCanChangeRepoVisibility     *bool  `json:"members_can_change_repo_visibility"`
}

func (s *discoveryScanner) scanOrgSettings(ctx context.Context, sheet *FactSheet) error {
	org, err := s.fetchOrganization(ctx)
	if err != nil {
		return err
	}

	sheet.Facts = append(sheet.Facts, orgFacts(org)...)

	disabled2FA, err := s.count2FADisabled(ctx)
	if err != nil {
		sheet.Limitations = append(sheet.Limitations, "cannot list 2FA-disabled members")
	} else {
		sheet.Facts = append(sheet.Facts, Fact{
			FactID:  "f-members-without-2fa",
			FactKey: "org_no_2fa_members",
			Scope:   "org",
			Value:   disabled2FA,
			APIRef:  "GET /orgs/{org}/members?filter=2fa_disabled",
		})
	}

	adminCount, memberCount, err := s.countAdmins(ctx)
	if err != nil {
		sheet.Limitations = append(sheet.Limitations, "cannot count org admins")
	} else {
		sheet.Facts = append(sheet.Facts, Fact{
			FactID:  "f-admin-ratio",
			FactKey: "org_admin_minimization",
			Scope:   "org",
			Value: map[string]int{
				"admins":  adminCount,
				"members": memberCount,
			},
			APIRef: "GET /orgs/{org}/members + memberships",
		})
	}

	return nil
}

func orgFacts(org *githubOrganization) []Fact {
	facts := make([]Fact, 0, 4)

	if org.TwoFactorRequirementEnabled != nil {
		facts = append(facts, Fact{
			FactID:  "f-org-mfa-required",
			FactKey: "org_mfa_required",
			Scope:   "org",
			Value:   *org.TwoFactorRequirementEnabled,
			APIRef:  "GET /orgs/{org}",
		})
	}

	if org.DefaultRepositoryPermission != "" {
		facts = append(facts, Fact{
			FactID:  "f-base-permissions",
			FactKey: "org_base_permissions",
			Scope:   "org",
			Value:   org.DefaultRepositoryPermission,
			APIRef:  "GET /orgs/{org}",
		})
	}

	if org.MembersCanCreatePublicRepositories != nil {
		facts = append(facts, Fact{
			FactID:  "f-no-public-repo-creation",
			FactKey: "org_no_public_repo_creation",
			Scope:   "org",
			Value:   !*org.MembersCanCreatePublicRepositories,
			APIRef:  "GET /orgs/{org}",
		})
	}

	if org.MembersCanChangeRepoVisibility != nil {
		facts = append(facts, Fact{
			FactID:  "f-no-visibility-change",
			FactKey: "org_no_visibility_change",
			Scope:   "org",
			Value:   !*org.MembersCanChangeRepoVisibility,
			APIRef:  "GET /orgs/{org}",
		})
	}

	return facts
}

func (s *discoveryScanner) fetchOrganization(ctx context.Context) (*githubOrganization, error) {
	return s.api.getOrganization(ctx, s.org)
}

func (s *discoveryScanner) count2FADisabled(ctx context.Context) (int, error) {
	driver := drivers.NewGitHubDriver(s.api.HTTPClient(), s.org, s.logger)

	records, err := driver.ListAccounts(ctx)
	if err != nil {
		return 0, fmt.Errorf("cannot list github accounts for 2fa check: %w", err)
	}

	count := 0

	for _, record := range records {
		if record.MFAStatus == coredata.MFAStatusDisabled {
			count++
		}
	}

	return count, nil
}

func (s *discoveryScanner) countAdmins(ctx context.Context) (int, int, error) {
	driver := drivers.NewGitHubDriver(s.api.HTTPClient(), s.org, s.logger)

	records, err := driver.ListAccounts(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot list github accounts for admin count: %w", err)
	}

	admins := 0

	for _, record := range records {
		if record.IsAdmin {
			admins++
		}
	}

	return admins, len(records), nil
}

// evaluateAdminMinimization returns measure state for admin ratio fact.
func evaluateAdminMinimization(value any) coredata.MeasureState {
	m, ok := value.(map[string]any)
	if !ok {
		return coredata.MeasureStateUnknown
	}

	admins, _ := toInt(m["admins"])
	members, _ := toInt(m["members"])

	if members == 0 {
		return coredata.MeasureStateUnknown
	}

	if admins <= 3 {
		return coredata.MeasureStateImplemented
	}

	if float64(admins)/float64(members) <= 0.15 {
		return coredata.MeasureStateImplemented
	}

	return coredata.MeasureStateNotImplemented
}

func toInt(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	default:
		return 0, false
	}
}
