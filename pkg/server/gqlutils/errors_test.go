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

package gqlutils_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/server/gqlutils"
	"go.probo.inc/probo/pkg/validator"
)

func TestMapError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := log.NewLogger(log.WithOutput(io.Discard))

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		err := gqlutils.MapError(
			ctx,
			logger,
			"cannot update connector",
			fmt.Errorf("cannot load connector: %w", coredata.ErrResourceNotFound),
		)

		gqlErr, ok := errors.AsType[*gqlerror.Error](err)
		require.True(t, ok)
		assert.Equal(t, "NOT_FOUND", gqlErr.Extensions["code"])
		assert.Equal(t, "cannot load connector: resource not found", gqlErr.Message)
	})

	t.Run("validation errors", func(t *testing.T) {
		t.Parallel()

		validationErrors := validator.ValidationErrors{
			&validator.ValidationError{
				Field:   "name",
				Code:    validator.ErrorCodeRequired,
				Message: "is required",
			},
		}

		err := gqlutils.MapError(
			ctx,
			logger,
			"cannot update connector",
			fmt.Errorf("cannot validate connector: %w", validationErrors),
		)

		list, ok := errors.AsType[gqlerror.List](err)
		require.True(t, ok)
		require.Len(t, list, 1)
		assert.Equal(t, "INVALID", list[0].Extensions["code"])
		assert.Equal(t, "name", list[0].Extensions["field"])
		assert.Equal(t, validator.ErrorCodeRequired, list[0].Extensions["cause"])
	})

	t.Run("unknown error", func(t *testing.T) {
		t.Parallel()

		err := gqlutils.MapError(
			ctx,
			logger,
			"cannot update connector",
			errors.New("database is down"),
		)

		gqlErr, ok := errors.AsType[*gqlerror.Error](err)
		require.True(t, ok)
		assert.Equal(t, "INTERNAL", gqlErr.Extensions["code"])
		assert.Equal(t, "An internal server error occurred. Please try again later.", gqlErr.Message)
	})
}
