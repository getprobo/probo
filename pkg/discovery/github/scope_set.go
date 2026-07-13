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
	"go.probo.inc/probo/pkg/connector"
)

type scopeSet map[string]struct{}

func newScopeSet(conn connector.Connection) scopeSet {
	set := scopeSet{}

	if conn == nil {
		return set
	}

	for _, scope := range conn.Scopes() {
		set[scope] = struct{}{}
	}

	return set
}

func (s scopeSet) has(scope string) bool {
	_, ok := s[scope]

	return ok
}

func (s scopeSet) hasRepoRead() bool {
	return s.has("repo") || s.has("public_repo")
}

func (s scopeSet) hasSecurityEvents() bool {
	return s.has("security_events")
}

func (s scopeSet) hasEnterpriseRead() bool {
	return s.has("read:enterprise")
}

func (s scopeSet) hasAuditLogRead() bool {
	return s.has("read:audit_log")
}
