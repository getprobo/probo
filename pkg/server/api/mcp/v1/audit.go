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

package mcp_v1

import (
	"context"
	"fmt"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/server/api/mcp/v1/types"
)

// mcpgen comments out Resolver methods in schema.resolvers.go that the spec does not name.

func auditIDs(audits []*coredata.Audit) []gid.GID {
	ids := make([]gid.GID, 0, len(audits))
	for _, audit := range audits {
		ids = append(ids, audit.ID)
	}

	return ids
}

func (r *Resolver) auditFrameworkIDs(
	ctx context.Context,
	scope coredata.Scoper,
	ids []gid.GID,
) (map[gid.GID][]gid.GID, error) {
	frameworkIDs, err := r.proboSvc.Audits.ListFrameworkIDsByAuditIDs(ctx, scope, ids)
	if err != nil {
		return nil, fmt.Errorf("cannot list audit frameworks: %w", err)
	}

	return frameworkIDs, nil
}

func (r *Resolver) newAudit(
	ctx context.Context,
	scope coredata.Scoper,
	audit *coredata.Audit,
	file *coredata.File,
) (*types.Audit, error) {
	frameworkIDs, err := r.auditFrameworkIDs(ctx, scope, []gid.GID{audit.ID})
	if err != nil {
		return nil, err
	}

	return types.NewAudit(audit, file, types.FrameworkIDsFor(frameworkIDs, audit.ID)), nil
}
