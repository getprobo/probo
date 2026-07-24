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

package coredata

import (
	"encoding"
	"fmt"
)

// DevicePostureCheckKey is a string identifier for a single posture check.
// New keys can be added freely without a database migration; the column is a
// plain TEXT.
type DevicePostureCheckKey string

const (
	DevicePostureCheckKeyDiskEncryption    DevicePostureCheckKey = "DISK_ENCRYPTION"
	DevicePostureCheckKeyScreenLock        DevicePostureCheckKey = "SCREEN_LOCK"
	DevicePostureCheckKeyFirewallEnabled   DevicePostureCheckKey = "FIREWALL_ENABLED"
	DevicePostureCheckKeyTimeSync          DevicePostureCheckKey = "TIME_SYNC"
	DevicePostureCheckKeyOSVersion         DevicePostureCheckKey = "OS_VERSION"
	DevicePostureCheckKeyAutoUpdate        DevicePostureCheckKey = "AUTO_UPDATE"
	DevicePostureCheckKeyPasswordPolicy    DevicePostureCheckKey = "PASSWORD_POLICY"
	DevicePostureCheckKeyRemoteLogin       DevicePostureCheckKey = "REMOTE_LOGIN"
	DevicePostureCheckKeyMalwareProtection DevicePostureCheckKey = "MALWARE_PROTECTION"
)

var (
	_ fmt.Stringer             = DevicePostureCheckKey("")
	_ encoding.TextMarshaler   = DevicePostureCheckKey("")
	_ encoding.TextUnmarshaler = (*DevicePostureCheckKey)(nil)
)

func DevicePostureCheckKeys() []DevicePostureCheckKey {
	return []DevicePostureCheckKey{
		DevicePostureCheckKeyDiskEncryption,
		DevicePostureCheckKeyScreenLock,
		DevicePostureCheckKeyFirewallEnabled,
		DevicePostureCheckKeyTimeSync,
		DevicePostureCheckKeyOSVersion,
		DevicePostureCheckKeyAutoUpdate,
		DevicePostureCheckKeyPasswordPolicy,
		DevicePostureCheckKeyRemoteLogin,
		DevicePostureCheckKeyMalwareProtection,
	}
}

func (v DevicePostureCheckKey) IsValid() bool {
	switch v {
	case
		DevicePostureCheckKeyDiskEncryption,
		DevicePostureCheckKeyScreenLock,
		DevicePostureCheckKeyFirewallEnabled,
		DevicePostureCheckKeyTimeSync,
		DevicePostureCheckKeyOSVersion,
		DevicePostureCheckKeyAutoUpdate,
		DevicePostureCheckKeyPasswordPolicy,
		DevicePostureCheckKeyRemoteLogin,
		DevicePostureCheckKeyMalwareProtection:
		return true
	}

	return false
}

func (v DevicePostureCheckKey) String() string {
	return string(v)
}

func (v DevicePostureCheckKey) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *DevicePostureCheckKey) UnmarshalText(text []byte) error {
	val := DevicePostureCheckKey(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid DevicePostureCheckKey value: %q", string(text))
	}

	*v = val

	return nil
}
