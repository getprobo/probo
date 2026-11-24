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
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/accessreview"
	"go.probo.inc/probo/pkg/accessreview/drivers"
)

func TestGrantTriplesFromAccount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		account drivers.AccountRecord
		want    []accessreview.GrantTriple
	}{
		{
			name: "nil roles defaults to member",
			account: drivers.AccountRecord{
				Email: "  Jane@Example.com ",
			},
			want: []accessreview.GrantTriple{
				{
					PrincipalKey:           "jane@example.com",
					ResourceKey:            accessreview.DefaultResourceKey,
					ResourceKind:           accessreview.DefaultResourceKind,
					EntitlementKey:         accessreview.ImplicitMemberEntitlementKey,
					EntitlementDisplayName: accessreview.ImplicitMemberEntitlementDisplayName,
				},
			},
		},
		{
			name: "empty roles defaults to member",
			account: drivers.AccountRecord{
				Email:      "user@example.com",
				ExternalID: "ext-1",
				Roles:      []string{},
			},
			want: []accessreview.GrantTriple{
				{
					PrincipalKey:           "user@example.com|ext-1",
					ResourceKey:            accessreview.DefaultResourceKey,
					ResourceKind:           accessreview.DefaultResourceKind,
					EntitlementKey:         accessreview.ImplicitMemberEntitlementKey,
					EntitlementDisplayName: accessreview.ImplicitMemberEntitlementDisplayName,
				},
			},
		},
		{
			name: "whitespace-only roles defaults to member",
			account: drivers.AccountRecord{
				Email: "user@example.com",
				Roles: []string{"", "  ", "\t"},
			},
			want: []accessreview.GrantTriple{
				{
					PrincipalKey:           "user@example.com",
					ResourceKey:            accessreview.DefaultResourceKey,
					ResourceKind:           accessreview.DefaultResourceKind,
					EntitlementKey:         accessreview.ImplicitMemberEntitlementKey,
					EntitlementDisplayName: accessreview.ImplicitMemberEntitlementDisplayName,
				},
			},
		},
		{
			name: "each role becomes an entitlement on the default resource",
			account: drivers.AccountRecord{
				Email:      "Admin@Example.com",
				ExternalID: " 42 ",
				Roles:      []string{"Admin", "Viewer"},
			},
			want: []accessreview.GrantTriple{
				{
					PrincipalKey:           "admin@example.com|42",
					ResourceKey:            accessreview.DefaultResourceKey,
					ResourceKind:           accessreview.DefaultResourceKind,
					EntitlementKey:         "Admin",
					EntitlementDisplayName: "Admin",
				},
				{
					PrincipalKey:           "admin@example.com|42",
					ResourceKey:            accessreview.DefaultResourceKey,
					ResourceKind:           accessreview.DefaultResourceKind,
					EntitlementKey:         "Viewer",
					EntitlementDisplayName: "Viewer",
				},
			},
		},
		{
			name: "trims roles and skips blanks without adding member",
			account: drivers.AccountRecord{
				Email: "user@example.com",
				Roles: []string{"  Admin  ", "", "Viewer"},
			},
			want: []accessreview.GrantTriple{
				{
					PrincipalKey:           "user@example.com",
					ResourceKey:            accessreview.DefaultResourceKey,
					ResourceKind:           accessreview.DefaultResourceKind,
					EntitlementKey:         "Admin",
					EntitlementDisplayName: "Admin",
				},
				{
					PrincipalKey:           "user@example.com",
					ResourceKey:            accessreview.DefaultResourceKey,
					ResourceKind:           accessreview.DefaultResourceKind,
					EntitlementKey:         "Viewer",
					EntitlementDisplayName: "Viewer",
				},
			},
		},
		{
			name: "dedupes entitlement keys after trim",
			account: drivers.AccountRecord{
				Email: "user@example.com",
				Roles: []string{"Admin", "  Admin  ", "Viewer"},
			},
			want: []accessreview.GrantTriple{
				{
					PrincipalKey:           "user@example.com",
					ResourceKey:            accessreview.DefaultResourceKey,
					ResourceKind:           accessreview.DefaultResourceKind,
					EntitlementKey:         "Admin",
					EntitlementDisplayName: "Admin",
				},
				{
					PrincipalKey:           "user@example.com",
					ResourceKey:            accessreview.DefaultResourceKey,
					ResourceKind:           accessreview.DefaultResourceKind,
					EntitlementKey:         "Viewer",
					EntitlementDisplayName: "Viewer",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				got := accessreview.GrantTriplesFromAccount(tt.account)
				assert.Equal(t, tt.want, got)
			},
		)
	}
}

func TestGrantTriplesFromAccounts(t *testing.T) {
	t.Parallel()

	accounts := []drivers.AccountRecord{
		{
			Email: "a@example.com",
			Roles: []string{"Admin"},
		},
		{
			Email: "b@example.com",
			Roles: nil,
		},
	}

	got := accessreview.GrantTriplesFromAccounts(accounts)
	require.Len(t, got, 2)
	assert.Equal(
		t,
		[]accessreview.GrantTriple{
			{
				PrincipalKey:           "a@example.com",
				ResourceKey:            accessreview.DefaultResourceKey,
				ResourceKind:           accessreview.DefaultResourceKind,
				EntitlementKey:         "Admin",
				EntitlementDisplayName: "Admin",
			},
			{
				PrincipalKey:           "b@example.com",
				ResourceKey:            accessreview.DefaultResourceKey,
				ResourceKind:           accessreview.DefaultResourceKind,
				EntitlementKey:         accessreview.ImplicitMemberEntitlementKey,
				EntitlementDisplayName: accessreview.ImplicitMemberEntitlementDisplayName,
			},
		},
		got,
	)
}
