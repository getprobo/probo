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

package iam

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/iam/oauth2scope"
	"go.probo.inc/probo/pkg/iam/policy"
)

func TestValidateServiceAccountCredentialScopes_Subset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		accountScopes    coredata.OAuth2Scopes
		credentialScopes coredata.OAuth2Scopes
		wantError        bool
	}{
		{
			name:             "equal scopes",
			accountScopes:    coredata.OAuth2Scopes{"v1:org", "v1:iam"},
			credentialScopes: coredata.OAuth2Scopes{"v1:org", "v1:iam"},
		},
		{
			name:             "strict subset",
			accountScopes:    coredata.OAuth2Scopes{"v1:org", "v1:iam"},
			credentialScopes: coredata.OAuth2Scopes{"v1:org"},
		},
		{
			name:             "empty subset",
			accountScopes:    coredata.OAuth2Scopes{"v1:org"},
			credentialScopes: nil,
		},
		{
			name:             "scope outside maximum",
			accountScopes:    coredata.OAuth2Scopes{"v1:org"},
			credentialScopes: coredata.OAuth2Scopes{"v1:iam"},
			wantError:        true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				err := validateServiceAccountCredentialScopes(tt.accountScopes, tt.credentialScopes)
				if tt.wantError {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			},
		)
	}
}

func TestServiceAccountPolicies_HumanOwnerOnly(t *testing.T) {
	t.Parallel()

	actions := []string{
		ActionServiceAccountCreate,
		ActionServiceAccountGet,
		ActionServiceAccountList,
		ActionServiceAccountUpdate,
		ActionServiceAccountDisable,
		ActionServiceAccountDelete,
		ActionServiceAccountCredentialCreate,
		ActionServiceAccountCredentialList,
		ActionServiceAccountCredentialRevoke,
	}
	conditionContext := policy.ConditionContext{
		Principal: map[string]string{"organization_id": "org-1"},
		Resource:  map[string]string{"organization_id": "org-1"},
	}
	evaluator := policy.NewEvaluator()

	for _, action := range actions {
		ownerResult := evaluator.Evaluate(
			policy.AuthorizationRequest{
				Action:           action,
				ConditionContext: conditionContext,
			},
			[]*policy.Policy{IAMOwnerPolicy},
		)
		adminResult := evaluator.Evaluate(
			policy.AuthorizationRequest{
				Action:           action,
				ConditionContext: conditionContext,
			},
			[]*policy.Policy{IAMAdminPolicy},
		)

		assert.Equal(t, policy.DecisionAllow, ownerResult.Decision, action)
		assert.Equal(t, policy.DecisionNoMatch, adminResult.Decision, action)
	}
}

func TestServiceAccountOAuth2ScopeMappings_ReadAndWrite(t *testing.T) {
	t.Parallel()

	registry := oauth2scope.NewRegistry().Register(IAMOAuth2ScopeMappings)

	assert.True(t, registry.Allows(coredata.OAuth2Scopes{ScopeV1IAMRead}, ActionServiceAccountGet))
	assert.True(t, registry.Allows(coredata.OAuth2Scopes{ScopeV1IAMRead}, ActionServiceAccountCredentialList))
	assert.False(t, registry.Allows(coredata.OAuth2Scopes{ScopeV1IAMRead}, ActionServiceAccountCreate))
	assert.False(t, registry.Allows(coredata.OAuth2Scopes{ScopeV1IAMRead}, ActionServiceAccountCredentialRevoke))
	assert.True(t, registry.Allows(coredata.OAuth2Scopes{ScopeV1IAM}, ActionServiceAccountCreate))
	assert.True(t, registry.Allows(coredata.OAuth2Scopes{ScopeV1IAM}, ActionServiceAccountCredentialRevoke))
}
