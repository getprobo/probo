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

package coredata

import (
	"encoding"
	"fmt"
)

type CompliancePortalCommitmentIcon string

const (
	CompliancePortalCommitmentIconLockKey       CompliancePortalCommitmentIcon = "LOCK_KEY"
	CompliancePortalCommitmentIconEyeSlash      CompliancePortalCommitmentIcon = "EYE_SLASH"
	CompliancePortalCommitmentIconFingerprint   CompliancePortalCommitmentIcon = "FINGERPRINT"
	CompliancePortalCommitmentIconShieldWarning CompliancePortalCommitmentIcon = "SHIELD_WARNING"
	CompliancePortalCommitmentIconShieldCheck   CompliancePortalCommitmentIcon = "SHIELD_CHECK"
	CompliancePortalCommitmentIconSiren         CompliancePortalCommitmentIcon = "SIREN"
	CompliancePortalCommitmentIconKey           CompliancePortalCommitmentIcon = "KEY"
	CompliancePortalCommitmentIconLock          CompliancePortalCommitmentIcon = "LOCK"
	CompliancePortalCommitmentIconCloud         CompliancePortalCommitmentIcon = "CLOUD"
	CompliancePortalCommitmentIconDatabase      CompliancePortalCommitmentIcon = "DATABASE"
	CompliancePortalCommitmentIconGlobe         CompliancePortalCommitmentIcon = "GLOBE"
	CompliancePortalCommitmentIconEye           CompliancePortalCommitmentIcon = "EYE"
	CompliancePortalCommitmentIconUsers         CompliancePortalCommitmentIcon = "USERS"
	CompliancePortalCommitmentIconCertificate   CompliancePortalCommitmentIcon = "CERTIFICATE"
	CompliancePortalCommitmentIconGavel         CompliancePortalCommitmentIcon = "GAVEL"
	CompliancePortalCommitmentIconHeartbeat     CompliancePortalCommitmentIcon = "HEARTBEAT"
	CompliancePortalCommitmentIconBell          CompliancePortalCommitmentIcon = "BELL"
	CompliancePortalCommitmentIconBug           CompliancePortalCommitmentIcon = "BUG"
	CompliancePortalCommitmentIconCode          CompliancePortalCommitmentIcon = "CODE"
	CompliancePortalCommitmentIconServer        CompliancePortalCommitmentIcon = "SERVER"
)

var (
	_ fmt.Stringer             = CompliancePortalCommitmentIcon("")
	_ encoding.TextMarshaler   = CompliancePortalCommitmentIcon("")
	_ encoding.TextUnmarshaler = (*CompliancePortalCommitmentIcon)(nil)
)

func CompliancePortalCommitmentIcons() []CompliancePortalCommitmentIcon {
	return []CompliancePortalCommitmentIcon{
		CompliancePortalCommitmentIconLockKey,
		CompliancePortalCommitmentIconEyeSlash,
		CompliancePortalCommitmentIconFingerprint,
		CompliancePortalCommitmentIconShieldWarning,
		CompliancePortalCommitmentIconShieldCheck,
		CompliancePortalCommitmentIconSiren,
		CompliancePortalCommitmentIconKey,
		CompliancePortalCommitmentIconLock,
		CompliancePortalCommitmentIconCloud,
		CompliancePortalCommitmentIconDatabase,
		CompliancePortalCommitmentIconGlobe,
		CompliancePortalCommitmentIconEye,
		CompliancePortalCommitmentIconUsers,
		CompliancePortalCommitmentIconCertificate,
		CompliancePortalCommitmentIconGavel,
		CompliancePortalCommitmentIconHeartbeat,
		CompliancePortalCommitmentIconBell,
		CompliancePortalCommitmentIconBug,
		CompliancePortalCommitmentIconCode,
		CompliancePortalCommitmentIconServer,
	}
}

func (v CompliancePortalCommitmentIcon) IsValid() bool {
	switch v {
	case
		CompliancePortalCommitmentIconLockKey,
		CompliancePortalCommitmentIconEyeSlash,
		CompliancePortalCommitmentIconFingerprint,
		CompliancePortalCommitmentIconShieldWarning,
		CompliancePortalCommitmentIconShieldCheck,
		CompliancePortalCommitmentIconSiren,
		CompliancePortalCommitmentIconKey,
		CompliancePortalCommitmentIconLock,
		CompliancePortalCommitmentIconCloud,
		CompliancePortalCommitmentIconDatabase,
		CompliancePortalCommitmentIconGlobe,
		CompliancePortalCommitmentIconEye,
		CompliancePortalCommitmentIconUsers,
		CompliancePortalCommitmentIconCertificate,
		CompliancePortalCommitmentIconGavel,
		CompliancePortalCommitmentIconHeartbeat,
		CompliancePortalCommitmentIconBell,
		CompliancePortalCommitmentIconBug,
		CompliancePortalCommitmentIconCode,
		CompliancePortalCommitmentIconServer:
		return true
	}

	return false
}

func (v CompliancePortalCommitmentIcon) String() string {
	return string(v)
}

func (v CompliancePortalCommitmentIcon) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *CompliancePortalCommitmentIcon) UnmarshalText(text []byte) error {
	val := CompliancePortalCommitmentIcon(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid CompliancePortalCommitmentIcon value: %q", string(text))
	}

	*v = val

	return nil
}
