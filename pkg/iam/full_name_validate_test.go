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

package iam

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/validator"
)

func TestUpdateIdentityRequest_Validate_TrimsAndRejectsWhitespace(t *testing.T) {
	t.Parallel()

	t.Run("trims surrounding whitespace", func(t *testing.T) {
		t.Parallel()

		req := &UpdateIdentityRequest{FullName: "  Ada Lovelace  "}
		require.NoError(t, req.Validate())
		assert.Equal(t, "Ada Lovelace", req.FullName)
	})

	t.Run("rejects whitespace-only", func(t *testing.T) {
		t.Parallel()

		req := &UpdateIdentityRequest{FullName: " \t "}
		err := req.Validate()
		require.Error(t, err)

		validationErrors, ok := errors.AsType[validator.ValidationErrors](err)
		require.True(t, ok)
		assert.NotEmpty(t, validationErrors.ByField("full_name"))
		assert.Equal(t, "", req.FullName)
	})
}

func TestCreateIdentityWithPasswordRequest_Validate_TrimsAndRejectsWhitespace(t *testing.T) {
	t.Parallel()

	t.Run("trims surrounding whitespace", func(t *testing.T) {
		t.Parallel()

		req := &CreateIdentityWithPasswordRequest{
			FullName: "  Ada Lovelace  ",
			Password: "a-strong-password-123",
		}
		require.NoError(t, req.Validate())
		assert.Equal(t, "Ada Lovelace", req.FullName)
	})

	t.Run("rejects whitespace-only", func(t *testing.T) {
		t.Parallel()

		req := &CreateIdentityWithPasswordRequest{
			FullName: "   ",
			Password: "a-strong-password-123",
		}
		err := req.Validate()
		require.Error(t, err)

		validationErrors, ok := errors.AsType[validator.ValidationErrors](err)
		require.True(t, ok)
		assert.NotEmpty(t, validationErrors.ByField("fullName"))
		assert.Equal(t, "", req.FullName)
	})
}

func TestCreateUserRequest_Validate_TrimsAndRejectsWhitespace(t *testing.T) {
	t.Parallel()

	orgID := gid.New(gid.NewTenantID(), coredata.OrganizationEntityType)

	t.Run("trims surrounding whitespace", func(t *testing.T) {
		t.Parallel()

		req := &CreateUserRequest{
			OrganizationID: orgID,
			FullName:       "  Ada Lovelace  ",
		}
		require.NoError(t, req.Validate())
		assert.Equal(t, "Ada Lovelace", req.FullName)
	})

	t.Run("rejects whitespace-only", func(t *testing.T) {
		t.Parallel()

		req := &CreateUserRequest{
			OrganizationID: orgID,
			FullName:       " \n ",
		}
		err := req.Validate()
		require.Error(t, err)

		validationErrors, ok := errors.AsType[validator.ValidationErrors](err)
		require.True(t, ok)
		assert.NotEmpty(t, validationErrors.ByField("full_name"))
		assert.Equal(t, "", req.FullName)
	})
}

func TestUpdateUserRequest_Validate_TrimsAndRejectsWhitespace(t *testing.T) {
	t.Parallel()

	profileID := gid.New(gid.NewTenantID(), coredata.MembershipProfileEntityType)

	t.Run("trims surrounding whitespace", func(t *testing.T) {
		t.Parallel()

		req := &UpdateUserRequest{
			ID:       profileID,
			FullName: "  Ada Lovelace  ",
		}
		require.NoError(t, req.Validate())
		assert.Equal(t, "Ada Lovelace", req.FullName)
	})

	t.Run("rejects whitespace-only", func(t *testing.T) {
		t.Parallel()

		req := &UpdateUserRequest{
			ID:       profileID,
			FullName: "\t  ",
		}
		err := req.Validate()
		require.Error(t, err)

		validationErrors, ok := errors.AsType[validator.ValidationErrors](err)
		require.True(t, ok)
		assert.NotEmpty(t, validationErrors.ByField("full_name"))
		assert.Equal(t, "", req.FullName)
	})
}
