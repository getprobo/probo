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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDisplayManagerAutoLoginParsers(t *testing.T) {
	t.Parallel()

	t.Run(
		"gdm",
		func(t *testing.T) {
			t.Parallel()

			assert.True(t, gdmAutoLogin("[daemon]\nAutomaticLoginEnable=True\nAutomaticLogin=jane\n"))
			assert.True(t, gdmAutoLogin("[daemon]\nTimedLoginEnable=true\n"))
			assert.False(t, gdmAutoLogin("[daemon]\n# AutomaticLoginEnable = true\n"))
			assert.False(t, gdmAutoLogin("[security]\nAutomaticLoginEnable=true\n"), "only [daemon] counts")
		},
	)

	t.Run(
		"lightdm",
		func(t *testing.T) {
			t.Parallel()

			assert.True(t, lightdmAutoLogin("[Seat:*]\nautologin-user=jane\n"))
			assert.True(t, lightdmAutoLogin("[SeatDefaults]\nautologin-user=jane\n"))
			assert.False(t, lightdmAutoLogin("[Seat:*]\n#autologin-user=jane\nautologin-user-timeout=0\n"))
			assert.False(
				t,
				lightdmAutoLogin("[Seat:*]\nautologin-user=jane\n[Seat:*]\nautologin-user=\n"),
				"a later file clears it",
			)
		},
	)

	t.Run(
		"sddm",
		func(t *testing.T) {
			t.Parallel()

			assert.True(t, sddmAutoLogin("[Autologin]\nUser=jane\nSession=plasma\n"))
			assert.False(t, sddmAutoLogin("[Autologin]\nUser=\n"))
			assert.False(t, sddmAutoLogin("[Theme]\nCurrent=breeze\n"))
		},
	)

	t.Run(
		"getty",
		func(t *testing.T) {
			t.Parallel()

			assert.True(t, gettyAutoLogin("[Service]\nExecStart=\nExecStart=-/sbin/agetty --autologin jane --noclear %I $TERM\n"))
			assert.True(t, gettyAutoLogin("[Service]\nExecStart=-/sbin/agetty -a root %I 115200\n"))
			assert.False(t, gettyAutoLogin("[Service]\nExecStart=-/sbin/agetty --noclear %I $TERM\n"))
		},
	)
}

func TestDisplayManagerAutoLogin(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	write := func(name, body string) {
		path := filepath.Join(root, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(body), 0o644))
	}

	write("etc/gdm3/custom.conf", "[daemon]\nAutomaticLoginEnable=false\n")
	write("usr/share/lightdm/lightdm.conf.d/50-vendor.conf", "[Seat:*]\nautologin-user=kiosk\n")
	write("etc/lightdm/lightdm.conf.d/90-local.conf", "[Seat:*]\nautologin-user=\n")
	write("etc/sddm.conf.d/autologin.conf", "[Autologin]\nUser=jane\n")

	assert.Equal(t, []string{"sddm"}, displayManagerAutoLogin(root))
	assert.Empty(t, displayManagerAutoLogin(t.TempDir()))
}

func TestCountLinuxAccountsWithoutPassword(t *testing.T) {
	t.Parallel()

	passwd := "root:x:0:0:root:/root:/bin/bash\n" +
		"daemon:x:1:1:daemon:/usr/sbin:/usr/sbin/nologin\n" +
		"jane:x:1000:1000:Jane:/home/jane:/bin/bash\n" +
		"kiosk:x:1001:1001::/home/kiosk:/bin/sh\n" +
		"legacy::1002:1002::/home/legacy:/bin/sh\n" +
		"locked:x:1003:1003::/home/locked:/bin/bash\n" +
		"svc:x:999:999::/var/lib/svc:/bin/false\n"
	shadow := "root:$y$j9T$abc:19000:0:99999:7:::\n" +
		"daemon::19000:0:99999:7:::\n" +
		"jane:$y$j9T$def:19000:0:99999:7:::\n" +
		"kiosk::19000:0:99999:7:::\n" +
		"locked:!:19000:0:99999:7:::\n" +
		"svc::19000:0:99999:7:::\n"

	assert.Equal(
		t,
		2,
		countLinuxAccountsWithoutPassword(passwd, shadow),
		"kiosk (empty shadow) and legacy (empty passwd); nologin and false shells do not count",
	)
	assert.Equal(t, 1, countLinuxAccountsWithoutPassword(passwd, ""), "without shadow only the passwd field is known")
}

func TestCountFreeBSDAccountsWithoutPassword(t *testing.T) {
	t.Parallel()

	master := "# $FreeBSD$\n" +
		"root:$6$abc:0:0::0:0:Charlie &:/root:/bin/sh\n" +
		"toor:*:0:0::0:0:Bourne-again Superuser:/root:\n" +
		"daemon:*:1:1::0:0:Owner of many system processes:/root:/usr/sbin/nologin\n" +
		"kiosk::1001:1001::0:0:Kiosk:/home/kiosk:/bin/sh\n" +
		"nobody::65534:65534::0:0:Unprivileged user:/nonexistent:/usr/sbin/nologin\n"

	assert.Equal(t, 1, countFreeBSDAccountsWithoutPassword(master))
}

func TestFreeBSDGettyAutoLogin(t *testing.T) {
	t.Parallel()

	gettytab := "default:\\\n\t:cb:ce:ck:lc:fd#1000:im=\\r\\n%s/%m (%h) (%t)\\r\\n\\r\\n:sp#1200:\\\n\t:if=/etc/issue:\n" +
		"autologin|al.Pc:\\\n\t:al=root:tc=Pc:\n" +
		"Pc|Pc console:\\\n\t:np:sp#9600:\n"

	assert.True(t, freebsdGettyAutoLogin(gettytab, `ttyv0	"/usr/libexec/getty autologin"	xterm	onifexists secure
`))
	assert.True(t, freebsdGettyAutoLogin(gettytab, `ttyv1	"/usr/libexec/getty al.Pc"	xterm	on  secure
`), "any name of the entry matches")
	assert.False(t, freebsdGettyAutoLogin(gettytab, `ttyv0	"/usr/libexec/getty Pc"	xterm	onifexists secure
ttyv1	"/usr/libexec/getty autologin"	xterm	off secure
#ttyv2	"/usr/libexec/getty autologin"	xterm	on  secure
`))
	assert.False(t, freebsdGettyAutoLogin("Pc|Pc console:\\\n\t:np:sp#9600:\n", `ttyv0	"/usr/libexec/getty Pc"	xterm	on secure`))
}

func TestParsePlistTopLevel(t *testing.T) {
	t.Parallel()

	strs, bools, err := parsePlistTopLevel(`<?xml version="1.0" encoding="UTF-8"?>
<plist version="1.0">
<dict>
	<key>AccountInfo</key>
	<dict>
		<key>autoLoginUser</key>
		<string>nested-must-not-count</string>
		<key>GuestEnabled</key>
		<true/>
	</dict>
	<key>GuestEnabled</key>
	<false/>
	<key>autoLoginUser</key>
	<string>jane</string>
	<key>lastUserName</key>
	<string>jane</string>
</dict>
</plist>`)
	require.NoError(t, err)
	assert.Equal(t, "jane", strs["autoLoginUser"])
	assert.False(t, bools["GuestEnabled"])

	strs, _, err = parsePlistTopLevel("<plist version=\"1.0\"><dict><key>GuestEnabled</key><false/></dict></plist>")
	require.NoError(t, err)
	assert.Empty(t, strs["autoLoginUser"])
}

func TestLoginPasswordResult(t *testing.T) {
	t.Parallel()

	ev := loginPasswordEvidence(nil, 0)
	assert.Equal(t, false, ev["auto_login"])
	assert.Equal(t, []string{}, ev["auto_login_sources"])
	assert.Equal(t, StatusPass, loginPasswordResult(ev, nil, 0).Status)

	assert.Equal(t, StatusFail, loginPasswordResult(ev, []string{"gdm"}, 0).Status)
	assert.Equal(t, StatusFail, loginPasswordResult(ev, nil, 1).Status)

	_, ok := loginPasswordEvidence(nil, -1)["accounts_without_password"]
	assert.False(t, ok, "a platform that cannot count leaves the key out")
}
