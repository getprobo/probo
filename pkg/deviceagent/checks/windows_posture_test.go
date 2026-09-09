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

package checks

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseWindowsBitLockerVolumes(t *testing.T) {
	t.Parallel()

	t.Run(
		"operating system volume protected",
		func(t *testing.T) {
			t.Parallel()

			volumes, allProtected := parseWindowsBitLockerVolumes("C:=On")
			require.Equal(t, map[string]string{"C:": "On"}, volumes)
			assert.True(t, allProtected)
		},
	)

	t.Run(
		"operating system volume unprotected at full encryption",
		func(t *testing.T) {
			t.Parallel()

			volumes, allProtected := parseWindowsBitLockerVolumes("C:=Off")
			require.Equal(t, map[string]string{"C:": "Off"}, volumes)
			assert.False(t, allProtected)
		},
	)

	t.Run(
		"empty output is unprotected",
		func(t *testing.T) {
			t.Parallel()

			volumes, allProtected := parseWindowsBitLockerVolumes("")
			require.Empty(t, volumes)
			assert.False(t, allProtected)
		},
	)
}

func TestWindowsFirewallOn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		raw           string
		expectedOn    bool
		expectedKnown bool
	}{
		{
			name:          "all profiles enabled",
			raw:           "Domain=True;Private=True;Public=True",
			expectedOn:    true,
			expectedKnown: true,
		},
		{
			name:          "one profile disabled",
			raw:           "Domain=True;Private=False;Public=True",
			expectedKnown: true,
		},
		{
			name: "empty output is unknown",
			raw:  "",
		},
		{
			name: "localized output is unknown, not disabled",
			raw:  "Domaine=Actif;Privé=Actif;Public=Actif",
		},
		{
			name: "null COM properties are unknown, not disabled",
			raw:  "Domain=;Private=;Public=",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				on, known := windowsFirewallOn(parseWindowsJoinedPairs(tt.raw))
				assert.Equal(t, tt.expectedOn, on)
				assert.Equal(t, tt.expectedKnown, known)
			},
		)
	}
}

func TestWindowsScreenLockOn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		values         map[string]string
		expectedSource string
		expectedOn     bool
		expectedKnown  bool
	}{
		{
			// "Do not display the lock screen" swaps the glance screen for the
			// credential prompt; it must not veto a host that does lock.
			name:           "a hidden lock screen does not override the inactivity limit",
			values:         map[string]string{"NoLockScreen": "1", "InactivityTimeoutSecs": "900"},
			expectedSource: "machine_inactivity_limit",
			expectedOn:     true,
			expectedKnown:  true,
		},
		{
			name:           "machine inactivity limit enforces lock",
			values:         map[string]string{"InactivityTimeoutSecs": "900"},
			expectedSource: "machine_inactivity_limit",
			expectedOn:     true,
			expectedKnown:  true,
		},
		{
			name:           "mdm device lock enforces lock",
			values:         map[string]string{"MaxInactivityTimeDeviceLock": "15"},
			expectedSource: "mdm_device_lock",
			expectedOn:     true,
			expectedKnown:  true,
		},
		{
			name: "secure and active screensaver with timeout enforces lock",
			values: map[string]string{
				"ScreenSaverIsSecure": "1",
				"ScreenSaveActive":    "1",
				"ScreenSaveTimeOut":   "600",
			},
			expectedSource: "machine_policy",
			expectedOn:     true,
			expectedKnown:  true,
		},
		{
			name: "secure but inactive screensaver does not enforce lock",
			values: map[string]string{
				"ScreenSaverIsSecure": "1",
				"ScreenSaveActive":    "0",
				"ScreenSaveTimeOut":   "600",
			},
			expectedSource: "machine_policy",
			expectedKnown:  true,
		},
		{
			name: "secure and active screensaver without timeout does not enforce lock",
			values: map[string]string{
				"ScreenSaverIsSecure": "1",
				"ScreenSaveActive":    "1",
				"ScreenSaveTimeOut":   "0",
			},
			expectedSource: "machine_policy",
			expectedKnown:  true,
		},
		{
			name:           "zero inactivity limit falls through to the screensaver policy",
			values:         map[string]string{"InactivityTimeoutSecs": "0", "ScreenSaverIsSecure": "0"},
			expectedSource: "machine_policy",
			expectedKnown:  true,
		},
		{
			name:   "no machine policy is unknown",
			values: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				source, on, known := windowsScreenLockOn(tt.values)
				assert.Equal(t, tt.expectedSource, source)
				assert.Equal(t, tt.expectedOn, on)
				assert.Equal(t, tt.expectedKnown, known)
			},
		)
	}
}

func TestParseWindowsJoinedPairs(t *testing.T) {
	t.Parallel()

	t.Run(
		"malformed segment skipped",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(
				t,
				map[string]string{"Domain": "True", "Public": "True"},
				parseWindowsJoinedPairs("Domain=True;not-a-pair;Public=True"),
			)
		},
	)

	t.Run(
		"whitespace around keys and values",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(
				t,
				map[string]string{"C:": "On", "D:": "Off"},
				parseWindowsJoinedPairs(" C: = On ; D: = Off "),
			)
		},
	)

	t.Run(
		"trailing semicolon",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(
				t,
				map[string]string{"Domain": "True"},
				parseWindowsJoinedPairs("Domain=True;"),
			)
		},
	)
}

func TestWindowsTimeSyncOn(t *testing.T) {
	t.Parallel()

	t.Run(
		"manual NTP service is on",
		func(t *testing.T) {
			t.Parallel()

			assert.True(t, windowsTimeSyncOn("3", "NTP"))
		},
	)

	t.Run(
		"disabled service is off",
		func(t *testing.T) {
			t.Parallel()

			assert.False(t, windowsTimeSyncOn("4", "NTP"))
		},
	)

	t.Run(
		"manual NoSync service is off",
		func(t *testing.T) {
			t.Parallel()

			assert.False(t, windowsTimeSyncOn("3", "NoSync"))
		},
	)

	t.Run(
		"automatic NT5DS service is on",
		func(t *testing.T) {
			t.Parallel()

			assert.True(t, windowsTimeSyncOn("2", "NT5DS"))
		},
	)
}

func TestWindowsAutoUpdateOn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		noAutoUpdate  string
		auOptions     string
		serviceStart  string
		expectedOn    bool
		expectedKnown bool
	}{
		{
			name:          "default trigger-start service is on",
			serviceStart:  "3",
			expectedOn:    true,
			expectedKnown: true,
		},
		{
			name:          "automatic service and install policy are on",
			auOptions:     "4",
			serviceStart:  "2",
			expectedOn:    true,
			expectedKnown: true,
		},
		{
			name:          "disabled by policy is off",
			noAutoUpdate:  "1",
			serviceStart:  "3",
			expectedKnown: true,
		},
		{
			name:          "notify-only policy is off",
			auOptions:     "2",
			serviceStart:  "3",
			expectedKnown: true,
		},
		{
			name:          "disabled service is off",
			auOptions:     "4",
			serviceStart:  "4",
			expectedKnown: true,
		},
		{
			name:          "missing service state is unknown",
			expectedOn:    false,
			expectedKnown: false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				on, known := windowsAutoUpdateOn(
					tt.noAutoUpdate,
					tt.auOptions,
					tt.serviceStart,
				)
				assert.Equal(t, tt.expectedOn, on)
				assert.Equal(t, tt.expectedKnown, known)
			},
		)
	}
}
