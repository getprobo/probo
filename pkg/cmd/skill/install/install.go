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

package install

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/cmd/skill/content"
)

func NewCmdInstall(f *cmdutil.Factory) *cobra.Command {
	var (
		flagDir     string
		flagTargets []string
		flagForce   bool
	)

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Copy the built-in agent skills into agent skill directories",
		Example: `  # Install for every supported agent under the home directory
  prb skills install

  # Install only for Cursor
  prb skills install --target cursor

  # Install into one directory
  prb skills install --dir .cursor/skills`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagDir != "" && len(flagTargets) > 0 {
				return fmt.Errorf("pass --dir or --target, not both")
			}

			dirs, err := destinationDirs(flagDir, flagTargets)
			if err != nil {
				return err
			}

			for _, dir := range dirs {
				written, err := content.Install(dir, flagForce)
				if err != nil {
					return err
				}

				for _, path := range written {
					_, _ = fmt.Fprintf(f.IOStreams.Out, "Installed %s\n", path)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagDir, "dir", "", "Directory to receive <skill>/SKILL.md trees")
	cmd.Flags().StringArrayVar(
		&flagTargets,
		"target",
		nil,
		"Agent to install for: cursor, claude, codex, opencode (repeatable; default is all)",
	)
	cmd.Flags().BoolVar(&flagForce, "force", false, "Replace skill files that differ from the copy shipped with prb")

	return cmd
}

func destinationDirs(dir string, targets []string) ([]string, error) {
	if dir != "" {
		return []string{dir}, nil
	}

	if len(targets) == 0 {
		targets = []string{"cursor", "claude", "codex", "opencode"}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot locate home directory: %w", err)
	}

	seen := map[string]struct{}{}

	dirs := make([]string, 0, len(targets))
	for _, target := range targets {
		rel, err := targetPath(target)
		if err != nil {
			return nil, err
		}

		dir := filepath.Join(home, rel)
		if _, ok := seen[dir]; ok {
			continue
		}

		seen[dir] = struct{}{}
		dirs = append(dirs, dir)
	}

	return dirs, nil
}

func targetPath(target string) (string, error) {
	switch target {
	case "cursor":
		return filepath.Join(".cursor", "skills"), nil
	case "claude":
		return filepath.Join(".claude", "skills"), nil
	case "codex":
		return filepath.Join(".codex", "skills"), nil
	case "opencode":
		return filepath.Join(".config", "opencode", "skills"), nil
	default:
		return "", fmt.Errorf("unknown --target %q; valid targets are cursor, claude, codex, opencode", target)
	}
}
