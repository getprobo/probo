// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package task

import (
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/iam/policy"
)

var (
	organizationCondition = policy.Equals("principal.organization_id", "resource.organization_id")
	ownerCondition        = policy.Equals("principal.id", "resource.owner_id")
)

var readActions = []string{
	ActionTaskGet, ActionTaskList,
	ActionTaskCommentGet, ActionTaskCommentList,
	ActionTaskActivityGet, ActionTaskActivityList,
}

var writeActions = []string{
	ActionTaskCreate, ActionTaskUpdate, ActionTaskDelete,
	ActionTaskAssign, ActionTaskUnassign,
	ActionTaskCommentCreate, ActionTaskCommentUpdate, ActionTaskCommentDelete,
}

// FullAccessPolicy grants complete task access to organization owners and
// admins. Owner and admin already match core:* on the probo role policy;
// this set keeps task authorization next to the domain after extraction.
var FullAccessPolicy = policy.NewPolicy(
	"task:full-access",
	"Task Full Access",
	policy.Allow(append(append([]string{}, readActions...), writeActions...)...).
		WithSID("task-full-access").
		When(organizationCondition),
).WithDescription("Full task, comment, and activity access")

// ReadAccessPolicy grants read-only task, comment, and activity access to
// organization viewers.
var ReadAccessPolicy = policy.NewPolicy(
	"task:read-access",
	"Task Read Access",
	policy.Allow(readActions...).
		WithSID("task-read-access").
		When(organizationCondition),
).WithDescription("Read-only task, comment, and activity access")

// TaskCommentOwnershipPolicy is attached to owner, admin, and viewer so
// permission(action:) on a comment matches mutation authorization: only the
// author can update; the author can delete; owner/admin still delete any
// comment through core:* on their role policy.
var TaskCommentOwnershipPolicy = policy.NewPolicy(
	"probo:task-comment-ownership",
	"Task Comment Ownership",
	policy.Deny(ActionTaskCommentUpdate).
		WithSID("deny-update-others-task-comments").
		When(policy.NotEquals("principal.id", "resource.owner_id")),
	policy.Allow(ActionTaskCommentUpdate, ActionTaskCommentDelete).
		WithSID("manage-own-task-comment").
		When(organizationCondition, ownerCondition),
).WithDescription("Authors can update and delete their own task comments; nobody else can update them")

// PolicySet returns the PolicySet for the task service. It is owned by this
// package and registered into the authorizer at composition time so task
// authorization rules live alongside the domain logic instead of in the
// core probo policy set.
func PolicySet() *iam.PolicySet {
	return iam.NewPolicySet().
		AddRolePolicy("OWNER", FullAccessPolicy, TaskCommentOwnershipPolicy).
		AddRolePolicy("ADMIN", FullAccessPolicy, TaskCommentOwnershipPolicy).
		AddRolePolicy("VIEWER", ReadAccessPolicy, TaskCommentOwnershipPolicy)
}
