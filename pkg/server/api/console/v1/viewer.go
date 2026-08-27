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

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/api/console/v1/types"
	"go.probo.inc/probo/pkg/server/gqlutils"
)

func (r *viewerResolver) listEmployeeDocuments(
	ctx context.Context,
	organizationID gid.GID,
	first *int,
	after *page.CursorKey,
	last *int,
	before *page.CursorKey,
	orderBy *types.DocumentOrderBy,
	filter *types.EmployeeDocumentFilter,
	mode coredata.EmployeeFilterMode,
	filterMode types.EmployeeDocumentFilterMode,
	listErrorMessage string,
) (*types.EmployeeDocumentConnection, error) {
	scope, err := r.authorize(ctx, organizationID, probo.ActionEmployeeDocumentList)
	if err != nil {
		return nil, err
	}

	identity := authn.IdentityFromContext(ctx)

	documentFilter := coredata.NewDocumentFilter(nil).WithEmployeeIdentityID(&identity.ID, mode)
	if filter != nil {
		documentFilter.WithSigned(filter.Signed)

		if len(filter.ApprovalStates) > 0 {
			documentFilter.WithApprovalStates(filter.ApprovalStates)
		}
	}

	if gqlutils.OnlyTotalCountSelected(ctx) {
		return &types.EmployeeDocumentConnection{
			Resolver: r,
			ParentID: organizationID,
			Filters:  documentFilter,
		}, nil
	}

	pageOrderBy := page.OrderBy[coredata.DocumentOrderField]{
		Field:     coredata.DocumentOrderFieldCreatedAt,
		Direction: page.OrderDirectionDesc,
	}

	if orderBy != nil {
		pageOrderBy = page.OrderBy[coredata.DocumentOrderField]{
			Field:     orderBy.Field,
			Direction: orderBy.Direction,
		}
	}

	cursor := types.NewCursor(first, after, last, before, pageOrderBy)

	documentsPage, err := r.probo.Documents.ListByOrganizationID(
		ctx,
		scope,
		organizationID,
		cursor,
		documentFilter,
	)
	if err != nil {
		r.logger.ErrorCtx(ctx, listErrorMessage, log.Error(err))
		return nil, gqlutils.Internal(ctx)
	}

	employeeDocuments := make([]*types.EmployeeDocument, len(documentsPage.Data))
	for i, doc := range documentsPage.Data {
		employeeDocuments[i] = &types.EmployeeDocument{
			ID:           doc.ID,
			Title:        doc.Title,
			DocumentType: doc.DocumentType,
			CreatedAt:    doc.CreatedAt,
			UpdatedAt:    doc.UpdatedAt,
			FilterMode:   filterMode,
		}
	}

	return types.NewEmployeeDocumentConnection(
		page.NewPage(employeeDocuments, documentsPage.Cursor),
		r,
		organizationID,
		documentFilter,
	), nil
}
