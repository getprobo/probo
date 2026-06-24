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

package console_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestCommonThirdParties_QueryWithLogo(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	name := factory.SafeName("CommonTP")
	id := seedCommonThirdParty(t, name)

	const query = `
		query($name: String!) {
			commonThirdParties(name: $name) {
				id
				name
				logo {
					downloadUrl
				}
			}
		}
	`

	var result struct {
		CommonThirdParties []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Logo *struct {
				DownloadURL string `json:"downloadUrl"`
			} `json:"logo"`
		} `json:"commonThirdParties"`
	}

	err := owner.Execute(query, map[string]any{"name": name}, &result)
	require.NoError(t, err, "querying commonThirdParties.logo must not surface a resource-not-found error")
	require.Len(t, result.CommonThirdParties, 1)
	assert.Equal(t, id.String(), result.CommonThirdParties[0].ID)
	assert.Equal(t, name, result.CommonThirdParties[0].Name)
	assert.Nil(t, result.CommonThirdParties[0].Logo)
}

func seedCommonThirdParty(t *testing.T, name string) gid.GID {
	t.Helper()

	ctx := context.Background()
	conn := dialTestPg(t, ctx)
	t.Cleanup(func() { _ = conn.Close(ctx) })

	id := gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType)
	slug := "e2e-" + id.String()
	now := time.Now().UTC()

	_, err := conn.Exec(ctx, `
		INSERT INTO common_third_parties (
			id, name, slug, category, certifications, enrichment_attempts, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
	`, id, name, slug, "OTHER", []string{}, 0, now, now)
	require.NoError(t, err)

	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cleanupConn := dialTestPg(t, cleanupCtx)

		defer func() { _ = cleanupConn.Close(cleanupCtx) }()

		_, err := cleanupConn.Exec(cleanupCtx, `DELETE FROM common_third_parties WHERE id = $1`, id)
		assert.NoError(t, err, "cleanup: cannot delete seeded common third party %s", id)
	})

	return id
}

func dialTestPg(t *testing.T, ctx context.Context) *pgx.Conn {
	t.Helper()

	dsn := os.Getenv("PROBO_E2E_PG_URL")
	if dsn == "" {
		dsn = "postgres://probod:probod@localhost:5432/probod_test?sslmode=disable"
	}

	conn, err := pgx.Connect(ctx, dsn)
	require.NoError(t, err, "cannot connect to e2e test database")

	return conn
}
