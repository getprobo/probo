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

package types

import (
	"encoding/json"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

const (
	AccessInventoryChangeEventParentOrganization = "organization"
	AccessInventoryChangeEventParentGrant        = "grant"
)

type (
	AccessInventoryGrantOrderBy       OrderBy[coredata.AccessInventoryGrantOrderField]
	AccessInventoryPrincipalOrderBy   OrderBy[coredata.AccessInventoryPrincipalOrderField]
	AccessInventoryResourceOrderBy    OrderBy[coredata.AccessInventoryResourceOrderField]
	AccessInventoryChangeEventOrderBy OrderBy[coredata.AccessInventoryChangeEventOrderField]

	AccessInventoryGrantConnection struct {
		TotalCount int
		Edges      []*AccessInventoryGrantEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		Filter   *coredata.AccessInventoryGrantFilter
	}

	AccessInventoryPrincipalConnection struct {
		TotalCount int
		Edges      []*AccessInventoryPrincipalEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		Filter   *coredata.AccessInventoryPrincipalFilter
	}

	AccessInventoryResourceConnection struct {
		TotalCount int
		Edges      []*AccessInventoryResourceEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		Filter   *coredata.AccessInventoryResourceFilter
	}

	AccessInventoryChangeEventConnection struct {
		TotalCount int
		Edges      []*AccessInventoryChangeEventEdge
		PageInfo   PageInfo

		Resolver   any
		ParentID   gid.GID
		ParentKind string
		Filter     *coredata.AccessInventoryChangeEventFilter
	}
)

func NewAccessInventoryGrantConnection(
	p *page.Page[*coredata.AccessInventoryGrant, coredata.AccessInventoryGrantOrderField],
	parentType any,
	parentID gid.GID,
	filter *coredata.AccessInventoryGrantFilter,
) *AccessInventoryGrantConnection {
	edges := make([]*AccessInventoryGrantEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewAccessInventoryGrantEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &AccessInventoryGrantConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
		Filter:   filter,
	}
}

func NewAccessInventoryGrantEdge(
	g *coredata.AccessInventoryGrant,
	orderBy coredata.AccessInventoryGrantOrderField,
) *AccessInventoryGrantEdge {
	return &AccessInventoryGrantEdge{
		Cursor: g.CursorKey(orderBy),
		Node:   NewAccessInventoryGrant(g),
	}
}

func NewAccessInventoryGrant(g *coredata.AccessInventoryGrant) *AccessInventoryGrant {
	return &AccessInventoryGrant{
		ID: g.ID,
		Organization: &Organization{
			ID: g.OrganizationID,
		},
		Source: &AccessReviewSource{
			ID: g.AccessReviewSourceID,
		},
		Principal: &AccessInventoryPrincipal{
			ID: g.PrincipalID,
		},
		Resource: &AccessInventoryResource{
			ID: g.ResourceID,
		},
		PrincipalKey:            g.PrincipalKey,
		ResourceKey:             g.ResourceKey,
		EntitlementKey:          g.EntitlementKey,
		EntitlementDisplayName:  g.EntitlementDisplayName,
		Presence:                g.Presence,
		ConsecutiveMissingSyncs: g.ConsecutiveMissingSyncs,
		FirstSeenAt:             g.FirstSeenAt,
		LastSeenAt:              g.LastSeenAt,
		ProviderChangedAt:       g.ProviderChangedAt,
		CreatedAt:               g.CreatedAt,
		UpdatedAt:               g.UpdatedAt,
	}
}

func NewAccessInventoryPrincipalConnection(
	p *page.Page[*coredata.AccessInventoryPrincipal, coredata.AccessInventoryPrincipalOrderField],
	parentType any,
	parentID gid.GID,
	filter *coredata.AccessInventoryPrincipalFilter,
) *AccessInventoryPrincipalConnection {
	edges := make([]*AccessInventoryPrincipalEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewAccessInventoryPrincipalEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &AccessInventoryPrincipalConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
		Filter:   filter,
	}
}

func NewAccessInventoryPrincipalEdge(
	p *coredata.AccessInventoryPrincipal,
	orderBy coredata.AccessInventoryPrincipalOrderField,
) *AccessInventoryPrincipalEdge {
	return &AccessInventoryPrincipalEdge{
		Cursor: p.CursorKey(orderBy),
		Node:   NewAccessInventoryPrincipal(p),
	}
}

func NewAccessInventoryPrincipal(p *coredata.AccessInventoryPrincipal) *AccessInventoryPrincipal {
	return &AccessInventoryPrincipal{
		ID: p.ID,
		Organization: &Organization{
			ID: p.OrganizationID,
		},
		Source: &AccessReviewSource{
			ID: p.AccessReviewSourceID,
		},
		PrincipalKey:     p.PrincipalKey,
		Email:            p.Email,
		FullName:         p.FullName,
		ExternalID:       p.ExternalID,
		JobTitle:         p.JobTitle,
		IsAdmin:          p.IsAdmin,
		MfaStatus:        p.MFAStatus,
		AuthMethod:       p.AuthMethod,
		AccountType:      p.AccountType,
		Active:           p.Active,
		LastLogin:        p.LastLogin,
		AccountCreatedAt: p.AccountCreatedAt,
		FirstSeenAt:      p.FirstSeenAt,
		LastSeenAt:       p.LastSeenAt,
		CreatedAt:        p.CreatedAt,
		UpdatedAt:        p.UpdatedAt,
	}
}

func NewAccessInventoryResourceConnection(
	p *page.Page[*coredata.AccessInventoryResource, coredata.AccessInventoryResourceOrderField],
	parentType any,
	parentID gid.GID,
	filter *coredata.AccessInventoryResourceFilter,
) *AccessInventoryResourceConnection {
	edges := make([]*AccessInventoryResourceEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewAccessInventoryResourceEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &AccessInventoryResourceConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
		Filter:   filter,
	}
}

func NewAccessInventoryResourceEdge(
	r *coredata.AccessInventoryResource,
	orderBy coredata.AccessInventoryResourceOrderField,
) *AccessInventoryResourceEdge {
	return &AccessInventoryResourceEdge{
		Cursor: r.CursorKey(orderBy),
		Node:   NewAccessInventoryResource(r),
	}
}

func NewAccessInventoryResource(r *coredata.AccessInventoryResource) *AccessInventoryResource {
	return &AccessInventoryResource{
		ID: r.ID,
		Organization: &Organization{
			ID: r.OrganizationID,
		},
		Source: &AccessReviewSource{
			ID: r.AccessReviewSourceID,
		},
		ResourceKey: r.ResourceKey,
		DisplayName: r.DisplayName,
		Kind:        r.Kind,
		FirstSeenAt: r.FirstSeenAt,
		LastSeenAt:  r.LastSeenAt,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

func NewAccessInventoryChangeEventConnection(
	p *page.Page[*coredata.AccessInventoryChangeEvent, coredata.AccessInventoryChangeEventOrderField],
	parentType any,
	parentID gid.GID,
	parentKind string,
	filter *coredata.AccessInventoryChangeEventFilter,
) *AccessInventoryChangeEventConnection {
	edges := make([]*AccessInventoryChangeEventEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewAccessInventoryChangeEventEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &AccessInventoryChangeEventConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver:   parentType,
		ParentID:   parentID,
		ParentKind: parentKind,
		Filter:     filter,
	}
}

func NewAccessInventoryChangeEventEdge(
	e *coredata.AccessInventoryChangeEvent,
	orderBy coredata.AccessInventoryChangeEventOrderField,
) *AccessInventoryChangeEventEdge {
	return &AccessInventoryChangeEventEdge{
		Cursor: e.CursorKey(orderBy),
		Node:   NewAccessInventoryChangeEvent(e),
	}
}

func NewAccessInventoryChangeEvent(e *coredata.AccessInventoryChangeEvent) *AccessInventoryChangeEvent {
	event := &AccessInventoryChangeEvent{
		ID: e.ID,
		Organization: &Organization{
			ID: e.OrganizationID,
		},
		Source: &AccessReviewSource{
			ID: e.AccessReviewSourceID,
		},
		PrincipalKey:       e.PrincipalKey,
		ResourceKey:        e.ResourceKey,
		EntitlementKey:     e.EntitlementKey,
		ChangeType:         e.ChangeType,
		ObservedAt:         e.ObservedAt,
		ProviderOccurredAt: e.ProviderOccurredAt,
		EffectiveAt:        e.EffectiveAt(),
		Before:             newAccessInventoryGrantAttributeSnapshot(e.Before),
		After:              newAccessInventoryGrantAttributeSnapshot(e.After),
		CreatedAt:          e.CreatedAt,
	}

	if e.GrantID != nil {
		event.Grant = &AccessInventoryGrant{ID: *e.GrantID}
	}

	if e.SyncRunID != nil {
		event.SyncRun = &AccessSourceSyncRun{ID: *e.SyncRunID}
	}

	return event
}

func newAccessInventoryGrantAttributeSnapshot(
	raw *json.RawMessage,
) *AccessInventoryGrantAttributeSnapshot {
	snapshot := coredata.DecodeAccessInventoryGrantAttributeSnapshot(raw)
	if snapshot == nil {
		return nil
	}

	return &AccessInventoryGrantAttributeSnapshot{
		EntitlementDisplayName: snapshot.EntitlementDisplayName,
		Presence:               snapshot.Presence,
	}
}
