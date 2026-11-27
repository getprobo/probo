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

package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/deviceagent"
	"go.probo.inc/probo/pkg/deviceagent/service"
)

func newDiagnosticsCmd() *cobra.Command {
	var runDir string

	cmd := &cobra.Command{
		Use:   "diagnostics",
		Short: "Write a support bundle for this device",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := resolveDir(cmd)
			if runDir == "" {
				runDir = deviceagent.EnrollmentRunDir(dir)
			}

			path, err := deviceagent.WriteSupportBundle(
				cmd.Context(),
				deviceagent.SupportBundleRequest{
					Dir:          dir,
					RunDir:       runDir,
					Version:      version,
					DescribeUnit: service.DescribeInstalledUnit,
					ReadLog:      service.ReadServiceLogTail,
				},
			)
			if err != nil {
				return err
			}

			if _, err := fmt.Fprintln(cmd.OutOrStdout(), path); err != nil {
				return fmt.Errorf("cannot print support bundle path: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(
		&runDir,
		"run-dir",
		"",
		"directory containing the enrollment marker (default: the run directory for --dir)",
	)

	return cmd
}
