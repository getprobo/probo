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
	"fmt"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	NotificationConnection struct {
		TotalCount int
		Edges      []*NotificationEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		Filter   *coredata.NotificationFilter
	}
)

func NewNotification(
	notification *coredata.Notification,
) (*Notification, error) {
	title, err := notification.Title()
	if err != nil {
		return nil, fmt.Errorf("cannot read notification title: %w", err)
	}

	return &Notification{
		ID:        notification.ID,
		Type:      notification.Type,
		SubjectID: notification.SubjectID,
		Title:     title,
		Status:    notification.Status,
		ReadAt:    notification.ReadAt,
		CreatedAt: notification.CreatedAt,
	}, nil
}

func NewNotificationConnection(
	p *page.Page[*coredata.Notification, coredata.NotificationOrderField],
	parentType any,
	parentID gid.GID,
	filter *coredata.NotificationFilter,
) (*NotificationConnection, error) {
	edges := make([]*NotificationEdge, len(p.Data))

	for i, notification := range p.Data {
		node, err := NewNotification(notification)
		if err != nil {
			return nil, err
		}

		edges[i] = &NotificationEdge{
			Cursor: notification.CursorKey(p.Cursor.OrderBy.Field),
			Node:   node,
		}
	}

	return &NotificationConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
		Filter:   filter,
	}, nil
}
