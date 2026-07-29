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
	"time"

	"github.com/jackc/pgx/v5"
)

type SCIMEventFilter struct {
	createdAtGte *time.Time
	createdAtLt  *time.Time
}

func NewSCIMEventFilter() *SCIMEventFilter {
	return &SCIMEventFilter{}
}

func (f *SCIMEventFilter) WithCreatedAtGte(t time.Time) *SCIMEventFilter {
	f.createdAtGte = &t
	return f
}

func (f *SCIMEventFilter) WithCreatedAtLt(t time.Time) *SCIMEventFilter {
	f.createdAtLt = &t
	return f
}

func (f *SCIMEventFilter) SQLFragment() string {
	return `
(
    CASE
        WHEN @filter_created_at_gte::timestamptz IS NOT NULL THEN
            created_at >= @filter_created_at_gte::timestamptz
        ELSE TRUE
    END
    AND
    CASE
        WHEN @filter_created_at_lt::timestamptz IS NOT NULL THEN
            created_at < @filter_created_at_lt::timestamptz
        ELSE TRUE
    END
)`
}

func (f *SCIMEventFilter) SQLArguments() pgx.StrictNamedArgs {
	args := pgx.StrictNamedArgs{
		"filter_created_at_gte": nil,
		"filter_created_at_lt":  nil,
	}

	if f.createdAtGte != nil {
		args["filter_created_at_gte"] = *f.createdAtGte
	}

	if f.createdAtLt != nil {
		args["filter_created_at_lt"] = *f.createdAtLt
	}

	return args
}
