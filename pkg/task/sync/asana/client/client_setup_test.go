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

package client

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectSetupTreatsFreeWorkspaceAsNoCustomFields(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case strings.HasSuffix(r.URL.Path, "/custom_field_settings"):
			w.WriteHeader(http.StatusPaymentRequired)
			_, _ = w.Write([]byte(`{"errors":[{"message":"Custom Field Settings are not available for free users."}]}`))
		case strings.HasSuffix(r.URL.Path, "/sections"):
			_, _ = w.Write([]byte(`{"data":[{"gid":"s1","name":"Doing"}]}`))
		default:
			_, _ = w.Write([]byte(`{"data":{"gid":"p1","name":"Work","workspace":{"gid":"w1"}}}`))
		}
	}))
	t.Cleanup(server.Close)

	setup, err := NewClient(server.Client(), server.URL).ProjectSetup(t.Context(), "p1")
	require.NoError(t, err)
	assert.False(t, setup.CustomFieldsAvailable)
	assert.Equal(t, "Doing", setup.Sections[0].Name)
	assert.Empty(t, setup.EnumFields)
}

func TestProjectSetup_PaginatesSectionsAndCustomFields(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case strings.HasSuffix(r.URL.Path, "/custom_field_settings"):
			if r.URL.Query().Get("offset") == "" {
				_, _ = w.Write([]byte(`{"data":[{"custom_field":{"gid":"f1","name":"Priority","resource_subtype":"enum","enum_options":[{"gid":"o1","name":"High","enabled":true}]}}],"next_page":{"offset":"fields-2"}}`))

				return
			}

			_, _ = w.Write([]byte(`{"data":[{"custom_field":{"gid":"f2","name":"Status","resource_subtype":"enum","enum_options":[{"gid":"o2","name":"Open","enabled":true},{"gid":"o3","name":"Old","enabled":false}]}}]}`))
		case strings.HasSuffix(r.URL.Path, "/sections"):
			if r.URL.Query().Get("offset") == "" {
				_, _ = w.Write([]byte(`{"data":[{"gid":"s1","name":"Doing"}],"next_page":{"offset":"sections-2"}}`))

				return
			}

			_, _ = w.Write([]byte(`{"data":[{"gid":"s2","name":"Done"}]}`))
		default:
			_, _ = w.Write([]byte(`{"data":{"gid":"p1","name":"Work","team":{"gid":"t1"},"workspace":{"gid":"w1"}}}`))
		}
	}))
	t.Cleanup(server.Close)

	setup, err := NewClient(server.Client(), server.URL).ProjectSetup(t.Context(), "p1")
	require.NoError(t, err)
	require.Len(t, setup.Sections, 2)
	assert.Equal(t, "Doing", setup.Sections[0].Name)
	assert.Equal(t, "Done", setup.Sections[1].Name)
	require.Len(t, setup.EnumFields, 2)
	assert.Equal(t, "Priority", setup.EnumFields[0].Name)
	assert.Equal(t, "Status", setup.EnumFields[1].Name)
	require.Len(t, setup.EnumFields[1].Options, 1)
	assert.Equal(t, "Open", setup.EnumFields[1].Options[0].Name)
	assert.True(t, setup.CustomFieldsAvailable)
	assert.Equal(t, "t1", setup.TeamGID)
}

func TestProjectSetup_RejectsArchivedProject(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/projects/p1") {
			t.Errorf("unexpected request %s", r.URL.Path)
			http.NotFound(w, r)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"gid":"p1","name":"Old","archived":true}}`))
	}))
	t.Cleanup(server.Close)

	_, err := NewClient(server.Client(), server.URL).ProjectSetup(t.Context(), "p1")
	require.ErrorIs(t, err, ErrProjectArchived)
	assert.ErrorContains(t, err, "asana project is archived")
}
