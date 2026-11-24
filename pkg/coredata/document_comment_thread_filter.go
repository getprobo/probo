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

type DocumentCommentThreadFilter struct {
	resolved     *bool
	autoRejected *bool
}

// NewDocumentCommentThreadFilter filters threads by resolution and by whether
// their suggestion was rejected automatically. A thread without a suggestion
// counts as not rejected automatically.
func NewDocumentCommentThreadFilter(resolved *bool, autoRejected *bool) *DocumentCommentThreadFilter {
	return &DocumentCommentThreadFilter{
		resolved:     resolved,
		autoRejected: autoRejected,
	}
}

func (f *DocumentCommentThreadFilter) SQLFragment() string {
	return `
(
    CASE
        WHEN @has_thread_resolved::boolean = FALSE THEN TRUE
        WHEN @thread_resolved::boolean THEN resolved_at IS NOT NULL
        ELSE resolved_at IS NULL
    END
    AND CASE
        WHEN @has_thread_auto_rejected::boolean = FALSE THEN TRUE
        WHEN @thread_auto_rejected::boolean THEN EXISTS (
            SELECT
                1
            FROM
                document_suggestions s
            WHERE
                s.id = document_comment_threads.suggestion_id
                AND s.auto_rejected_at IS NOT NULL
        )
        ELSE NOT EXISTS (
            SELECT
                1
            FROM
                document_suggestions s
            WHERE
                s.id = document_comment_threads.suggestion_id
                AND s.auto_rejected_at IS NOT NULL
        )
    END
)`
}

func (f *DocumentCommentThreadFilter) SQLArguments() pgx.StrictNamedArgs {
	resolved := false
	hasResolved := false
	autoRejected := false
	hasAutoRejected := false

	if f != nil && f.resolved != nil {
		resolved = *f.resolved
		hasResolved = true
	}

	if f != nil && f.autoRejected != nil {
		autoRejected = *f.autoRejected
		hasAutoRejected = true
	}

	return pgx.StrictNamedArgs{
		"has_thread_resolved":      hasResolved,
		"thread_resolved":          resolved,
		"has_thread_auto_rejected": hasAutoRejected,
		"thread_auto_rejected":     autoRejected,
	}
}
