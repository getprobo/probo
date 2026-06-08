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

package itam

// ITAM Service Actions
// Format: core:<entity>:<action>
const (
	// Device actions
	ActionDeviceList   = "core:device:list"
	ActionDeviceGet    = "core:device:get"
	ActionDeviceRevoke = "core:device:revoke"
	ActionDeviceAssign = "core:device:assign"

	// DevicePosture actions
	ActionDevicePostureList = "core:device-posture:list"

	// DeviceEnrollmentToken actions
	ActionDeviceEnrollmentTokenList   = "core:device-enrollment-token:list"
	ActionDeviceEnrollmentTokenGet    = "core:device-enrollment-token:get"
	ActionDeviceEnrollmentTokenCreate = "core:device-enrollment-token:create"
	ActionDeviceEnrollmentTokenRevoke = "core:device-enrollment-token:revoke"
)
