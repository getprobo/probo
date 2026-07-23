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

package coredata

import (
	"github.com/jackc/pgx/v5"
	"go.probo.inc/probo/pkg/gid"
)

type (
	BusinessFunctionFilter struct {
		classification *BusinessFunctionClassification
		ownerID        *gid.GID
		cifOnly        *bool
	}
)

func NewBusinessFunctionFilter(
	classification *BusinessFunctionClassification,
	ownerID *gid.GID,
	cifOnly *bool,
) *BusinessFunctionFilter {
	return &BusinessFunctionFilter{
		classification: classification,
		ownerID:        ownerID,
		cifOnly:        cifOnly,
	}
}

func (f *BusinessFunctionFilter) SQLArguments() pgx.StrictNamedArgs {
	args := pgx.StrictNamedArgs{
		"has_classification_filter": false,
		"filter_classification":     nil,
		"has_owner_filter":          false,
		"filter_owner_id":           nil,
		"has_cif_only_filter":       false,
		"filter_critical":           string(BusinessFunctionClassificationCritical),
		"filter_important":          string(BusinessFunctionClassificationImportant),
	}

	if f.classification != nil {
		args["has_classification_filter"] = true
		args["filter_classification"] = string(*f.classification)
	}

	if f.ownerID != nil {
		args["has_owner_filter"] = true
		args["filter_owner_id"] = *f.ownerID
	}

	if f.cifOnly != nil && *f.cifOnly {
		args["has_cif_only_filter"] = true
	}

	return args
}

func (f *BusinessFunctionFilter) SQLFragment() string {
	return `
(
    CASE
        WHEN @has_classification_filter::boolean = false THEN TRUE
        WHEN @has_classification_filter::boolean = true THEN
            classification = @filter_classification::business_function_classifications
        ELSE TRUE
    END
    AND
    CASE
        WHEN @has_owner_filter::boolean = false THEN TRUE
        WHEN @has_owner_filter::boolean = true THEN
            owner_id = @filter_owner_id::text
        ELSE TRUE
    END
    AND
    CASE
        WHEN @has_cif_only_filter::boolean = false THEN TRUE
        WHEN @has_cif_only_filter::boolean = true THEN
            classification IN (
                @filter_critical::business_function_classifications,
                @filter_important::business_function_classifications
            )
        ELSE TRUE
    END
)`
}
