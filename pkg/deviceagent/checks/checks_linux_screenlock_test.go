//go:build linux

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

	"github.com/stretchr/testify/require"
)

func TestLinuxOrderGsettingsSchemas(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		desktop  string
		wantBack []string
	}{
		{
			name:     "cinnamon first",
			desktop:  "cinnamon",
			wantBack: []string{"cinnamon", "gnome", "mate", "ukui"},
		},
		{
			name:     "mate first",
			desktop:  "mate",
			wantBack: []string{"mate", "gnome", "cinnamon", "ukui"},
		},
		{
			name:     "default gnome first",
			desktop:  "ubuntu:gnome",
			wantBack: []string{"gnome", "cinnamon", "mate", "ukui"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := linuxOrderGsettingsSchemas(tt.desktop)
			require.Len(t, got, len(tt.wantBack))

			for i, backend := range tt.wantBack {
				require.Equal(t, backend, got[i].backend)
			}
		})
	}
}

func TestLinuxDesktopPrefers(t *testing.T) {
	t.Parallel()

	require.True(t, linuxDesktopPrefersKDE("KDE"))
	require.True(t, linuxDesktopPrefersKDE("plasma"))
	require.False(t, linuxDesktopPrefersKDE("gnome"))

	require.True(t, linuxDesktopPrefersXFCE("XFCE"))
	require.False(t, linuxDesktopPrefersXFCE("kde"))

	require.True(t, linuxDesktopPrefersI3("i3"))
	require.True(t, linuxDesktopPrefersI3("i3-wm"))
	require.False(t, linuxDesktopPrefersI3("gnome"))
}

func TestParseI3IdleLock(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		config    string
		wantOK    bool
		mechanism string
		idleMins  int
	}{
		{
			name:      "xautolock with i3lock",
			config:    `exec --no-startup-id xautolock -time 10 -locker "i3lock -c 000000"`,
			wantOK:    true,
			mechanism: "xautolock",
			idleMins:  10,
		},
		{
			name:      "xss-lock with i3lock",
			config:    `exec --no-startup-id xss-lock --transfer-sleep-lock -- i3lock -c 000000`,
			wantOK:    true,
			mechanism: "xss-lock",
			idleMins:  -1,
		},
		{
			name:   "manual i3lock bind only",
			config: `bindsym $mod+Shift+x exec i3lock`,
			wantOK: false,
		},
		{
			name:   "xautolock without locker",
			config: `exec xautolock -time 10`,
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			idleMins, _, mechanism, ok := parseI3IdleLock(tt.config)
			require.Equal(t, tt.wantOK, ok)

			if !tt.wantOK {
				return
			}

			require.Equal(t, tt.mechanism, mechanism)
			require.Equal(t, tt.idleMins, idleMins)
		})
	}
}

func TestLinuxScreenLockProbesOrder(t *testing.T) {
	t.Parallel()

	kdeFirst := linuxScreenLockProbes("KDE")
	require.Len(t, kdeFirst, 4)

	xfceFirst := linuxScreenLockProbes("xfce")
	require.Len(t, xfceFirst, 4)

	i3First := linuxScreenLockProbes("i3")
	require.Len(t, i3First, 4)

	defaultFirst := linuxScreenLockProbes("ubuntu")
	require.Len(t, defaultFirst, 4)
}
