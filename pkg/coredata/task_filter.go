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

import "github.com/jackc/pgx/v5"

type TaskFilter struct {
	query *string
	state *TaskState
}

func NewTaskFilter(query *string, state *TaskState) *TaskFilter {
	return &TaskFilter{
		query: query,
		state: state,
	}
}

func (f *TaskFilter) SQLFragment() string {
	return `
(
	CASE
		WHEN @filter_query::text IS NOT NULL AND @filter_query::text <> '' THEN
			name ILIKE '%' || @filter_query || '%' ESCAPE '\'
		ELSE TRUE
	END
	AND CASE
		WHEN @filter_state::text IS NOT NULL THEN
			state = @filter_state::task_state
		ELSE TRUE
	END
)`
}

func (f *TaskFilter) SQLArguments() pgx.StrictNamedArgs {
	args := pgx.StrictNamedArgs{
		"filter_query": nil,
		"filter_state": nil,
	}

	if f.query != nil && *f.query != "" {
		args["filter_query"] = escapeLikePattern(*f.query)
	}

	if f.state != nil {
		args["filter_state"] = string(*f.state)
	}

	return args
}
