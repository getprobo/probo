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

package testutil

import "testing"

type (
	// OrganizationRoles is one organization with a signed-up owner and invited
	// admin and viewer members, each using a real authenticated client.
	OrganizationRoles struct {
		Owner  *Client
		Admin  *Client
		Viewer *Client
	}
)

// NewOrganizationRoles signs up an owner and invites an admin and viewer through
// the same black-box flows as NewClient and NewClientInOrg.
func NewOrganizationRoles(t testing.TB) OrganizationRoles {
	t.Helper()

	owner := NewClient(t, RoleOwner)

	return OrganizationRoles{
		Owner:  owner,
		Admin:  NewClientInOrg(t, RoleAdmin, owner),
		Viewer: NewClientInOrg(t, RoleViewer, owner),
	}
}

// Client returns the authenticated client for a standard organization role.
func (o OrganizationRoles) Client(t testing.TB, role TestRole) *Client {
	t.Helper()

	switch role {
	case RoleOwner:
		return o.Owner
	case RoleAdmin:
		return o.Admin
	case RoleViewer:
		return o.Viewer
	default:
		t.Fatalf("unsupported organization role: %s", role)

		return nil
	}
}
