// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package accesssource

import (
	"context"
	"strings"
	"testing"
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
}
