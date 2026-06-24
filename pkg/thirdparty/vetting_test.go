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

package thirdparty

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/validator"
)

func TestVetRequest_Validate(t *testing.T) {
	t.Parallel()

	validID := gid.New(gid.NewTenantID(), coredata.ThirdPartyEntityType)

	t.Run("accepts a valid request", func(t *testing.T) {
		t.Parallel()

		procedure := "Focus on SOC 2"

		err := VetRequest{
			ID:         validID,
			WebsiteURL: "https://example.com",
			Procedure:  &procedure,
		}.Validate()
		require.NoError(t, err)
	})

	t.Run("requires id", func(t *testing.T) {
		t.Parallel()

		err := VetRequest{
			WebsiteURL: "https://example.com",
		}.Validate()
		require.Error(t, err)

		validationErrors, ok := errors.AsType[validator.ValidationErrors](err)
		require.True(t, ok)
		assert.NotEmpty(t, validationErrors.ByField("id"))
	})

	t.Run("requires website url", func(t *testing.T) {
		t.Parallel()

		err := VetRequest{ID: validID}.Validate()
		require.Error(t, err)

		validationErrors, ok := errors.AsType[validator.ValidationErrors](err)
		require.True(t, ok)
		assert.NotEmpty(t, validationErrors.ByField("website_url"))
	})

	t.Run("rejects an invalid third party id", func(t *testing.T) {
		t.Parallel()

		err := VetRequest{
			ID:         gid.New(gid.NewTenantID(), coredata.OrganizationEntityType),
			WebsiteURL: "https://example.com",
		}.Validate()
		require.Error(t, err)

		validationErrors, ok := errors.AsType[validator.ValidationErrors](err)
		require.True(t, ok)
		assert.NotEmpty(t, validationErrors.ByField("id"))
	})
}

func TestSanitizeVettingError(t *testing.T) {
	t.Parallel()

	t.Run("returns short messages unchanged", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "cannot vet third party", sanitizeVettingError(errors.New("cannot vet third party")))
	})

	t.Run("truncates long messages on a rune boundary", func(t *testing.T) {
		t.Parallel()

		msg := strings.Repeat("x", vettingErrorMessageMaxLen+10)

		sanitized := sanitizeVettingError(errors.New(msg))

		assert.LessOrEqual(t, len(sanitized), vettingErrorMessageMaxLen+len("…"))
		assert.True(t, strings.HasSuffix(sanitized, "…"))
	})
}

func TestDisabledVetter_Assess(t *testing.T) {
	t.Parallel()

	_, err := DisabledVetter{}.Assess(context.Background(), "https://example.com", "", nil, nil)
	require.ErrorIs(t, err, ErrVettingDisabled)
}

func TestDisabledVetter_ImplementsVetter(t *testing.T) {
	t.Parallel()

	var _ Vetter = DisabledVetter{}
}
