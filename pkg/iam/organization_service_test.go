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

package iam

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestIsUserRemovalDependencyError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "returns true for sentinel error",
			err:  coredata.ErrResourceInUse,
			want: true,
		},
		{
			name: "returns true for wrapped sentinel error",
			err:  fmt.Errorf("wrapped: %w", coredata.ErrResourceInUse),
			want: true,
		},
		{
			name: "returns true for wrapped postgres foreign key error",
			err: fmt.Errorf(
				"wrapped: %w",
				&pgconn.PgError{Code: "23503"},
			),
			want: true,
		},
		{
			name: "returns false for non foreign key postgres error",
			err: fmt.Errorf(
				"wrapped: %w",
				&pgconn.PgError{Code: "23505"},
			),
			want: false,
		},
		{
			name: "returns false for unrelated error",
			err:  errors.New("boom"),
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, isUserRemovalDependencyError(tt.err))
		})
	}
}

func TestNewUserReferencedByRecordsError_Message(t *testing.T) {
	t.Parallel()

	err := NewUserReferencedByRecordsError(gid.Nil)
	resourceErr, ok := errors.AsType[*ErrUserReferencedByRecords](err)
	require.True(t, ok)
	assert.Equal(
		t,
		"cannot remove user because they are referenced by existing records (for example signatures, tasks, assets, or risks)",
		resourceErr.Error(),
	)
}
