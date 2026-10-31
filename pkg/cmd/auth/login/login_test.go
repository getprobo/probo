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

package login

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/cli/config"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/cmd/iostreams"
)

func TestHostOptions(t *testing.T) {
	t.Parallel()

	options := hostOptions()
	require.Len(t, options, 3)
	assert.Equal(t, "Probo US (us.probo.com)", options[0].Key)
	assert.Equal(t, regionUS, options[0].Value)
	assert.Equal(t, "Probo EU (eu.probo.com)", options[1].Key)
	assert.Equal(t, regionEU, options[1].Value)
	assert.Equal(t, "Custom", options[2].Key)
	assert.Equal(t, regionCustom, options[2].Value)
}

func TestLoginHost(t *testing.T) {
	t.Parallel()

	t.Run(
		"uses an explicit hostname without prompting",
		func(t *testing.T) {
			t.Parallel()

			called := false
			host, err := loginHost(
				"probo.example.com",
				true,
				func() (string, error) {
					called = true
					return hostEU, nil
				},
			)
			require.NoError(t, err)
			assert.Equal(t, "probo.example.com", host)
			assert.False(t, called)
		},
	)

	t.Run(
		"trims an explicit hostname",
		func(t *testing.T) {
			t.Parallel()

			called := false
			host, err := loginHost(
				"  us.probo.com  ",
				false,
				func() (string, error) {
					called = true
					return hostEU, nil
				},
			)
			require.NoError(t, err)
			assert.Equal(t, hostUS, host)
			assert.False(t, called)
		},
	)

	t.Run(
		"rejects a blank hostname flag",
		func(t *testing.T) {
			t.Parallel()

			called := false
			host, err := loginHost(
				"   ",
				false,
				func() (string, error) {
					called = true
					return hostEU, nil
				},
			)
			require.Error(t, err)
			assert.Empty(t, host)
			assert.False(t, called)
			assert.Contains(t, err.Error(), "hostname is required")
		},
	)

	t.Run(
		"prompts for a host when the session is interactive",
		func(t *testing.T) {
			t.Parallel()

			host, err := loginHost(
				"",
				true,
				func() (string, error) {
					return hostUS, nil
				},
			)
			require.NoError(t, err)
			assert.Equal(t, hostUS, host)
		},
	)

	t.Run(
		"trims a prompted hostname",
		func(t *testing.T) {
			t.Parallel()

			host, err := loginHost(
				"",
				true,
				func() (string, error) {
					return "  eu.probo.com  ", nil
				},
			)
			require.NoError(t, err)
			assert.Equal(t, hostEU, host)
		},
	)

	t.Run(
		"rejects an empty prompted hostname",
		func(t *testing.T) {
			t.Parallel()

			host, err := loginHost(
				"",
				true,
				func() (string, error) {
					return "   ", nil
				},
			)
			require.Error(t, err)
			assert.Empty(t, host)
			assert.EqualError(t, err, "hostname is required")
		},
	)

	t.Run(
		"requires a hostname when the session is not interactive",
		func(t *testing.T) {
			t.Parallel()

			called := false
			host, err := loginHost(
				"",
				false,
				func() (string, error) {
					called = true
					return hostEU, nil
				},
			)
			require.Error(t, err)
			assert.Empty(t, host)
			assert.False(t, called)
			assert.Contains(t, err.Error(), "us.probo.com")
			assert.Contains(t, err.Error(), "eu.probo.com")
			assert.Contains(t, err.Error(), "custom")
		},
	)
}

func TestNewCmdLogin_HostnameFlagHasNoDefault(t *testing.T) {
	t.Parallel()

	ios, _, _ := iostreams.Test()
	f := &cmdutil.Factory{IOStreams: ios}
	cmd := NewCmdLogin(f)

	flag := cmd.Flags().Lookup("hostname")
	require.NotNil(t, flag)

	assert.Empty(t, flag.DefValue)
	assert.Contains(t, flag.Usage, "hostname")
	assert.Contains(t, flag.Usage, "eu.probo.com")
	assert.Contains(t, flag.Usage, "us.probo.com")
}

func TestNewCmdLogin_NonInteractiveRequiresHostname(t *testing.T) {
	t.Parallel()

	ios, _, errOut := iostreams.Test()
	ios.ForceNonInteractive = true

	f := &cmdutil.Factory{
		IOStreams: ios,
		Config: func() (*config.Config, error) {
			t.Fatal("login continued without a hostname")
			return nil, nil
		},
	}

	cmd := NewCmdLogin(f)
	cmd.SetOut(io.Discard)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "hostname is required")
	assert.NotContains(t, errOut.String(), "Discovering")
}
