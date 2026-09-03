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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	cloudresourcemanager "google.golang.org/api/cloudresourcemanager/v1"
)

func TestParseGCPPrincipal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		member    string
		wantOK    bool
		wantKind  gcpPrincipalKind
		wantEmail string
	}{
		{
			name:      "user",
			member:    "user:alice@example.com",
			wantOK:    true,
			wantKind:  gcpPrincipalUser,
			wantEmail: "alice@example.com",
		},
		{
			name:      "group",
			member:    "group:eng@example.com",
			wantOK:    true,
			wantKind:  gcpPrincipalGroup,
			wantEmail: "eng@example.com",
		},
		{
			name:      "service account",
			member:    "serviceAccount:ci@my-project.iam.gserviceaccount.com",
			wantOK:    true,
			wantKind:  gcpPrincipalServiceAccount,
			wantEmail: "ci@my-project.iam.gserviceaccount.com",
		},
		{name: "allUsers", member: "allUsers"},
		{name: "allAuthenticatedUsers", member: "allAuthenticatedUsers"},
		{name: "domain", member: "domain:example.com"},
		{name: "deleted user", member: "deleted:user:alice@example.com?uid=123"},
		{name: "principal", member: "principal://iam.googleapis.com/locations/global/workforcePools/p/subject/s"},
		{name: "principalSet", member: "principalSet://iam.googleapis.com/locations/global/workforcePools/p/*"},
		{name: "empty user", member: "user:"},
		{name: "unknown kind", member: "folder:123"},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				got, ok := parseGCPPrincipal(tt.member)
				assert.Equal(t, tt.wantOK, ok)

				if !tt.wantOK {
					return
				}

				assert.Equal(t, tt.member, got.Principal)
				assert.Equal(t, tt.wantKind, got.Kind)
				assert.Equal(t, tt.wantEmail, got.Email)
			},
		)
	}
}

func TestGCPIsAdmin(t *testing.T) {
	t.Parallel()

	require.NotNil(t, gcpIsAdmin([]string{"roles/viewer", gcpRoleOwner}))
	assert.True(t, *gcpIsAdmin([]string{gcpRoleOwner}))
	assert.Nil(t, gcpIsAdmin([]string{"roles/editor"}))
	assert.Nil(t, gcpIsAdmin([]string{"projects/my-project/roles/customAdmin"}))
	assert.Nil(t, gcpIsAdmin(nil))
}

func TestGCPIdentityRecord_MapsOwnerUser(t *testing.T) {
	t.Parallel()

	record := gcpIdentityRecord(
		gcpIdentity{
			Principal: "user:alice@example.com",
			Email:     "alice@example.com",
			Kind:      gcpPrincipalUser,
			Roles:     []string{gcpRoleOwner},
		},
	)

	assert.Equal(t, "alice@example.com", record.Email)
	assert.Equal(t, []string{gcpRoleOwner}, record.Roles)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, record.AccountType)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodSSO, record.AuthMethod)
	assert.Equal(t, coredata.MFAStatusUnknown, record.MFAStatus)
	assert.Equal(t, "user:alice@example.com", record.ExternalID)
	require.NotNil(t, record.IsAdmin)
	assert.True(t, *record.IsAdmin)
	assert.Nil(t, record.Active)
	assert.Nil(t, record.LastLogin)
}

func TestGCPIdentityRecord_MapsDisabledServiceAccount(t *testing.T) {
	t.Parallel()

	record := gcpIdentityRecord(
		gcpIdentity{
			Principal:         "serviceAccount:bot@my-project.iam.gserviceaccount.com",
			Email:             "bot@my-project.iam.gserviceaccount.com",
			Kind:              gcpPrincipalServiceAccount,
			Roles:             []string{"roles/editor"},
			UniqueID:          "333",
			Disabled:          new(true),
			HasUserManagedKey: true,
		},
	)

	assert.Equal(t, coredata.AccessReviewEntryAccountTypeServiceAccount, record.AccountType)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodAPIKey, record.AuthMethod)
	assert.Equal(t, "333", record.ExternalID)
	assert.Nil(t, record.IsAdmin)
	require.NotNil(t, record.Active)
	assert.False(t, *record.Active)
}

func TestPrincipalsFromPolicy_SkipsAndUnionsRoles(t *testing.T) {
	t.Parallel()

	principals := principalsFromPolicy(
		&cloudresourcemanager.Policy{
			Bindings: []*cloudresourcemanager.Binding{
				{
					Role: gcpRoleOwner,
					Members: []string{
						"user:alice@example.com",
						"allUsers",
						"domain:example.com",
					},
				},
				{
					Role: "roles/viewer",
					Members: []string{
						"user:alice@example.com",
						"group:eng@example.com",
					},
				},
			},
		},
	)

	byPrincipal := make(map[string]gcpIdentity, len(principals))
	for _, principal := range principals {
		byPrincipal[principal.Principal] = principal
	}

	require.Len(t, principals, 2)
	assert.ElementsMatch(
		t,
		[]string{gcpRoleOwner, "roles/viewer"},
		byPrincipal["user:alice@example.com"].Roles,
	)
	assert.Equal(t, []string{"roles/viewer"}, byPrincipal["group:eng@example.com"].Roles)
}

func TestUnionGCPIdentities_IncludesUnboundServiceAccount(t *testing.T) {
	t.Parallel()

	identities := unionGCPIdentities(
		[]gcpIdentity{
			{
				Principal: "user:alice@example.com",
				Email:     "alice@example.com",
				Kind:      gcpPrincipalUser,
				Roles:     []string{gcpRoleOwner},
			},
			{
				Principal: "serviceAccount:ci@my-project.iam.gserviceaccount.com",
				Email:     "ci@my-project.iam.gserviceaccount.com",
				Kind:      gcpPrincipalServiceAccount,
				Roles:     []string{"roles/editor"},
			},
		},
		[]gcpServiceAccount{
			{
				Email:       "ci@my-project.iam.gserviceaccount.com",
				UniqueID:    "111",
				DisplayName: "CI",
			},
			{
				Email:    "unbound@my-project.iam.gserviceaccount.com",
				UniqueID: "222",
				Disabled: true,
			},
		},
	)

	require.Len(t, identities, 3)

	var ci, unbound gcpIdentity

	for _, identity := range identities {
		switch identity.Email {
		case "ci@my-project.iam.gserviceaccount.com":
			ci = identity
		case "unbound@my-project.iam.gserviceaccount.com":
			unbound = identity
		}
	}

	assert.Equal(t, "111", ci.UniqueID)
	assert.Equal(t, "CI", ci.DisplayName)
	assert.Equal(t, []string{"roles/editor"}, ci.Roles)
	require.NotNil(t, ci.Disabled)
	assert.False(t, *ci.Disabled)

	assert.Equal(t, "222", unbound.UniqueID)
	assert.Empty(t, unbound.Roles)
	require.NotNil(t, unbound.Disabled)
	assert.True(t, *unbound.Disabled)
}

func TestProjectIDFromServiceAccountEmail(t *testing.T) {
	t.Parallel()

	assert.Equal(
		t,
		"my-project",
		projectIDFromServiceAccountEmail("probo-audit@my-project.iam.gserviceaccount.com"),
	)
	assert.Equal(
		t,
		"my-project",
		projectIDFromServiceAccountEmail("probo-audit@my-project.s3ns.iam.gserviceaccount.com"),
	)
	assert.Empty(t, projectIDFromServiceAccountEmail("alice@example.com"))
	assert.Empty(t, projectIDFromServiceAccountEmail(""))
}
