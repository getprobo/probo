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

func TestCommonThirdParties_QueryWithLogoURL(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	name := factory.SafeName("CommonTP")
	id := seedCommonThirdParty(t, name)

	const query = `
		query($name: String!) {
			commonThirdParties(name: $name) {
				id
				name
				logoUrl
			}
		}
	`

	var result struct {
		CommonThirdParties []struct {
			ID      string  `json:"id"`
			Name    string  `json:"name"`
			LogoURL *string `json:"logoUrl"`
		} `json:"commonThirdParties"`
	}

	err := owner.Execute(query, map[string]any{"name": name}, &result)
	require.NoError(t, err, "querying commonThirdParties.logoUrl must not surface a resource-not-found error")
	require.Len(t, result.CommonThirdParties, 1)
	assert.Equal(t, id.String(), result.CommonThirdParties[0].ID)
	assert.Equal(t, name, result.CommonThirdParties[0].Name)
	assert.Nil(t, result.CommonThirdParties[0].LogoURL)
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
			id, name, slug, category, certifications, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
	`, id, name, slug, "OTHER", []string{}, now, now)
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
