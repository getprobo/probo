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

package accessreview_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/accessreview"
	"go.probo.inc/probo/pkg/iam/policy"
)

func TestPolicySet_AuditorCannotAccessReviewData(t *testing.T) {
	t.Parallel()

	policySet := accessreview.PolicySet()
	policies := append(
		policySet.RolePolicies["AUDITOR"],
		policy.NewPolicy(
			"test:access-review-allow",
			"Test Access Review Allow",
			policy.Allow("access-review:*"),
		),
	)
	evaluator := policy.NewEvaluator()

	actions := []string{
		accessreview.ActionCampaignGet,
		accessreview.ActionCampaignList,
		accessreview.ActionCampaignCreate,
		accessreview.ActionCampaignUpdate,
		accessreview.ActionCampaignDelete,
		accessreview.ActionCampaignStart,
		accessreview.ActionCampaignClose,
		accessreview.ActionCampaignCancel,
		accessreview.ActionCampaignAddSource,
		accessreview.ActionCampaignRemoveSource,
		accessreview.ActionEntryGet,
		accessreview.ActionEntryList,
		accessreview.ActionEntryDecide,
		accessreview.ActionEntryFlag,
		accessreview.ActionSourceGet,
		accessreview.ActionSourceList,
		accessreview.ActionSourceCreate,
		accessreview.ActionSourceUpdate,
		accessreview.ActionSourceDelete,
		accessreview.ActionSourceSync,
	}

	for _, action := range actions {
		action := action
		t.Run(
			action,
			func(t *testing.T) {
				t.Parallel()

				result := evaluator.Evaluate(
					policy.AuthorizationRequest{Action: action},
					policies,
				)

				assert.Equal(t, policy.DecisionDeny, result.Decision)
			},
		)
	}
}

func TestPolicySet_AuditorCanListDriverCatalog(t *testing.T) {
	t.Parallel()

	policySet := accessreview.PolicySet()
	policies := append(
		policySet.RolePolicies["AUDITOR"],
		policySet.IdentityScopedPolicies...,
	)
	evaluator := policy.NewEvaluator()

	result := evaluator.Evaluate(
		policy.AuthorizationRequest{Action: accessreview.ActionDriverCatalogList},
		policies,
	)

	assert.Equal(t, policy.DecisionAllow, result.Decision)
}
