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

//go:build darwin || windows

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/deviceagent"
	"go.probo.inc/probo/pkg/deviceagent/tray"
)

func registerPlatformCommands(root *cobra.Command) {
	root.AddCommand(newTrayCmd())
}

func registerTrayAutoStart(exePath string, runDir string) error {
	if err := tray.RegisterAutoStart(exePath, runDir); err != nil {
		return fmt.Errorf("cannot register tray auto-start: %w", err)
	}

	return nil
}

func newTrayCmd() *cobra.Command {
	var runDir string

	cmd := &cobra.Command{
		Use:   "tray",
		Short: "Run the menu bar / system tray enrollment helper",
		RunE: func(cmd *cobra.Command, args []string) error {
			if runDir == "" {
				runDir = deviceagent.DefaultEnrollmentRunDir()
			}

			exePath, err := os.Executable()
			if err != nil {
				return fmt.Errorf("cannot resolve current executable path: %w", err)
			}

			return tray.Run(
				tray.Options{
					RunDir:    runDir,
					ExePath:   exePath,
					ServerURL: deviceagent.DefaultServerURL,
					Version:   version,
				},
			)
		},
	}

	cmd.Flags().StringVar(
		&runDir,
		"run-dir",
		deviceagent.DefaultEnrollmentRunDir(),
		"directory containing the public enrollment marker",
	)

	return cmd
}
