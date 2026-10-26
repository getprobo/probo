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

package cmdutil_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/cli/config"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/cmd/iostreams"
)

func TestPrintList(t *testing.T) {
	t.Parallel()

	view := cmdutil.ListView[string]{
		Empty:   "No frameworks found.",
		Noun:    "frameworks",
		Headers: []string{"ID"},
		Row: func(id string) ([]string, error) {
			return []string{id}, nil
		},
	}

	t.Run("json null becomes an empty array", func(t *testing.T) {
		t.Parallel()

		var out, errOut bytes.Buffer

		err := cmdutil.PrintList(streams(&out, &errOut), cmdutil.OutputJSON, []string(nil), 0, view)
		require.NoError(t, err)
		assert.JSONEq(t, "[]", out.String())
		assert.Empty(t, errOut.String())
	})

	t.Run("empty table", func(t *testing.T) {
		t.Parallel()

		var out, errOut bytes.Buffer

		err := cmdutil.PrintList(streams(&out, &errOut), cmdutil.OutputTable, []string{}, 0, view)
		require.NoError(t, err)
		assert.Equal(t, "No frameworks found.\n", out.String())
		assert.Empty(t, errOut.String())
	})

	t.Run("truncation", func(t *testing.T) {
		t.Parallel()

		var out, errOut bytes.Buffer

		err := cmdutil.PrintList(streams(&out, &errOut), cmdutil.OutputTable, []string{"fw_1"}, 4, view)
		require.NoError(t, err)
		assert.Contains(t, out.String(), "fw_1")
		assert.Equal(t, "\nShowing 1 of 4 frameworks\n", errOut.String())
	})

	t.Run("row error prints nothing", func(t *testing.T) {
		t.Parallel()

		failing := view
		failing.Row = func(string) ([]string, error) {
			return nil, errors.New("cannot format description")
		}

		var out, errOut bytes.Buffer

		err := cmdutil.PrintList(streams(&out, &errOut), cmdutil.OutputTable, []string{"fw_1"}, 1, failing)
		require.EqualError(t, err, "cannot format description")
		assert.Empty(t, out.String())
		assert.Empty(t, errOut.String())
	})
}

func TestGraphQLClient(t *testing.T) {
	t.Setenv("PROBO_HOST", "")
	t.Setenv("PROBO_TOKEN", "")

	f := &cmdutil.Factory{
		Config: func() (*config.Config, error) {
			return &config.Config{
				ActiveHost: "https://example.test",
				Hosts: map[string]*config.HostConfig{
					"https://example.test": {Token: "token", Organization: "org_1"},
				},
			}, nil
		},
	}

	client, hc, err := f.GraphQLClient(cmdutil.ConsoleEndpoint)
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.Equal(t, "org_1", hc.Organization)
}

func streams(out, errOut *bytes.Buffer) *iostreams.IOStreams {
	return &iostreams.IOStreams{Out: out, ErrOut: errOut}
}
