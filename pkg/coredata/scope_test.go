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

package coredata_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestScope_PrincipalID(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	principalID := gid.New(tenantID, coredata.OrganizationEntityType)

	t.Run("starts empty", func(t *testing.T) {
		t.Parallel()

		scope := coredata.NewScope(tenantID, gid.Nil)

		assert.Equal(t, gid.Nil, scope.PrincipalID())
		assert.Equal(t, tenantID, scope.GetTenantID())
	})

	t.Run("keeps the principal and the tenant", func(t *testing.T) {
		t.Parallel()

		scope := coredata.NewScope(tenantID, principalID)

		assert.Equal(t, principalID, scope.PrincipalID())
		assert.Equal(t, tenantID, scope.GetTenantID())

		_, inSQL := scope.SQLArguments()["principal_id"]
		assert.False(t, inSQL)
		assert.Equal(t, tenantID, scope.SQLArguments()["tenant_id"])
	})

	t.Run("object scope has no principal", func(t *testing.T) {
		t.Parallel()

		objectID := gid.New(tenantID, coredata.OrganizationEntityType)
		scope := coredata.NewScopeFromObjectID(objectID)

		assert.Equal(t, gid.Nil, scope.PrincipalID())
		assert.Equal(t, tenantID, scope.GetTenantID())
	})

	t.Run("no scope has no principal", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, gid.Nil, coredata.NewNoScope().PrincipalID())
	})
}
