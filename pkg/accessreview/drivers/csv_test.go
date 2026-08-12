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
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCSVDriverRequiresEmailHeader(t *testing.T) {
	t.Parallel()

	driver := NewCSVDriver(strings.NewReader("full_name,role\nJane Doe,Admin\n"))

	_, err := driver.ListAccounts(context.Background())
	if err == nil {
		t.Fatalf("expected error when email header is missing")
	}
}

func TestCSVDriverParsesRequiredAndOptionalColumns(t *testing.T) {
	t.Parallel()

	driver := NewCSVDriver(strings.NewReader(
		"email,full_name,role,external_id\njane@example.com,Jane Doe,Admin,42\n",
	))

	records, err := driver.ListAccounts(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	if records[0].Email != "jane@example.com" {
		t.Fatalf("unexpected email: %s", records[0].Email)
	}

	if records[0].ExternalID != "42" {
		t.Fatalf("unexpected external id: %s", records[0].ExternalID)
	}

	if records[0].IsAdmin != nil {
		t.Fatal("expected admin status to be unknown when is_admin is omitted")
	}
}

func TestCSVDriverParsesAdminStatus(t *testing.T) {
	t.Parallel()

	driver := NewCSVDriver(strings.NewReader(
		"email,is_admin\nadmin@example.com,true\nmember@example.com,false\nunknown@example.com,\n",
	))

	records, err := driver.ListAccounts(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(records) != 3 {
		t.Fatalf("expected 3 records, got %d", len(records))
	}

	if records[0].IsAdmin == nil || !*records[0].IsAdmin {
		t.Fatal("expected true admin status")
	}

	if records[1].IsAdmin == nil || *records[1].IsAdmin {
		t.Fatal("expected false admin status")
	}

	if records[2].IsAdmin != nil {
		t.Fatal("expected blank admin status to be unknown")
	}
}

func TestCSVDriverParsesActiveStatus(t *testing.T) {
	t.Parallel()

	driver := NewCSVDriver(strings.NewReader(
		"email,active\nactive@example.com,true\ninactive@example.com,false\nblank@example.com,\ninvalid@example.com,yes\n",
	))

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 4)
	assert.Equal(t, new(true), records[0].Active)
	assert.Equal(t, new(false), records[1].Active)
	assert.Nil(t, records[2].Active)
	assert.Nil(t, records[3].Active)
}

func TestCSVDriverSplitsRoles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		role     string
		expected []string
	}{
		{
			name:     "single role",
			role:     "Admin",
			expected: []string{"Admin"},
		},
		{
			name:     "semicolon-separated roles",
			role:     "Admin; Editor;Viewer",
			expected: []string{"Admin", "Editor", "Viewer"},
		},
		{
			name:     "empty roles are ignored",
			role:     "Admin; ;Viewer;",
			expected: []string{"Admin", "Viewer"},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				input := "email,role\nuser@example.com," + tt.role + "\n"
				driver := NewCSVDriver(strings.NewReader(input))

				records, err := driver.ListAccounts(context.Background())
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if len(records) != 1 {
					t.Fatalf("expected 1 record, got %d", len(records))
				}

				if !slices.Equal(records[0].Roles, tt.expected) {
					t.Errorf("expected roles %v, got %v", tt.expected, records[0].Roles)
				}
			},
		)
	}
}
