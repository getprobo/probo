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
	"maps"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/server/api/console/v1/types"
	"go.probo.inc/probo/pkg/server/gqlx"
	"go.probo.inc/probo/pkg/task"
)

// searchNodes loads every search hit through an authorized loader. The
// search index is organization-wide, so each resource must still pass its
// own get permission before its node is returned.
func (r *queryResolver) searchNodes(
	ctx context.Context,
	results []*coredata.SearchResult,
) ([]types.Taggable, error) {
	grouped := map[uint16][]gid.GID{}

	for _, result := range results {
		entityType := result.ID.EntityType()
		grouped[entityType] = append(grouped[entityType], result.ID)
	}

	loaded := map[gid.GID]types.Taggable{}

	for entityType, ids := range grouped {
		nodes, err := r.loadSearchNodeGroup(ctx, entityType, ids)
		if err != nil {
			return nil, err
		}

		maps.Copy(loaded, nodes)
	}

	nodes := make([]types.Taggable, len(results))
	for i, result := range results {
		node, ok := loaded[result.ID]
		if !ok {
			return nil, gqlx.NotFoundf(ctx, "resource not found")
		}

		nodes[i] = node
	}

	return nodes, nil
}

func (r *queryResolver) loadSearchNodeGroup(
	ctx context.Context,
	entityType uint16,
	ids []gid.GID,
) (map[gid.GID]types.Taggable, error) {
	switch entityType {
	case coredata.DocumentEntityType:
		return authorizeSearchGroup(ctx, r, probo.ActionDocumentGet, ids, func(ctx context.Context, scope *coredata.Scope) (map[gid.GID]types.Taggable, error) {
			return loadSearchGroup(ctx, ids, func(ctx context.Context, ids ...gid.GID) ([]*coredata.Document, error) {
				return r.probo.Documents.GetByIDs(ctx, scope, ids...)
			}, func(item *coredata.Document) gid.GID { return item.ID }, func(item *coredata.Document) types.Taggable {
				return types.NewDocument(item)
			})
		})
	case coredata.TaskEntityType:
		return authorizeSearchGroup(ctx, r, task.ActionTaskGet, ids, func(ctx context.Context, scope *coredata.Scope) (map[gid.GID]types.Taggable, error) {
			return loadSearchGroup(ctx, ids, func(ctx context.Context, ids ...gid.GID) ([]*coredata.Task, error) {
				return r.task.GetByIDs(ctx, scope, ids...)
			}, func(item *coredata.Task) gid.GID { return item.ID }, func(item *coredata.Task) types.Taggable {
				return types.NewTask(item)
			})
		})
	case coredata.RiskEntityType:
		return authorizeSearchGroup(ctx, r, probo.ActionRiskGet, ids, func(ctx context.Context, scope *coredata.Scope) (map[gid.GID]types.Taggable, error) {
			return loadSearchGroup(ctx, ids, func(ctx context.Context, ids ...gid.GID) ([]*coredata.Risk, error) {
				return r.probo.Risks.GetByIDs(ctx, scope, ids...)
			}, func(item *coredata.Risk) gid.GID { return item.ID }, func(item *coredata.Risk) types.Taggable {
				return types.NewRisk(item)
			})
		})
	case coredata.ThirdPartyEntityType:
		return authorizeSearchGroup(ctx, r, probo.ActionThirdPartyGet, ids, func(ctx context.Context, scope *coredata.Scope) (map[gid.GID]types.Taggable, error) {
			return loadSearchGroup(ctx, ids, func(ctx context.Context, ids ...gid.GID) ([]*coredata.ThirdParty, error) {
				return r.probo.ThirdParties.GetByIDs(ctx, scope, ids...)
			}, func(item *coredata.ThirdParty) gid.GID { return item.ID }, func(item *coredata.ThirdParty) types.Taggable {
				return types.NewThirdParty(item)
			})
		})
	case coredata.FrameworkEntityType:
		return authorizeSearchGroup(ctx, r, probo.ActionFrameworkGet, ids, func(ctx context.Context, scope *coredata.Scope) (map[gid.GID]types.Taggable, error) {
			return loadSearchGroup(ctx, ids, func(ctx context.Context, ids ...gid.GID) ([]*coredata.Framework, error) {
				return r.probo.Frameworks.GetByIDs(ctx, scope, ids...)
			}, func(item *coredata.Framework) gid.GID { return item.ID }, func(item *coredata.Framework) types.Taggable {
				return types.NewFramework(item)
			})
		})
	case coredata.InternalControlEntityType:
		return authorizeSearchGroup(ctx, r, probo.ActionInternalControlGet, ids, func(ctx context.Context, scope *coredata.Scope) (map[gid.GID]types.Taggable, error) {
			return loadSearchGroup(ctx, ids, func(ctx context.Context, ids ...gid.GID) ([]*coredata.InternalControl, error) {
				return r.probo.InternalControls.GetByIDs(ctx, scope, ids...)
			}, func(item *coredata.InternalControl) gid.GID { return item.ID }, func(item *coredata.InternalControl) types.Taggable {
				return types.NewInternalControl(item)
			})
		})
	case coredata.AuditEntityType:
		return authorizeSearchGroup(ctx, r, probo.ActionAuditList, ids, func(ctx context.Context, scope *coredata.Scope) (map[gid.GID]types.Taggable, error) {
			return loadSearchGroup(ctx, ids, func(ctx context.Context, ids ...gid.GID) ([]*coredata.Audit, error) {
				return r.probo.Audits.GetByIDs(ctx, scope, ids...)
			}, func(item *coredata.Audit) gid.GID { return item.ID }, func(item *coredata.Audit) types.Taggable {
				return types.NewAudit(item)
			})
		})
	case coredata.ControlEntityType:
		return authorizeSearchGroup(ctx, r, probo.ActionControlList, ids, func(ctx context.Context, scope *coredata.Scope) (map[gid.GID]types.Taggable, error) {
			return loadSearchGroup(ctx, ids, func(ctx context.Context, ids ...gid.GID) ([]*coredata.Control, error) {
				return r.probo.Controls.GetByIDs(ctx, scope, ids...)
			}, func(item *coredata.Control) gid.GID { return item.ID }, func(item *coredata.Control) types.Taggable {
				return types.NewControl(item)
			})
		})
	default:
		nodes := make(map[gid.GID]types.Taggable, len(ids))
		for _, id := range ids {
			node, err := r.Node(ctx, id)
			if err != nil {
				return nil, err
			}

			taggable, ok := node.(types.Taggable)
			if !ok || taggable == nil {
				return nil, gqlx.Internal(ctx)
			}

			nodes[id] = taggable
		}

		return nodes, nil
	}
}

func authorizeSearchGroup(
	ctx context.Context,
	r *queryResolver,
	action iam.Action,
	ids []gid.GID,
	load func(context.Context, *coredata.Scope) (map[gid.GID]types.Taggable, error),
) (map[gid.GID]types.Taggable, error) {
	scope, err := r.batchAuthorize(ctx, action, ids)
	if err != nil {
		return nil, err
	}

	return load(ctx, scope)
}

func loadSearchGroup[T any](
	ctx context.Context,
	ids []gid.GID,
	load func(context.Context, ...gid.GID) ([]*T, error),
	idOf func(*T) gid.GID,
	wrap func(*T) types.Taggable,
) (map[gid.GID]types.Taggable, error) {
	items, err := load(ctx, ids...)
	if err != nil {
		return nil, err
	}

	nodes := make(map[gid.GID]types.Taggable, len(items))
	for _, item := range items {
		nodes[idOf(item)] = wrap(item)
	}

	for _, id := range ids {
		if _, ok := nodes[id]; !ok {
			return nil, gqlx.NotFoundf(ctx, "resource not found")
		}
	}

	return nodes, nil
}
