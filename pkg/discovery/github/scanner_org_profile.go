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

	"go.probo.inc/probo/pkg/discovery/vfs"
)

func (s *discoveryScanner) scanOrgProfile(ctx context.Context, sheet *FactSheet) {
	if !s.scopes.hasRepoRead() {
		sheet.Limitations = append(
			sheet.Limitations,
			"repo scope not granted; skipping org profile repository scan",
		)

		return
	}

	repo := vfs.Repo{Owner: s.org, Name: ".github"}
	repoFS := s.orgFS.Open(repo)

	endpoint, err := s.api.repoEndpoint(s.org, ".github")
	if err != nil {
		sheet.Limitations = append(sheet.Limitations, "cannot build org profile repository URL")

		return
	}

	var meta repoListItem

	if _, err := s.api.getJSON(ctx, endpoint, &meta); err != nil {
		sheet.Limitations = append(sheet.Limitations, "org profile repository .github not accessible")

		return
	}

	hasSecurity, _ := repoFS.Exists(ctx, "SECURITY.md")
	hasContributing, _ := repoFS.Exists(ctx, "CONTRIBUTING.md")

	securityContact := false

	if content, ok := s.readRepoFile(ctx, repoFS, "SECURITY.md"); ok {
		securityContact = securityContactInMarkdown(string(content))
	}

	sheet.Facts = append(sheet.Facts, Fact{
		FactID:  "f-org-profile-security-md",
		FactKey: "org_profile_security_md",
		Scope:   "org",
		Value: map[string]any{
			"present":          hasSecurity,
			"security_contact": securityContact,
		},
		APIRef: "GET /repos/{org}/.github/contents/SECURITY.md",
	})

	sheet.Facts = append(sheet.Facts, Fact{
		FactID:  "f-org-profile-contributing",
		FactKey: "org_profile_contributing_md",
		Scope:   "org",
		Value:   hasContributing,
		APIRef:  "GET /repos/{org}/.github/contents/CONTRIBUTING.md",
	})
}
