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

const ubuntuCommonPassword = `#
# /etc/pam.d/common-password - password-related modules common to all services
password	requisite			pam_pwquality.so retry=3
password	[success=1 default=ignore]	pam_unix.so obscure use_authtok try_first_pass yescrypt
password	requisite			pam_deny.so
password	required			pam_permit.so
`

const rhelSystemAuth = `auth        required      pam_env.so
password    requisite     pam_pwquality.so local_users_only minlen=14
password    sufficient    pam_unix.so yescrypt shadow use_authtok
-password   optional      pam_gnome_keyring.so use_authtok
password    required      pam_deny.so
`

func TestParsePAMPasswordRules(t *testing.T) {
	t.Parallel()

	rules := parsePAMPasswordRules(ubuntuCommonPassword)
	require.Len(t, rules, 4)

	assert.Equal(t, "pam_pwquality", rules[0].module)
	assert.Equal(t, "3", rules[0].args["retry"])
	assert.Equal(t, "pam_unix", rules[1].module, "bracketed control must be skipped")
	assert.Contains(t, rules[1].args, "yescrypt")

	rules = parsePAMPasswordRules(rhelSystemAuth)
	require.Len(t, rules, 4, "auth lines are ignored, -password lines are kept")
	assert.Equal(t, "14", rules[0].args["minlen"])
	assert.Equal(t, "pam_gnome_keyring", rules[2].module)

	assert.Empty(t, parsePAMPasswordRules("# password requisite pam_passwdqc.so min=disabled,24,12,8,7\n"))
	assert.Equal(
		t,
		"pam_unix",
		parsePAMPasswordRules("password required /lib/security/pam_unix.so\n")[0].module,
	)
}

func TestPAMPasswordLengths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		pam      string
		conf     string
		expected passwordLengths
	}{
		{
			name:     "pwquality without minlen uses its lowest default",
			pam:      ubuntuCommonPassword,
			expected: passwordLengths{"pam_pwquality": 8, "pam_unix": 6},
		},
		{
			name:     "pwquality.conf sets the length",
			pam:      ubuntuCommonPassword,
			conf:     "# minlen = 8\nminlen = 12\n",
			expected: passwordLengths{"pam_pwquality": 12, "pam_unix": 6},
		},
		{
			name:     "the last pwquality.conf assignment wins",
			pam:      ubuntuCommonPassword,
			conf:     "minlen = 12\nminlen = 10\n",
			expected: passwordLengths{"pam_pwquality": 10, "pam_unix": 6},
		},
		{
			name:     "a module argument overrides pwquality.conf",
			pam:      rhelSystemAuth,
			conf:     "minlen = 9\n",
			expected: passwordLengths{"pam_pwquality": 14, "pam_unix": 6},
		},
		{
			name:     "pam_unix alone uses its built-in minimum",
			pam:      "password required pam_unix.so sha512\n",
			expected: passwordLengths{"pam_unix": 6},
		},
		{
			name:     "explicit pam_unix minlen",
			pam:      "password required pam_unix.so sha512 minlen=10\n",
			expected: passwordLengths{"pam_unix": 10},
		},
		{
			name:     "cracklib without minlen is not counted",
			pam:      "password requisite pam_cracklib.so retry=3\n",
			expected: passwordLengths{},
		},
		{
			name:     "passwdqc uses its shortest enabled minimum",
			pam:      "password requisite pam_passwdqc.so min=disabled,24,12,8,7\n",
			expected: passwordLengths{"pam_passwdqc": 7},
		},
		{
			name:     "no length module",
			pam:      "password required pam_deny.so\n",
			expected: passwordLengths{},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				assert.Equal(
					t,
					tt.expected,
					pamPasswordLengths(parsePAMPasswordRules(tt.pam), tt.conf),
				)
			},
		)
	}
}

func TestPasswordPolicyEvidence(t *testing.T) {
	t.Parallel()

	t.Run(
		"the strictest source decides",
		func(t *testing.T) {
			t.Parallel()

			ev, length := passwordPolicyEvidence(passwordLengths{"pam_pwquality": 12, "pam_unix": 6})
			assert.Equal(t, 12, length)
			assert.Equal(t, "pam_pwquality", ev["backend"])
			assert.Equal(t, 12, ev["min_password_length"])
		},
	)

	t.Run(
		"ties resolve by source name",
		func(t *testing.T) {
			t.Parallel()

			ev, _ := passwordPolicyEvidence(passwordLengths{"secedit": 8, "mdm_device_lock": 8})
			assert.Equal(t, "mdm_device_lock", ev["backend"])
		},
	)

	t.Run(
		"no source reports none",
		func(t *testing.T) {
			t.Parallel()

			ev, length := passwordPolicyEvidence(passwordLengths{})
			assert.Equal(t, 0, length)
			assert.Equal(t, "none", ev["backend"])
			assert.Equal(t, 0, ev["min_password_length"])
		},
	)
}

const darwinPwpolicyOutput = `Getting global account policies
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>policyCategoryAuthentication</key>
	<array>
		<dict>
			<key>policyContent</key>
			<string>policyAttributeFailedAuthentications &lt; policyAttributeMaximumFailedAuthentications</string>
			<key>policyParameters</key>
			<dict>
				<key>policyAttributeMaximumFailedAuthentications</key>
				<integer>10</integer>
			</dict>
		</dict>
	</array>
	<key>policyCategoryPasswordContent</key>
	<array>
		<dict>
			<key>policyContent</key>
			<string>policyAttributePassword matches '.{7,}'</string>
			<key>policyContentDescription</key>
			<dict>
				<key>policyDefaultContentDescription</key>
				<string>Contain at least 7 characters.</string>
			</dict>
			<key>policyIdentifier</key>
			<string>ProfilePayload:3B01C197-EA69-4DAE-A872-5C6AB9816A2B:minLength</string>
		</dict>
		<dict>
			<key>policyContent</key>
			<string>policyAttributePassword matches '.{4,}+'</string>
			<key>policyIdentifier</key>
			<string>com.apple.defaultpasswordpolicy.fde</string>
		</dict>
	</array>
</dict>
</plist>`

func TestParsePwpolicyMinLengths(t *testing.T) {
	t.Parallel()

	t.Run(
		"profile and default rules",
		func(t *testing.T) {
			t.Parallel()

			lengths, err := parsePwpolicyMinLengths(darwinPwpolicyOutput)
			require.NoError(t, err)
			assert.Equal(t, []int{7, 4}, lengths, "the failed-authentications limit is not a length")
		},
	)

	t.Run(
		"anchored pattern and minChars parameter",
		func(t *testing.T) {
			t.Parallel()

			lengths, err := parsePwpolicyMinLengths(
				`<plist><dict><key>policyCategoryPasswordContent</key><array><dict>` +
					`<key>policyContent</key><string>policyAttributePassword matches '^(?=.*[0-9]).{12,}$'</string>` +
					`<key>policyParameters</key><dict><key>minChars</key><integer>12</integer></dict>` +
					`</dict></array></dict></plist>`,
			)
			require.NoError(t, err)
			assert.Equal(t, []int{12, 12}, lengths)
		},
	)

	t.Run(
		"no account policies",
		func(t *testing.T) {
			t.Parallel()

			lengths, err := parsePwpolicyMinLengths("No global account policies have been set.")
			require.NoError(t, err)
			assert.Empty(t, lengths)
		},
	)
}

const freebsdLoginConf = `# login.conf - login class capabilities database.
#
#default:\
#	:minpasswordlen=4:
default:\
	:passwd_format=sha512:\
	:copyright=/etc/COPYRIGHT:\
	:welcome=/var/run/motd:\
	:umask=022:\
	:passwordtime=90d:

standard:\
	:tc=default:
russian|Russian Users Accounts:\
	:minpasswordlen=20:\
	:tc=default:
`

func TestParseLoginConfMinPasswordLen(t *testing.T) {
	t.Parallel()

	_, ok := parseLoginConfMinPasswordLen(freebsdLoginConf)
	assert.False(t, ok, "comments and other classes must not count, nor passwordtime")

	n, ok := parseLoginConfMinPasswordLen(
		"default:\\\n\t:passwd_format=sha512:\\\n\t:minpasswordlen=10:\\\n\t:umask=022:\n",
	)
	require.True(t, ok)
	assert.Equal(t, 10, n)

	n, ok = parseLoginConfMinPasswordLen("default|Default Accounts:\\\n\t:minpasswordlen=12:\n")
	require.True(t, ok)
	assert.Equal(t, 12, n)
}

func TestParseLoginDefsMinLen(t *testing.T) {
	t.Parallel()

	n, ok := parseLoginDefsMinLen("# PASS_MIN_LEN 20\nPASS_MAX_DAYS 99999\nPASS_MIN_LEN\t8\n")
	require.True(t, ok)
	assert.Equal(t, 8, n)

	_, ok = parseLoginDefsMinLen("PASS_MAX_DAYS 99999\n")
	assert.False(t, ok)
}
