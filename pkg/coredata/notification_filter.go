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
)

type (
	NotificationFilter struct {
		types    []string
		statuses []string
	}
)

func NewNotificationFilter(
	types []NotificationType,
	statuses []NotificationStatus,
) *NotificationFilter {
	filter := &NotificationFilter{}
	if types != nil {
		values := make([]string, len(types))
		for i, notificationType := range types {
			values[i] = notificationType.String()
		}

		filter.types = values
	}

	if statuses != nil {
		values := make([]string, len(statuses))
		for i, status := range statuses {
			values[i] = status.String()
		}

		filter.statuses = values
	}

	return filter
}

func (f *NotificationFilter) SQLArguments() pgx.StrictNamedArgs {
	return pgx.StrictNamedArgs{
		"filter_types":    f.types,
		"filter_statuses": f.statuses,
	}
}

func (f *NotificationFilter) SQLFragment() string {
	return `
(
	CASE
		WHEN @filter_types::text[] IS NOT NULL THEN
			type::text = ANY(@filter_types::text[])
		ELSE TRUE
	END
	AND CASE
		WHEN @filter_statuses::text[] IS NOT NULL THEN
			status::text = ANY(@filter_statuses::text[])
		ELSE TRUE
	END
)`
}
