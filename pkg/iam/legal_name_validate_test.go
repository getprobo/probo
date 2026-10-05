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

package iam_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/validator"
)

func TestCreateOrganizationRequest_Validate_LegalName(t *testing.T) {
	t.Parallel()

	t.Run("trims surrounding whitespace", func(t *testing.T) {
		t.Parallel()

		legalName := "  Acme Inc.  "
		req := &iam.CreateOrganizationRequest{
			Name:      "Acme",
			LegalName: &legalName,
		}
		require.NoError(t, req.Validate())
		require.NotNil(t, req.LegalName)
		assert.Equal(t, "Acme Inc.", *req.LegalName)
	})

	t.Run("rejects whitespace-only", func(t *testing.T) {
		t.Parallel()

		legalName := " \t "
		req := &iam.CreateOrganizationRequest{
			Name:      "Acme",
			LegalName: &legalName,
		}
		err := req.Validate()
		require.Error(t, err)

		validationErrors, ok := errors.AsType[validator.ValidationErrors](err)
		require.True(t, ok)
		assert.NotEmpty(t, validationErrors.ByField("legalName"))
	})

	t.Run("allows omitted", func(t *testing.T) {
		t.Parallel()

		req := &iam.CreateOrganizationRequest{Name: "Acme"}
		require.NoError(t, req.Validate())
		assert.Nil(t, req.LegalName)
	})
}

func TestUpdateOrganizationRequest_Validate_LegalName(t *testing.T) {
	t.Parallel()

	t.Run("trims surrounding whitespace", func(t *testing.T) {
		t.Parallel()

		legalName := "  Acme Inc.  "
		legalNamePtr := &legalName
		req := &iam.UpdateOrganizationRequest{LegalName: &legalNamePtr}
		require.NoError(t, req.Validate())
		require.NotNil(t, req.LegalName)
		require.NotNil(t, *req.LegalName)
		assert.Equal(t, "Acme Inc.", **req.LegalName)
	})

	t.Run("rejects whitespace-only", func(t *testing.T) {
		t.Parallel()

		legalName := "   "
		legalNamePtr := &legalName
		req := &iam.UpdateOrganizationRequest{LegalName: &legalNamePtr}
		err := req.Validate()
		require.Error(t, err)

		validationErrors, ok := errors.AsType[validator.ValidationErrors](err)
		require.True(t, ok)
		assert.NotEmpty(t, validationErrors.ByField("legalName"))
	})

	t.Run("rejects empty string", func(t *testing.T) {
		t.Parallel()

		legalName := ""
		legalNamePtr := &legalName
		req := &iam.UpdateOrganizationRequest{LegalName: &legalNamePtr}
		err := req.Validate()
		require.Error(t, err)

		validationErrors, ok := errors.AsType[validator.ValidationErrors](err)
		require.True(t, ok)
		assert.NotEmpty(t, validationErrors.ByField("legalName"))
	})

	t.Run("allows explicit null", func(t *testing.T) {
		t.Parallel()

		var legalName *string
		req := &iam.UpdateOrganizationRequest{LegalName: &legalName}
		require.NoError(t, req.Validate())
		require.NotNil(t, req.LegalName)
		assert.Nil(t, *req.LegalName)
	})

	t.Run("allows omitted", func(t *testing.T) {
		t.Parallel()

		req := &iam.UpdateOrganizationRequest{}
		require.NoError(t, req.Validate())
		assert.Nil(t, req.LegalName)
	})
}
