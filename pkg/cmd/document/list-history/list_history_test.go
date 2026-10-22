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

package listhistory_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/cli/config"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	listhistory "go.probo.inc/probo/pkg/cmd/document/list-history"
	"go.probo.inc/probo/pkg/cmd/iostreams"
)

func TestNewCmdListHistory_RejectsNonPositiveLimit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		limit string
	}{
		{name: "zero", limit: "0"},
		{name: "negative", limit: "-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			streams, _, _ := iostreams.Test()
			factory := &cmdutil.Factory{
				IOStreams: streams,
				Config: func() (*config.Config, error) {
					return nil, fmt.Errorf("config should not load")
				},
			}

			cmd := listhistory.NewCmdListHistory(factory)
			cmd.SetOut(streams.Out)
			cmd.SetErr(streams.ErrOut)
			cmd.SetArgs([]string{"document-id", "--limit", tt.limit})

			err := cmd.Execute()

			require.Error(t, err)
			assert.ErrorContains(t, err, "invalid --limit")
		})
	}
}
