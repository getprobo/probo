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
