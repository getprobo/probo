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

package root_test

import (
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.probo.inc/probo/pkg/cli/config"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/cmd/iostreams"
	"go.probo.inc/probo/pkg/cmd/root"
)

var kebabCase = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func TestCommandSurface(t *testing.T) {
	t.Parallel()

	cmd := root.NewCmdRoot(&cmdutil.Factory{
		IOStreams: iostreams.System(),
		Config: func() (*config.Config, error) {
			return &config.Config{Hosts: map[string]*config.HostConfig{}}, nil
		},
	})

	var walk func(*cobra.Command)
	walk = func(c *cobra.Command) {
		t.Run(c.CommandPath(), func(t *testing.T) {
			t.Parallel()

			name := c.Name()
			if name != "" && !kebabCase.MatchString(name) {
				t.Errorf("command %q is not kebab-case", c.CommandPath())
			}

			if name == "list" && !slices.Contains(c.Aliases, "ls") {
				t.Errorf("command %q is missing the ls alias", c.CommandPath())
			}

			if strings.Contains(c.Example, "--json") && c.Flags().Lookup("json") == nil {
				t.Errorf("command %q documents --json but does not register it", c.CommandPath())
			}

			c.LocalFlags().VisitAll(func(flag *pflag.Flag) {
				if flag.Hidden || flag.Name == "help" {
					return
				}

				if !kebabCase.MatchString(flag.Name) {
					t.Errorf("command %q flag --%s is not kebab-case", c.CommandPath(), flag.Name)
				}
			})
		})

		for _, child := range c.Commands() {
			walk(child)
		}
	}

	walk(cmd)
}
