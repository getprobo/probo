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
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"go.probo.inc/probo/pkg/coredata"
)

// CSVDriver supports both identity and access use cases from uploaded CSV
// files. No external connector is needed.
//
// Expected CSV columns (header required): email, full_name, role, job_title,
// is_admin, active, external_id
type CSVDriver struct {
	reader io.Reader
}

func NewCSVDriver(reader io.Reader) *CSVDriver {
	return &CSVDriver{reader: reader}
}

func (d *CSVDriver) ListAccounts(_ context.Context) ([]AccountRecord, error) {
	r := csv.NewReader(d.reader)

	// Read header
	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("cannot read CSV header: %w", err)
	}

	colIndex := make(map[string]int)
	for i, col := range header {
		colIndex[strings.TrimSpace(strings.ToLower(col))] = i
	}
	if _, ok := colIndex["email"]; !ok {
		return nil, fmt.Errorf("cannot parse CSV: missing required column email")
	}

	var records []AccountRecord

	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("cannot read CSV row: %w", err)
		}

		record := AccountRecord{
			MFAStatus:  coredata.MFAStatusUnknown,
			AuthMethod: coredata.AccessEntryAuthMethodUnknown,
		}

		if idx, ok := colIndex["email"]; ok && idx < len(row) {
			record.Email = strings.TrimSpace(row[idx])
		}
		if idx, ok := colIndex["full_name"]; ok && idx < len(row) {
			record.FullName = strings.TrimSpace(row[idx])
		}
		if idx, ok := colIndex["role"]; ok && idx < len(row) {
			record.Role = strings.TrimSpace(row[idx])
		}
		if idx, ok := colIndex["job_title"]; ok && idx < len(row) {
			record.JobTitle = strings.TrimSpace(row[idx])
		}
		if idx, ok := colIndex["is_admin"]; ok && idx < len(row) {
			record.IsAdmin = strings.TrimSpace(strings.ToLower(row[idx])) == "true"
		}
		if idx, ok := colIndex["active"]; ok && idx < len(row) {
			record.Active = strings.TrimSpace(strings.ToLower(row[idx])) == "true"
		}
		if idx, ok := colIndex["external_id"]; ok && idx < len(row) {
			record.ExternalID = strings.TrimSpace(row[idx])
		}

		if record.Email != "" {
			records = append(records, record)
		}
	}

	return records, nil
}
