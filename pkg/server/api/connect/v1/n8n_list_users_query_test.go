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

package connect_v1_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
	"github.com/vektah/gqlparser/v2/validator"
)

// Keep this document identical to the query in
// packages/n8n-node/nodes/Probo/actions/user/listUsers.operation.ts.
const n8nListUsersQuery = `
	query ListUsers($organizationId: ID!, $first: Int, $after: CursorKey, $orderBy: ProfileOrder, $filter: ProfileFilter) {
		node(id: $organizationId) {
			... on Organization {
				profiles(first: $first, after: $after, orderBy: $orderBy, filter: $filter) {
					edges {
						node {
							id
							fullName
							emailAddress
							source
							state
							additionalEmailAddresses
							kind
							position
							contract {
								start
								end
							}
							createdAt
							updatedAt
							organization { id name }
							membership { id role createdAt }
						}
					}
					pageInfo {
						hasNextPage
						endCursor
					}
				}
			}
		}
	}
`

func TestN8nListUsersQuery_ValidatesAgainstConnectSchema(t *testing.T) {
	t.Parallel()

	// schema.graphql is a generated merge (gitignored). Load the same
	// committed sources gqlgen.yaml uses so CI can run this without Relay.
	sources, err := filepath.Glob("graphql/*.graphql")
	require.NoError(t, err)
	require.NotEmpty(t, sources)

	sources = append(
		sources,
		"../../../gqlutils/directives/authentication/schema.graphql",
		"../../../gqlutils/directives/session/schema.graphql",
	)

	astSources := make([]*ast.Source, 0, len(sources))

	for _, path := range sources {
		schemaSource, err := os.ReadFile(path)
		require.NoError(t, err, "read %s", path)

		astSources = append(astSources, &ast.Source{
			Name:  path,
			Input: string(schemaSource),
		})
	}

	schema, err := gqlparser.LoadSchema(astSources...)
	require.NoError(t, err)

	query, err := parser.ParseQuery(&ast.Source{
		Name:  "listUsers.operation.ts",
		Input: n8nListUsersQuery,
	})
	require.NoError(t, err)

	errs := validator.ValidateWithRules(schema, query, nil)
	assert.Empty(t, errs, errs.Error())
}
