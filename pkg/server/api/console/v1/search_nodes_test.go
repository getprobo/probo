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

package console_v1

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/server/api/console/v1/types"
)

type fakeSearchNode struct {
	id gid.GID
}

func (n *fakeSearchNode) IsTaggable() {}

func (n *fakeSearchNode) GetID() gid.GID {
	return n.id
}

func (n *fakeSearchNode) GetTags() *types.TagConnection {
	return nil
}

func TestLoadSearchGroup_MapsInternalErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := log.NewLogger(log.WithOutput(io.Discard))
	scope := coredata.NewScopeFromObjectID(gid.New(gid.NewTenantID(), coredata.OrganizationEntityType))
	id := gid.New(scope.GetTenantID(), coredata.AuditEntityType)

	_, err := loadSearchGroup(
		ctx,
		logger,
		scope,
		[]gid.GID{id},
		func(context.Context, coredata.Scoper, ...gid.GID) ([]*fakeSearchNode, error) {
			return nil, errors.New(`cannot load audits by ids: cannot query audits: ERROR: column "framework_id" does not exist (SQLSTATE 42703)`)
		},
		func(n *fakeSearchNode) types.Taggable {
			return n
		},
	)

	gqlErr, ok := errors.AsType[*gqlerror.Error](err)
	require.True(t, ok)
	assert.Equal(t, "INTERNAL", gqlErr.Extensions["code"])
	assert.Equal(t, "An internal server error occurred. Please try again later.", gqlErr.Message)
	assert.NotContains(t, gqlErr.Message, "framework_id")
	assert.NotContains(t, gqlErr.Message, "SQLSTATE")
}

func TestLoadSearchGroup_MissingResource(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := log.NewLogger(log.WithOutput(io.Discard))
	scope := coredata.NewScopeFromObjectID(gid.New(gid.NewTenantID(), coredata.OrganizationEntityType))
	id := gid.New(scope.GetTenantID(), coredata.AuditEntityType)

	_, err := loadSearchGroup(
		ctx,
		logger,
		scope,
		[]gid.GID{id},
		func(context.Context, coredata.Scoper, ...gid.GID) ([]*fakeSearchNode, error) {
			return []*fakeSearchNode{}, nil
		},
		func(n *fakeSearchNode) types.Taggable {
			return n
		},
	)

	gqlErr, ok := errors.AsType[*gqlerror.Error](err)
	require.True(t, ok)
	assert.Equal(t, "NOT_FOUND", gqlErr.Extensions["code"])
	assert.Equal(t, "resource not found", gqlErr.Message)
}
