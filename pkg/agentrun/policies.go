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
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/iam/policy"
)

var organizationCondition = policy.Equals("principal.organization_id", "resource.organization_id")

// FullAccessPolicy grants complete agent-run access, including approval
// decisions, to organization owners and admins.
var FullAccessPolicy = policy.NewPolicy(
	"agentrun:full-access",
	"Agent Run Full Access",
	policy.Allow(
		ActionAgentRunGet,
		ActionAgentRunList,
		ActionAgentRunApprove,
	).WithSID("agent-run-full-access").When(organizationCondition),
).WithDescription("Full agent-run access including approval decisions")

// ReadAccessPolicy grants read-only agent-run access to viewers and auditors.
var ReadAccessPolicy = policy.NewPolicy(
	"agentrun:read-access",
	"Agent Run Read Access",
	policy.Allow(
		ActionAgentRunGet,
		ActionAgentRunList,
	).WithSID("agent-run-read-access").When(organizationCondition),
).WithDescription("Read-only agent-run access")

// PolicySet returns the PolicySet for the agent-run service. It is owned by
// this package and registered into the authorizer at composition time so the
// agent-run authorization rules live alongside the agent-run domain logic
// instead of in the core probo policy set.
func PolicySet() *iam.PolicySet {
	return iam.NewPolicySet().
		AddRolePolicy("OWNER", FullAccessPolicy).
		AddRolePolicy("ADMIN", FullAccessPolicy).
		AddRolePolicy("VIEWER", ReadAccessPolicy).
		AddRolePolicy("AUDITOR", ReadAccessPolicy)
}
