// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

package types

import (
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	DeviceOrderBy OrderBy[coredata.DeviceOrderField]

	DeviceConnection struct {
		TotalCount int
		Edges      []*DeviceEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewDeviceConnection(
	p *page.Page[*coredata.Device, coredata.DeviceOrderField],
	parentType any,
	parentID gid.GID,
) *DeviceConnection {
	edges := make([]*DeviceEdge, len(p.Data))
	for i := range edges {
		edges[i] = NewDeviceEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}
	return &DeviceConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),
		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewDeviceEdge(d *coredata.Device, orderBy coredata.DeviceOrderField) *DeviceEdge {
	return &DeviceEdge{
		Cursor: d.CursorKey(orderBy),
		Node:   NewDevice(d),
	}
}

func NewDevice(d *coredata.Device) *Device {
	return &Device{
		ID:           d.ID,
		Hostname:     d.Hostname,
		SerialNumber: d.SerialNumber,
		HardwareUUID: d.HardwareUUID,
		Platform:     d.Platform,
		OsVersion:    d.OSVersion,
		AgentVersion: d.AgentVersion,
		EnrolledAt:   d.EnrolledAt,
		LastSeenAt:   d.LastSeenAt,
		RevokedAt:    d.RevokedAt,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}

func NewDevicePosture(p *coredata.DevicePosture) *DevicePosture {
	return &DevicePosture{
		ID:         p.ID,
		DeviceID:   p.DeviceID,
		CheckKey:   p.CheckKey,
		Status:     p.Status,
		ObservedAt: p.ObservedAt,
	}
}

func NewDevicePostures(ps coredata.DevicePostures) []*DevicePosture {
	out := make([]*DevicePosture, len(ps))
	for i, p := range ps {
		out[i] = NewDevicePosture(p)
	}
	return out
}

func NewDeviceEnrollmentToken(t *coredata.DeviceEnrollmentToken) *DeviceEnrollmentToken {
	return &DeviceEnrollmentToken{
		ID:        t.ID,
		Name:      t.Name,
		ExpiresAt: t.ExpiresAt,
		RevokedAt: t.RevokedAt,
		MaxUses:   t.MaxUses,
		UsedCount: t.UsedCount,
		CreatedAt: t.CreatedAt,
	}
}
