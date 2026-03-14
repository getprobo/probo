// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/cli/config"
)

func TestDefaultHost(t *testing.T) {
	t.Run(
		"PROBO_TOKEN uses active host instead of first alphabetically",
		func(t *testing.T) {
			t.Setenv("PROBO_TOKEN", "tok_test")
			t.Setenv("PROBO_HOST", "")

			cfg := &config.Config{
				ActiveHost: "beta.probo.inc",
				Hosts: map[string]*config.HostConfig{
					"alpha.probo.inc": {
						Token:        "old-alpha",
						Organization: "org-alpha",
					},
					"beta.probo.inc": {
						Token:        "old-beta",
						Organization: "org-beta",
					},
				},
			}

			host, hc, err := cfg.DefaultHost()
			require.NoError(t, err)
			assert.Equal(t, "beta.probo.inc", host)
			assert.Equal(t, "tok_test", hc.Token)
			assert.Equal(t, "org-beta", hc.Organization)
		},
	)

	t.Run(
		"PROBO_TOKEN falls back to first host when no active host",
		func(t *testing.T) {
			t.Setenv("PROBO_TOKEN", "tok_test")
			t.Setenv("PROBO_HOST", "")

			cfg := &config.Config{
				Hosts: map[string]*config.HostConfig{
					"alpha.probo.inc": {
						Token:        "old",
						Organization: "org-alpha",
					},
					"beta.probo.inc": {
						Token:        "old",
						Organization: "org-beta",
					},
				},
			}

			host, hc, err := cfg.DefaultHost()
			require.NoError(t, err)
			assert.Equal(t, "alpha.probo.inc", host)
			assert.Equal(t, "tok_test", hc.Token)
			assert.Equal(t, "org-alpha", hc.Organization)
		},
	)

	t.Run(
		"PROBO_HOST takes precedence over everything",
		func(t *testing.T) {
			t.Setenv("PROBO_HOST", "custom.probo.inc")
			t.Setenv("PROBO_TOKEN", "tok_env")

			cfg := &config.Config{
				ActiveHost: "beta.probo.inc",
				Hosts: map[string]*config.HostConfig{
					"beta.probo.inc": {
						Token:        "old",
						Organization: "org-beta",
					},
				},
			}

			host, hc, err := cfg.DefaultHost()
			require.NoError(t, err)
			assert.Equal(t, "custom.probo.inc", host)
			assert.Equal(t, "tok_env", hc.Token)
		},
	)

	t.Run(
		"active host is used when no env vars set",
		func(t *testing.T) {
			t.Setenv("PROBO_HOST", "")
			t.Setenv("PROBO_TOKEN", "")

			cfg := &config.Config{
				ActiveHost: "beta.probo.inc",
				Hosts: map[string]*config.HostConfig{
					"alpha.probo.inc": {
						Token:        "tok-alpha",
						Organization: "org-alpha",
					},
					"beta.probo.inc": {
						Token:        "tok-beta",
						Organization: "org-beta",
					},
				},
			}

			host, hc, err := cfg.DefaultHost()
			require.NoError(t, err)
			assert.Equal(t, "beta.probo.inc", host)
			assert.Equal(t, "tok-beta", hc.Token)
			assert.Equal(t, "org-beta", hc.Organization)
		},
	)
}
