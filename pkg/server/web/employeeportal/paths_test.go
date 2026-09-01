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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/server/web/employeeportal"
)

func TestOrganizationPath(t *testing.T) {
	t.Parallel()

	t.Run(
		"organization home",
		func(t *testing.T) {
			t.Parallel()

			path, err := employeeportal.OrganizationPath("org_123")
			require.NoError(t, err)
			assert.Equal(t, "/employee-portal/org_123", path)
		},
	)

	t.Run(
		"section and document",
		func(t *testing.T) {
			t.Parallel()

			path, err := employeeportal.OrganizationPath("org_123", "signatures", "doc_456")
			require.NoError(t, err)
			assert.Equal(t, "/employee-portal/org_123/signatures/doc_456", path)
		},
	)
}

func TestLegacyRedirectPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		path   string
		want   string
		wantOK bool
	}{
		{name: "enroll", path: "/enroll", want: "/employee-portal/enroll", wantOK: true},
		{name: "employee home", path: "/organizations/org_123/employee", want: "/employee-portal/org_123", wantOK: true},
		{name: "signatures", path: "/organizations/org_123/employee/signatures", want: "/employee-portal/org_123/signatures", wantOK: true},
		{name: "signature document", path: "/organizations/org_123/employee/signatures/doc_456", want: "/employee-portal/org_123/signatures/doc_456", wantOK: true},
		{name: "approvals", path: "/organizations/org_123/employee/approvals", want: "/employee-portal/org_123/approvals", wantOK: true},
		{name: "approval document", path: "/organizations/org_123/employee/approvals/doc_456", want: "/employee-portal/org_123/approvals/doc_456", wantOK: true},
		{name: "devices", path: "/organizations/org_123/employee/devices", want: "/employee-portal/org_123/devices", wantOK: true},
		{name: "legacy document alias", path: "/organizations/org_123/employee/doc_456", want: "/employee-portal/org_123/signatures/doc_456", wantOK: true},
		{name: "bind stays on console", path: "/organizations/org_123/employee/bind", wantOK: false},
		{name: "bindings stays on console", path: "/organizations/org_123/employee/bindings", wantOK: false},
		{name: "unrelated path", path: "/organizations/org_123/governance/tasks", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				got, ok := employeeportal.LegacyRedirectPath(tt.path)
				assert.Equal(t, tt.wantOK, ok)
				assert.Equal(t, tt.want, got)
			},
		)
	}
}

func TestLegacyRedirectMiddleware(t *testing.T) {
	t.Parallel()

	t.Run(
		"redirects enroll",
		func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})
			handler := employeeportal.LegacyRedirectMiddleware(next)

			req := httptest.NewRequest(http.MethodGet, "/enroll?foo=bar", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusFound, rec.Code)
			assert.Equal(t, "/employee-portal/enroll?foo=bar", rec.Header().Get("Location"))
		},
	)

	t.Run(
		"preserves query on employee signatures",
		func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})
			handler := employeeportal.LegacyRedirectMiddleware(next)

			req := httptest.NewRequest(http.MethodGet, "/organizations/org_123/employee/signatures?foo=bar", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusFound, rec.Code)
			assert.Equal(t, "/employee-portal/org_123/signatures?foo=bar", rec.Header().Get("Location"))
		},
	)

	t.Run(
		"falls through for bind",
		func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})
			handler := employeeportal.LegacyRedirectMiddleware(next)

			req := httptest.NewRequest(http.MethodGet, "/organizations/org_123/employee/bind?token=abc", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Empty(t, rec.Header().Get("Location"))
		},
	)
}
