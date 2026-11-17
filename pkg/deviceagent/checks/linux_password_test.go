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
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinuxPasswordPolicyFromSources(t *testing.T) {
	t.Parallel()

	pwquality := "password\trequisite\tpam_pwquality.so retry=3\n"
	tests := []struct {
		name      string
		conf      string
		dropins   []linuxPwqualityDropin
		pam       []string
		loginDefs string
		want      Status
		source    string
		length    int
		known     bool
		detail    string
	}{
		{
			name:   "pwquality minlen is the enforced length",
			conf:   "minlen = 14\n",
			pam:    []string{pwquality},
			want:   StatusPass,
			source: linuxPasswordSourcePwqualityConf,
			length: 14,
			known:  true,
		},
		{
			name:   "commented minlen uses the pam_pwquality default",
			conf:   "# minlen = 8\n",
			pam:    []string{pwquality},
			want:   StatusPass,
			source: linuxPasswordSourcePwquality,
			length: linuxPwqualityDefaultMinLen,
			known:  true,
		},
		{
			name: "pwquality.conf overrides drop-ins",
			conf: "minlen = 14\n",
			dropins: []linuxPwqualityDropin{
				{Name: "20-later.conf", Body: "minlen = 6\n"},
				{Name: "10-early.conf", Body: "minlen = 12\n"},
			},
			pam:    []string{pwquality},
			want:   StatusPass,
			source: linuxPasswordSourcePwqualityConf,
			length: 14,
			known:  true,
		},
		{
			name: "last drop-in applies when pwquality.conf omits minlen",
			conf: "dcredit = -1\n",
			dropins: []linuxPwqualityDropin{
				{Name: "20-later.conf", Body: "minlen = 7\n"},
				{Name: "10-early.conf", Body: "minlen = 12\n"},
			},
			pam:    []string{pwquality},
			want:   StatusFail,
			source: linuxPasswordSourcePwqualityDir,
			length: 7,
			known:  true,
		},
		{
			name: "broken pwquality.conf is unknown despite a drop-in",
			conf: "minlen = no\n",
			dropins: []linuxPwqualityDropin{
				{Name: "10-local.conf", Body: "minlen = 14\n"},
			},
			pam:    []string{pwquality},
			want:   StatusUnknown,
			known:  false,
			detail: "parse_error",
		},
		{
			name:   "pwquality minlen below its floor reads as 6",
			conf:   "minlen = 0\n",
			pam:    []string{pwquality},
			want:   StatusFail,
			source: linuxPasswordSourcePwqualityConf,
			length: linuxPwqualityBaseMinLen,
			known:  true,
		},
		{
			name:   "pam argument overrides pwquality.conf",
			conf:   "minlen = 14\n",
			pam:    []string{"password requisite pam_pwquality.so minlen=7 retry=3\n"},
			want:   StatusFail,
			source: linuxPasswordSourcePAM,
			length: 7,
			known:  true,
		},
		{
			name:   "pam argument below the pwquality floor reads as 6",
			pam:    []string{"password requisite pam_pwquality.so minlen=-1\n"},
			want:   StatusFail,
			source: linuxPasswordSourcePAM,
			length: linuxPwqualityBaseMinLen,
			known:  true,
		},
		{
			name:   "inline comment in pam does not set minlen",
			conf:   "minlen = 14\n",
			pam:    []string{"password requisite pam_pwquality.so retry=3 # minlen=4\n"},
			want:   StatusPass,
			source: linuxPasswordSourcePwqualityConf,
			length: 14,
			known:  true,
		},
		{
			name:   "continued pam line keeps its arguments",
			pam:    []string{"password requisite pam_pwquality.so \\\n  minlen=12\n"},
			want:   StatusPass,
			source: linuxPasswordSourcePAM,
			length: 12,
			known:  true,
		},
		{
			name: "optional quality module does not enforce a length",
			pam: []string{
				"password optional pam_pwquality.so minlen=14\n" +
					"password required pam_unix.so obscure\n",
			},
			loginDefs: "PASS_MIN_LEN 0\n",
			want:      StatusFail,
			source:    linuxPasswordSourceLoginDefs,
			length:    0,
			known:     true,
		},
		{
			name: "sufficient quality module does not enforce a length",
			pam: []string{
				"password sufficient pam_pwquality.so minlen=14\n",
			},
			want:   StatusFail,
			source: linuxPasswordSourceNone,
			length: 0,
			known:  true,
		},
		{
			name: "bracket that ignores failure does not enforce a length",
			pam: []string{
				"password [success=1 default=ignore] pam_pwquality.so minlen=14\n",
			},
			want:   StatusFail,
			source: linuxPasswordSourceNone,
			length: 0,
			known:  true,
		},
		{
			name: "bracket without a default rejects failure",
			pam: []string{
				"password [success=ok] pam_pwquality.so minlen=12\n",
			},
			want:   StatusPass,
			source: linuxPasswordSourcePAM,
			length: 12,
			known:  true,
		},
		{
			name: "bracket authtok_err beats default",
			pam: []string{
				"password [authtok_err=die default=ignore] pam_pwquality.so minlen=12\n",
			},
			want:   StatusPass,
			source: linuxPasswordSourcePAM,
			length: 12,
			known:  true,
		},
		{
			name: "only enforcing modules set the strictest length",
			pam: []string{
				"password optional pam_cracklib.so minlen=20\n" +
					"password requisite pam_pwquality.so minlen=9\n",
			},
			want:   StatusPass,
			source: linuxPasswordSourcePAM,
			length: 9,
			known:  true,
		},
		{
			name:   "pam argument overrides a broken pwquality.conf",
			conf:   "minlen = no\n",
			pam:    []string{"password requisite pam_pwquality.so minlen=14\n"},
			want:   StatusPass,
			source: linuxPasswordSourcePAM,
			length: 14,
			known:  true,
		},
		{
			name: "drop-in without minlen keeps pwquality.conf",
			conf: "minlen = 14\n",
			dropins: []linuxPwqualityDropin{
				{Name: "10-credits.conf", Body: "dcredit = -1\n"},
			},
			pam:    []string{pwquality},
			want:   StatusPass,
			source: linuxPasswordSourcePwqualityConf,
			length: 14,
			known:  true,
		},
		{
			name:   "pam argument of 8 passes",
			conf:   "minlen = 4\n",
			pam:    []string{"-password requisite pam_pwquality.so minlen=8\n"},
			want:   StatusPass,
			source: linuxPasswordSourcePAM,
			length: 8,
			known:  true,
		},
		{
			name: "bracketed control that rejects failure reads minlen",
			pam: []string{
				"password [success=1 default=bad] pam_pwquality.so minlen=12\n",
			},
			want:   StatusPass,
			source: linuxPasswordSourcePAM,
			length: 12,
			known:  true,
		},
		{
			name: "module path still matches",
			pam: []string{
				"password required /usr/lib/security/pam_pwquality.so minlen=9\n",
			},
			want:   StatusPass,
			source: linuxPasswordSourcePAM,
			length: 9,
			known:  true,
		},
		{
			name: "auth line and include do not set the length",
			pam: []string{
				"auth required pam_pwquality.so minlen=4\n" +
					"password include common-password\n" +
					"password requisite pam_pwquality.so minlen=14\n",
			},
			want:   StatusPass,
			source: linuxPasswordSourcePAM,
			length: 14,
			known:  true,
		},
		{
			name: "commented module falls back to login.defs",
			pam: []string{
				"# password requisite pam_pwquality.so minlen=14\n" +
					"password required pam_unix.so obscure\n",
			},
			loginDefs: "PASS_MIN_LEN\t0\n",
			want:      StatusFail,
			source:    linuxPasswordSourceLoginDefs,
			length:    0,
			known:     true,
		},
		{
			name:      "login.defs zero without pam fails",
			loginDefs: "PASS_MIN_LEN 0\n",
			want:      StatusFail,
			source:    linuxPasswordSourceLoginDefs,
			length:    0,
			known:     true,
		},
		{
			name:      "login.defs at the threshold passes",
			loginDefs: "PASS_MIN_LEN 8\n",
			want:      StatusPass,
			source:    linuxPasswordSourceLoginDefs,
			length:    8,
			known:     true,
		},
		{
			name:   "unset login.defs without pam is unknown",
			want:   StatusUnknown,
			known:  false,
			detail: "note",
		},
		{
			name: "pam stack without a length module and no login.defs fails",
			pam: []string{
				"password required pam_unix.so obscure use_authtok\n",
			},
			want:   StatusFail,
			source: linuxPasswordSourceNone,
			length: 0,
			known:  true,
		},
		{
			name:      "weak login.defs is ignored when pam_pwquality is loaded",
			conf:      "minlen = 14\n",
			pam:       []string{pwquality},
			loginDefs: "PASS_MIN_LEN 0\n",
			want:      StatusPass,
			source:    linuxPasswordSourcePwqualityConf,
			length:    14,
			known:     true,
		},
		{
			name:   "invalid pwquality minlen is unknown",
			conf:   "minlen = later\n",
			pam:    []string{pwquality},
			want:   StatusUnknown,
			known:  false,
			detail: "parse_error",
		},
		{
			name:      "invalid login.defs length is unknown",
			loginDefs: "PASS_MIN_LEN nope\n",
			want:      StatusUnknown,
			known:     false,
			detail:    "parse_error",
		},
		{
			name:      "negative login.defs length is unknown",
			loginDefs: "PASS_MIN_LEN -1\n",
			want:      StatusUnknown,
			known:     false,
			detail:    "parse_error",
		},
		{
			name: "weakest pam stack wins",
			pam: []string{
				"password requisite pam_pwquality.so minlen=14\n",
				"password requisite pam_pwquality.so minlen=7\n",
			},
			want:   StatusFail,
			source: linuxPasswordSourcePAM,
			length: 7,
			known:  true,
		},
		{
			name:   "cracklib default is 9",
			pam:    []string{"password requisite pam_cracklib.so retry=3\n"},
			want:   StatusPass,
			source: linuxPasswordSourceCracklib,
			length: linuxCracklibDefaultMinLen,
			known:  true,
		},
		{
			name:   "cracklib argument below the threshold fails",
			pam:    []string{"password requisite pam_cracklib.so minlen=6\n"},
			want:   StatusFail,
			source: linuxPasswordSourcePAM,
			length: 6,
			known:  true,
		},
		{
			name: "both modules must be satisfied",
			conf: "minlen = 8\n",
			pam: []string{
				"password requisite pam_pwquality.so retry=3\n" +
					"password requisite pam_cracklib.so minlen=10\n",
			},
			want:   StatusPass,
			source: linuxPasswordSourcePAM,
			length: 10,
			known:  true,
		},
		{
			name: "a stack without a module weakens a strict one",
			conf: "minlen = 14\n",
			pam: []string{
				"password requisite pam_pwquality.so retry=3\n",
				"password required pam_unix.so obscure\n",
			},
			loginDefs: "PASS_MIN_LEN 0\n",
			want:      StatusFail,
			source:    linuxPasswordSourceLoginDefs,
			length:    0,
			known:     true,
		},
		{
			name:   "inline comment does not hide minlen",
			conf:   "minlen = 11 # site policy\ndcredit = -1\n",
			pam:    []string{pwquality},
			want:   StatusPass,
			source: linuxPasswordSourcePwqualityConf,
			length: 11,
			known:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := linuxPasswordPolicyFromSources(tt.conf, tt.dropins, tt.pam, tt.loginDefs)
			assert.Equal(t, tt.want, got.Status)

			if !tt.known {
				assert.NotContains(t, got.Evidence, "pass_min_len_value")

				if tt.detail != "" {
					assert.Contains(t, got.Evidence, tt.detail)
				}

				return
			}

			assert.Equal(t, tt.source, got.Evidence["source"])
			assert.Equal(t, tt.length, got.Evidence["pass_min_len_value"])
			assert.Equal(t, strconv.Itoa(tt.length), got.Evidence["pass_min_len"])
		})
	}
}

func TestLinuxPasswordPolicyFromSources_KeepsMaxDays(t *testing.T) {
	t.Parallel()

	got := linuxPasswordPolicyFromSources(
		"",
		nil,
		nil,
		"PASS_MIN_LEN 8\nPASS_MAX_DAYS 90\n",
	)

	assert.Equal(t, StatusPass, got.Status)
	assert.Equal(t, "90", got.Evidence["pass_max_days"])
}

func TestReadOptionalFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "login.defs")
	require.NoError(t, os.WriteFile(path, []byte("PASS_MIN_LEN 8\n"), 0o644))

	body, present, err := readOptionalFile(path)
	require.NoError(t, err)
	assert.True(t, present)
	assert.Equal(t, "PASS_MIN_LEN 8\n", body)

	_, present, err = readOptionalFile(filepath.Join(dir, "missing"))
	require.NoError(t, err)
	assert.False(t, present)
}

func TestReadPwqualityDropins(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "20-later.conf"), []byte("minlen = 6\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "10-early.conf"), []byte("minlen = 14\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("minlen = 1\n"), 0o644))
	require.NoError(t, os.Mkdir(filepath.Join(dir, "skip.conf"), 0o755))

	dropins, err := readPwqualityDropins(dir)
	require.NoError(t, err)
	require.Len(t, dropins, 2)
	assert.Equal(t, "10-early.conf", dropins[0].Name)
	assert.Equal(t, "20-later.conf", dropins[1].Name)

	missing, err := readPwqualityDropins(filepath.Join(dir, "absent"))
	require.NoError(t, err)
	assert.Empty(t, missing)
}
