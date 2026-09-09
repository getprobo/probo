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

func TestParseWindowsFirewallProfiles(t *testing.T) {
	t.Parallel()

	t.Run(
		"all profiles enabled",
		func(t *testing.T) {
			t.Parallel()

			profiles, allEnabled := parseWindowsFirewallProfiles(
				"Domain=True;Private=True;Public=True",
			)
			require.Equal(
				t,
				map[string]string{
					"Domain":  "True",
					"Private": "True",
					"Public":  "True",
				},
				profiles,
			)
			assert.True(t, allEnabled)
		},
	)

	t.Run(
		"one profile disabled",
		func(t *testing.T) {
			t.Parallel()

			profiles, allEnabled := parseWindowsFirewallProfiles(
				"Domain=True;Private=False;Public=True",
			)
			require.Equal(
				t,
				map[string]string{
					"Domain":  "True",
					"Private": "False",
					"Public":  "True",
				},
				profiles,
			)
			assert.False(t, allEnabled)
		},
	)

	t.Run(
		"empty output is not enabled",
		func(t *testing.T) {
			t.Parallel()

			profiles, allEnabled := parseWindowsFirewallProfiles("")
			require.Empty(t, profiles)
			assert.False(t, allEnabled)
		},
	)
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
