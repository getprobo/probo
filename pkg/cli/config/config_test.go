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
