// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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

// DriverCatalogPolicy grants every authenticated identity read access to the
// global access-review driver catalog. The catalog is deployment-scoped and has
// no organization scoping, so the allow has no condition.
var DriverCatalogPolicy = policy.NewPolicy(
	"access-review:driver-catalog",
	"Access Review Driver Catalog",
	policy.Allow(
		ActionDriverCatalogList,
	).WithSID("read-access-review-driver-catalog"),
).WithDescription("Allows every authenticated user to read the global access-review driver catalog")

// PolicySet returns the PolicySet for the access-review service. It is owned by
// this package and registered into the authorizer at composition time so the
// access-review authorization rules live alongside the access-review domain
// logic instead of in the core probo policy set.
func PolicySet() *iam.PolicySet {
	return iam.NewPolicySet().
		AddRolePolicy("OWNER", FullAccessPolicy).
		AddRolePolicy("ADMIN", FullAccessPolicy).
		AddRolePolicy("VIEWER", ReadAccessPolicy).
		AddIdentityScopedPolicy(DriverCatalogPolicy)
}
