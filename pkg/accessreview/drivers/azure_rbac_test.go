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

package drivers

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

func TestAzureRoleDefinitionGUID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		id   string
		want string
	}{
		{
			name: "subscription-scoped definition",
			id:   "/subscriptions/11111111-1111-4111-8111-111111111111/providers/Microsoft.Authorization/roleDefinitions/8e3af657-a8ff-443c-a75c-2fe8c4bcb635",
			want: azureRoleDefinitionOwner,
		},
		{
			name: "uppercase guid is lowercased",
			id:   "/providers/Microsoft.Authorization/roleDefinitions/8E3AF657-A8FF-443C-A75C-2FE8C4BCB635",
			want: azureRoleDefinitionOwner,
		},
		{
			name: "bare guid",
			id:   azureRoleDefinitionOwner,
			want: azureRoleDefinitionOwner,
		},
		{name: "empty", id: ""},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tt.want, azureRoleDefinitionGUID(tt.id))
			},
		)
	}
}

func TestAzureIsAdmin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ids  []string
		want *bool
	}{
		{name: "owner", ids: []string{azureRoleDefinitionOwner}, want: new(true)},
		{name: "user access administrator", ids: []string{azureRoleDefinitionUserAccessAdmin}, want: new(true)},
		{name: "rbac administrator", ids: []string{azureRoleDefinitionRBACAdmin}, want: new(true)},
		{name: "custom role", ids: []string{"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"}, want: nil},
		{name: "empty", ids: nil, want: nil},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tt.want, azureIsAdmin(tt.ids))
			},
		)
	}
}

func TestAzureIdentityRecord_PrincipalTypes(t *testing.T) {
	t.Parallel()

	enabled := false

	tests := []struct {
		name        string
		identity    azureIdentity
		accountType coredata.AccessReviewEntryAccountType
		authMethod  coredata.AccessReviewEntryAuthMethod
		active      *bool
		email       string
	}{
		{
			name: "user",
			identity: azureIdentity{
				PrincipalType:  azurePrincipalUser,
				Email:          "alice@probo-azure.test",
				AccountEnabled: new(true),
			},
			accountType: coredata.AccessReviewEntryAccountTypeUser,
			authMethod:  coredata.AccessReviewEntryAuthMethodSSO,
			active:      new(true),
			email:       "alice@probo-azure.test",
		},
		{
			name: "group",
			identity: azureIdentity{
				PrincipalType: azurePrincipalGroup,
				Email:         "eng@probo-azure.test",
			},
			accountType: coredata.AccessReviewEntryAccountTypeUser,
			authMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			email:       "eng@probo-azure.test",
		},
		{
			name: "disabled service principal",
			identity: azureIdentity{
				PrincipalType:  azurePrincipalServicePrincipal,
				AccountEnabled: &enabled,
			},
			accountType: coredata.AccessReviewEntryAccountTypeServiceAccount,
			authMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			active:      &enabled,
		},
		{
			name: "device",
			identity: azureIdentity{
				PrincipalType: azurePrincipalDevice,
			},
			accountType: coredata.AccessReviewEntryAccountTypeServiceAccount,
			authMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
		},
		{
			name: "unrecognised type uses default",
			identity: azureIdentity{
				PrincipalType: "SomethingNew",
			},
			accountType: coredata.AccessReviewEntryAccountTypeUser,
			authMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				record := azureIdentityRecord(tt.identity)
				assert.Equal(t, tt.accountType, record.AccountType)
				assert.Equal(t, tt.authMethod, record.AuthMethod)
				assert.Equal(t, tt.active, record.Active)
				assert.Equal(t, tt.email, record.Email)
				assert.Equal(t, coredata.MFAStatusUnknown, record.MFAStatus)
				assert.Nil(t, record.LastLogin)
				assert.Nil(t, record.CreatedAt)
			},
		)
	}
}

func TestFoldAzureAssignments_DedupesByPrincipal(t *testing.T) {
	t.Parallel()

	owner := "/subscriptions/sub/providers/Microsoft.Authorization/roleDefinitions/" + azureRoleDefinitionOwner
	reader := "/subscriptions/sub/providers/Microsoft.Authorization/roleDefinitions/acdd72a7-3385-48ef-bd42-f606fba81ae7"
	alice := "22222222-2222-4222-8222-222222222222"

	identities := foldAzureAssignments(
		[]armauthorization.RoleAssignment{
			azureTestAssignment(alice, azurePrincipalUser, owner),
			azureTestAssignment(alice, azurePrincipalUser, reader),
		},
		map[string]string{
			owner:  "Owner",
			reader: "Reader",
		},
	)

	require.Len(t, identities, 1)
	assert.Equal(t, alice, identities[0].PrincipalID)
	assert.Equal(t, []string{"Owner", "Reader"}, identities[0].Roles)
	require.NotNil(t, azureIsAdmin(identities[0].RoleDefinitionIDs))
	assert.True(t, *azureIsAdmin(identities[0].RoleDefinitionIDs))
}

func azureTestAssignment(
	principalID string,
	principalType armauthorization.PrincipalType,
	roleDefinitionID string,
) armauthorization.RoleAssignment {
	return armauthorization.RoleAssignment{
		Properties: &armauthorization.RoleAssignmentProperties{
			PrincipalID:      new(principalID),
			PrincipalType:    new(principalType),
			RoleDefinitionID: new(roleDefinitionID),
		},
	}
}
