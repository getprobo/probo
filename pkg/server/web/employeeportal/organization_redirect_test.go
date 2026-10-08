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

package employeeportal_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/server/web/employeeportal"
)

func TestOrganizationGIDRedirectMiddleware(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	portalID := gid.New(tenantID, coredata.EmployeePortalEntityType)
	documentID := gid.New(tenantID, coredata.DocumentEntityType)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	newHandler := func(t *testing.T) http.Handler {
		t.Helper()

		lookup := func(ctx context.Context, id gid.GID) (gid.GID, error) {
			require.Equal(t, organizationID, id)
			return portalID, nil
		}

		return employeeportal.OrganizationGIDRedirectMiddleware(nil, lookup, next)
	}

	t.Run("redirects organization home", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/"+organizationID.String(), nil)
		rec := httptest.NewRecorder()
		newHandler(t).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusFound, rec.Code)
		assert.Equal(t, "/employee-portal/"+portalID.String(), rec.Header().Get("Location"))
	})

	t.Run("redirects nested path and query", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(
			http.MethodGet,
			"/"+organizationID.String()+"/signatures/"+documentID.String()+"?foo=bar",
			nil,
		)
		rec := httptest.NewRecorder()
		newHandler(t).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusFound, rec.Code)
		assert.Equal(
			t,
			"/employee-portal/"+portalID.String()+"/signatures/"+documentID.String()+"?foo=bar",
			rec.Header().Get("Location"),
		)
	})

	t.Run("does not double-escape encoded suffix", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(
			http.MethodGet,
			"/"+organizationID.String()+"/foo%20bar",
			nil,
		)
		rec := httptest.NewRecorder()
		newHandler(t).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusFound, rec.Code)
		assert.Equal(
			t,
			"/employee-portal/"+portalID.String()+"/foo%20bar",
			rec.Header().Get("Location"),
		)
	})

	t.Run("does not redirect portal gid", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/"+portalID.String()+"/devices", nil)
		rec := httptest.NewRecorder()
		newHandler(t).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("does not redirect enroll", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/enroll", nil)
		rec := httptest.NewRecorder()
		newHandler(t).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("not found when organization has no portal", func(t *testing.T) {
		t.Parallel()

		missing := employeeportal.OrganizationGIDRedirectMiddleware(
			nil,
			func(ctx context.Context, id gid.GID) (gid.GID, error) {
				return gid.Nil, coredata.ErrResourceNotFound
			},
			next,
		)

		req := httptest.NewRequest(http.MethodGet, "/"+organizationID.String(), nil)
		rec := httptest.NewRecorder()
		missing.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("internal error when lookup fails", func(t *testing.T) {
		t.Parallel()

		failing := employeeportal.OrganizationGIDRedirectMiddleware(
			nil,
			func(ctx context.Context, id gid.GID) (gid.GID, error) {
				return gid.Nil, errors.New("boom")
			},
			next,
		)

		req := httptest.NewRequest(http.MethodGet, "/"+organizationID.String(), nil)
		rec := httptest.NewRecorder()
		failing.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
