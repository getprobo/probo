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

package mcpx_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/server/api/mcp/mcpx"
	"go.probo.inc/probo/pkg/validator"
)

func TestMapError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := log.NewLogger(log.WithOutput(io.Discard))

	tests := []struct {
		name   string
		err    error
		want   string
		wantIs error
	}{
		{
			name:   "no connector account",
			err:    fmt.Errorf("cannot create access source: %w", accessreview.ErrNoConnectorAccount),
			want:   accessreview.ErrNoConnectorAccount.Error(),
			wantIs: accessreview.ErrNoConnectorAccount,
		},
		{
			name:   "multiple connector accounts",
			err:    fmt.Errorf("cannot resolve connector account: %w", coredata.ErrMultipleConnectorAccounts),
			want:   coredata.ErrMultipleConnectorAccounts.Error(),
			wantIs: coredata.ErrMultipleConnectorAccounts,
		},
		{
			name:   "not found",
			err:    fmt.Errorf("cannot load connector: %w", coredata.ErrResourceNotFound),
			want:   mcpx.ErrResourceNotFound.Error(),
			wantIs: mcpx.ErrResourceNotFound,
		},
		{
			name:   "already exists",
			err:    fmt.Errorf("cannot insert connector: %w", coredata.ErrResourceAlreadyExists),
			want:   mcpx.ErrResourceAlreadyExists.Error(),
			wantIs: mcpx.ErrResourceAlreadyExists,
		},
		{
			name:   "in use",
			err:    fmt.Errorf("cannot use connector: it is used by a SCIM configuration: %w", coredata.ErrResourceInUse),
			want:   mcpx.ErrResourceInUse.Error(),
			wantIs: mcpx.ErrResourceInUse,
		},
		{
			name:   "permission denied",
			err:    iam.NewInsufficientPermissionsError(gid.Nil, gid.Nil, ""),
			want:   mcpx.ErrForbidden.Error(),
			wantIs: mcpx.ErrForbidden,
		},
		{
			name:   "assumption required",
			err:    iam.NewAssumptionRequiredError(gid.Nil, gid.Nil),
			want:   mcpx.ErrAssumptionRequired.Error(),
			wantIs: mcpx.ErrAssumptionRequired,
		},
		{
			name:   "unknown error",
			err:    errors.New("database is down"),
			want:   mcpx.ErrInternal.Error(),
			wantIs: mcpx.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := mcpx.MapError(ctx, logger, "cannot create access source", tt.err)

			require.Error(t, got)
			assert.Equal(t, tt.want, got.Error())
			assert.ErrorIs(t, got, tt.wantIs)
		})
	}

	t.Run("validation errors", func(t *testing.T) {
		t.Parallel()

		validationErrors := validator.ValidationErrors{
			{
				Field:   "name",
				Code:    validator.ErrorCodeRequired,
				Message: "is required",
			},
		}

		got := mcpx.MapError(
			ctx,
			logger,
			"cannot create access source",
			fmt.Errorf("cannot validate access source: %w", validationErrors),
		)

		gotErrors, ok := errors.AsType[validator.ValidationErrors](got)
		require.True(t, ok)
		assert.Equal(t, validationErrors, gotErrors)
	})

	t.Run("validation error", func(t *testing.T) {
		t.Parallel()

		validationError := &validator.ValidationError{
			Field:   "name",
			Code:    validator.ErrorCodeTooLong,
			Message: "is too long",
		}

		got := mcpx.MapError(
			ctx,
			logger,
			"cannot create access source",
			fmt.Errorf("cannot validate access source: %w", validationError),
		)

		gotError, ok := errors.AsType[*validator.ValidationError](got)
		require.True(t, ok)
		assert.Equal(t, validationError, gotError)
	})
}
