// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
		OwnerID  *gid.GID
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

func NewOwnedDeviceConnection(
	p *page.Page[*coredata.Device, coredata.DeviceOrderField],
	parentType any,
	parentID gid.GID,
	ownerID gid.GID,
) *DeviceConnection {
	conn := NewDeviceConnection(p, parentType, parentID)
	conn.OwnerID = &ownerID

	return conn
}

func NewDeviceEdge(d *coredata.Device, orderBy coredata.DeviceOrderField) *DeviceEdge {
	return &DeviceEdge{
		Cursor: d.CursorKey(orderBy),
		Node:   NewDevice(d),
	}
}

func NewDevice(d *coredata.Device) *Device {
	device := &Device{
		ID:           d.ID,
		State:        d.State,
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
	if d.OwnerID != nil {
		device.Owner = &Profile{ID: *d.OwnerID}
	}

	return device
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
