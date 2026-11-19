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

// Package mdm reads managed devices from mobile device management providers
// so that they can be synced into an organization's device list.
package mdm

import (
	"context"
	"errors"
	"time"

	"go.probo.inc/probo/pkg/coredata"
)

// ErrPaginationLimitReached is returned when a provider keeps reporting more
// pages past the hard cap a driver allows itself. Returning a partial list
// would revoke every device past the cap, so the sync fails instead.
var ErrPaginationLimitReached = errors.New("pagination limit reached")

type (
	// Driver lists every device a provider manages for one tenant.
	Driver interface {
		ListDevices(ctx context.Context) ([]DeviceRecord, error)
	}

	// DeviceRecord is one managed device as the provider reports it. Every
	// field but ExternalID is optional: providers leave them empty for
	// devices that have not checked in yet.
	DeviceRecord struct {
		// ExternalID is the provider's stable identifier for the device.
		ExternalID   string
		Name         *string
		SerialNumber *string
		Platform     *coredata.DevicePlatform
		Model        *string
		OSVersion    *string
		// UserEmail is the email of the user the provider assigns the
		// device to, used to set the device owner.
		UserEmail     *string
		LastCheckInAt *time.Time
	}
)
