// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package accessreview

import (
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/iam/policy"
)

var organizationCondition = policy.Equals("principal.organization_id", "resource.organization_id")

// FullAccessPolicy grants complete access-review access, including campaign,
// entry, and source management, to organization owners and admins.
var FullAccessPolicy = policy.NewPolicy(
	"access-review:full-access",
	"Access Review Full Access",
	policy.Allow(
		ActionCampaignGet, ActionCampaignList, ActionCampaignCreate,
		ActionCampaignUpdate, ActionCampaignDelete, ActionCampaignStart,
		ActionCampaignClose, ActionCampaignCancel, ActionCampaignAddSource,
		ActionCampaignRemoveSource,
		ActionEntryGet, ActionEntryList, ActionEntryDecide, ActionEntryFlag,
		ActionSourceGet, ActionSourceList, ActionSourceCreate,
		ActionSourceUpdate, ActionSourceDelete, ActionSourceSync,
	).WithSID("access-review-full-access").When(organizationCondition),
).WithDescription("Full access-review access including campaign, entry, and source management")

// ReadAccessPolicy grants read-only access-review access to viewers.
var ReadAccessPolicy = policy.NewPolicy(
	"access-review:read-access",
	"Access Review Read Access",
	policy.Allow(
		ActionCampaignGet, ActionCampaignList,
		ActionEntryGet, ActionEntryList,
		ActionSourceGet, ActionSourceList,
	).WithSID("access-review-read-access").When(organizationCondition),
).WithDescription("Read-only access-review access")

// PolicySet returns the PolicySet for the access-review service. It is owned by
// this package and registered into the authorizer at composition time so the
// access-review authorization rules live alongside the access-review domain
// logic instead of in the core probo policy set.
func PolicySet() *iam.PolicySet {
	return iam.NewPolicySet().
		AddRolePolicy("OWNER", FullAccessPolicy).
		AddRolePolicy("ADMIN", FullAccessPolicy).
		AddRolePolicy("VIEWER", ReadAccessPolicy)
}
