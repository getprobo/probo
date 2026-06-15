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

package agentrun

import (
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/iam"
)

const (
	ScopeV1AgentRead coredata.OAuth2Scope = "v1:agent:read"
	ScopeV1Agent     coredata.OAuth2Scope = "v1:agent"
)

// OAuth2ScopeSet returns OAuth2 scope mappings for agent-run actions.
func OAuth2ScopeSet() *iam.ScopeSet {
	return iam.CreateScopeSet(
		map[coredata.OAuth2Scope][]iam.Action{
			ScopeV1AgentRead: {
				ActionAgentRunGet,
				ActionAgentRunList,
			},
			ScopeV1Agent: {
				ActionAgentRunApprove,
			},
		},
	)
}
