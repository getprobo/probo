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
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

func NewDevicePostureValue(v coredata.DevicePostureValue) *DevicePostureValue {
	return &DevicePostureValue{
		Kind:   v.Kind,
		Text:   v.Text,
		Number: v.Number,
	}
}

func NewDevicePosture(p *coredata.DevicePosture) *DevicePosture {
	value := coredata.ParseDevicePostureValue(p.CheckKey, p.Evidence)

	return &DevicePosture{
		ID:         p.ID,
		DeviceID:   p.DeviceID,
		CheckKey:   p.CheckKey,
		Status:     p.Status,
		Value:      NewDevicePostureValue(value),
		ObservedAt: p.ObservedAt,
	}
}

func NewDevicePostures(ps coredata.DevicePostures) []*DevicePosture {
	out := make([]*DevicePosture, 0, len(ps))
	for _, p := range ps {
		out = append(out, NewDevicePosture(p))
	}

	return out
}

func NewDevice(d *coredata.Device, postures coredata.DevicePostures) *Device {
	return &Device{
		ID:             d.ID,
		OrganizationID: d.OrganizationID,
		State:          d.State,
		Hostname:       d.Hostname,
		SerialNumber:   d.SerialNumber,
		HardwareUUID:   d.HardwareUUID,
		Platform:       d.Platform,
		OsVersion:      d.OSVersion,
		AgentVersion:   d.AgentVersion,
		OwnerID:        d.OwnerID,
		EnrolledAt:     d.EnrolledAt,
		LastSeenAt:     d.LastSeenAt,
		RevokedAt:      d.RevokedAt,
		LatestPostures: NewDevicePostures(postures),
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
	}
}

func NewListDevicesOutput(
	devicePage *page.Page[*coredata.Device, coredata.DeviceOrderField],
	posturesByDeviceID map[gid.GID]coredata.DevicePostures,
) ListDevicesOutput {
	devices := make([]*Device, 0, len(devicePage.Data))
	for _, d := range devicePage.Data {
		devices = append(devices, NewDevice(d, posturesByDeviceID[d.ID]))
	}

	var nextCursor *page.CursorKey

	if len(devicePage.Data) > 0 {
		cursorKey := devicePage.Data[len(devicePage.Data)-1].CursorKey(devicePage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListDevicesOutput{
		NextCursor: nextCursor,
		Devices:    devices,
	}
}
