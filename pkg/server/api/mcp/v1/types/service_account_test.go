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

package types_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/server/api/mcp/v1/types"
)

func TestServiceAccountTypes_OmitSecretsAndDeletionState(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	now := time.Now()

	t.Run(
		"service account omits deletion timestamp",
		func(t *testing.T) {
			t.Parallel()

			account := types.NewServiceAccount(
				&coredata.ServiceAccount{
					ID:             gid.New(tenantID, coredata.ServiceAccountEntityType),
					OrganizationID: gid.New(tenantID, coredata.OrganizationEntityType),
					Name:           "Automation",
					Scopes:         coredata.OAuth2Scopes{"v1:iam:read"},
					DeletedAt:      &now,
					CreatedAt:      now,
					UpdatedAt:      now,
				},
			)

			data, err := json.Marshal(account)
			require.NoError(t, err)

			var fields map[string]any
			require.NoError(t, json.Unmarshal(data, &fields))
			assert.NotContains(t, fields, "deleted_at")
		},
	)

	t.Run(
		"credential omits hashed and raw tokens",
		func(t *testing.T) {
			t.Parallel()

			credential := types.NewServiceAccountCredential(
				&coredata.ServiceAccountCredential{
					ID:               gid.New(tenantID, coredata.ServiceAccountCredentialEntityType),
					ServiceAccountID: gid.New(tenantID, coredata.ServiceAccountEntityType),
					Name:             "Deployment",
					HashedToken:      []byte("secret"),
					Scopes:           coredata.OAuth2Scopes{"v1:iam:read"},
					ExpiresAt:        now.Add(time.Hour),
					CreatedAt:        now,
				},
			)

			data, err := json.Marshal(credential)
			require.NoError(t, err)

			var fields map[string]any
			require.NoError(t, json.Unmarshal(data, &fields))
			assert.NotContains(t, fields, "hashed_token")
			assert.NotContains(t, fields, "token")
		},
	)
}
