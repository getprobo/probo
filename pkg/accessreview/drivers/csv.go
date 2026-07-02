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
	r.FieldsPerRecord = -1

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
			MFAStatus:   coredata.MFAStatusUnknown,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
		}

		if idx, ok := colIndex["email"]; ok && idx < len(row) {
			record.Email = strings.TrimSpace(row[idx])
		}

		if idx, ok := colIndex["full_name"]; ok && idx < len(row) {
			record.FullName = strings.TrimSpace(row[idx])
		}

		if idx, ok := colIndex["role"]; ok && idx < len(row) {
			role := strings.TrimSpace(row[idx])

			roles := []string{}
			if role != "" {
				roles = []string{role}
			}

			record.Roles = roles
		}

		if idx, ok := colIndex["job_title"]; ok && idx < len(row) {
			record.JobTitle = strings.TrimSpace(row[idx])
		}

		if idx, ok := colIndex["is_admin"]; ok && idx < len(row) {
			record.IsAdmin = strings.TrimSpace(strings.ToLower(row[idx])) == "true"
		}

		if idx, ok := colIndex["active"]; ok && idx < len(row) {
			record.Active = new(strings.TrimSpace(strings.ToLower(row[idx])) == "true")
		}

		if idx, ok := colIndex["external_id"]; ok && idx < len(row) {
			record.ExternalID = strings.TrimSpace(row[idx])
		}

		if idx, ok := colIndex["account_type"]; ok && idx < len(row) {
			if strings.TrimSpace(strings.ToUpper(row[idx])) == "SERVICE_ACCOUNT" {
				record.AccountType = coredata.AccessReviewEntryAccountTypeServiceAccount
			}
		}

		if record.Email != "" {
			records = append(records, record)
		}
	}

	return records, nil
}
